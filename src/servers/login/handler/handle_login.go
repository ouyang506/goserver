package handler

import (
	"common/mysqlutil"
	"common/redisutil"
	"encoding/json"
	"fmt"
	"framework/log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type AccountLoginReq struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

type AccountLoginResp struct {
	ErrCode  int    `json:"error_code"`
	ErrDesc  string `json:"error_desc"`
	GateIp   string `json:"gate_ip"`
	GatePort int    `json:"gate_port"`
	Token    string `json:"token"`
	RoleId   int64  `json:"role_id"`
}

// 用户名+密码登录
func handlerAccountLogin(c *gin.Context) {
	req := &AccountLoginReq{}
	resp := &AccountLoginResp{}

	// unmarshal request
	err := c.BindJSON(req)
	if err != nil {
		resp.ErrCode = ErrCodeUnmarshalParam
		resp.ErrDesc = ErrDescUnmarshalParam
		c.JSON(http.StatusOK, resp)
		return
	}

	// check username and passwd
	row, err := mysqlutil.QueryOne("select passwd from account where username=?", req.UserName)
	if err != nil {
		log.Error("query account error: %v", err)
		resp.ErrCode = ErrCodeDBFailed
		resp.ErrDesc = ErrDescDBFailed
		c.JSON(http.StatusOK, resp)
		return
	}

	if row == nil || row.FieldString("passwd") != req.Password {
		resp.ErrCode = ErrCodeInvalidUserOrPasswd
		resp.ErrDesc = ErrDescInvalidUserOrPasswd
		c.JSON(http.StatusOK, resp)
		return
	}

	// query role id if has created the role
	row, err = mysqlutil.QueryOne("select id, nickname from role where account = ?", req.UserName)
	if err != nil {
		log.Error("query role error: %v", err)
		resp.ErrCode = ErrCodeDBFailed
		resp.ErrDesc = ErrDescDBFailed
		c.JSON(http.StatusOK, resp)
		return
	}

	roleId := int64(0)
	if row != nil {
		roleId = row.FieldInt64("id")
	}

	addrInfos, err := redisutil.HGetAll(redisutil.RKeyGateOuterAddr)
	if err != nil {
		log.Error("query gate addr failed: %v", err)
		resp.ErrCode = ErrCodeDBFailed
		resp.ErrDesc = ErrDescDBFailed
		c.JSON(http.StatusOK, resp)
		return
	}

	now := time.Now().Unix()
	validAddrs := [][2]string{}
	for k, v := range addrInfos {
		expired, _ := strconv.ParseInt(v, 10, 64)
		if expired == 0 || now >= expired {
			continue
		}
		splits := strings.SplitN(k, "_", 2)
		if len(splits) != 2 {
			continue
		}
		validAddrs = append(validAddrs, [2]string{splits[0], splits[1]})
	}

	if len(validAddrs) == 0 {
		log.Error("available gate not found")
		resp.ErrCode = ErrCodeValidGateNotFound
		resp.ErrDesc = ErrDescValidGateNotFound
		c.JSON(http.StatusOK, resp)
		return
	}

	randIdx := rand.Int() % len(validAddrs)
	gateIp := validAddrs[randIdx][0]
	gatePort, _ := strconv.Atoi(validAddrs[randIdx][1])

	token := ""

	// 存在创角，生成登录gate的token
	if roleId > 0 {
		//TODO: random a string for token
		token = strconv.FormatInt(rand.Int63(), 32)

		rkey := fmt.Sprintf(redisutil.RKeyLoginGateToken, token)
		rvalueMap := map[string]string{
			"role_id":   strconv.FormatInt(roleId, 10),
			"gate_id":   gateIp,
			"gate_port": strconv.Itoa(gatePort),
		}

		rvalue, _ := json.Marshal(rvalueMap)
		err = redisutil.SetEx(rkey, string(rvalue), 7*24*3600)
		if err != nil {
			log.Error("save login gate token to redis failed, %v", err)
			resp.ErrCode = ErrCodeDBFailed
			resp.ErrDesc = ErrDescDBFailed
			c.JSON(http.StatusOK, resp)
			return
		}
	}

	resp.ErrCode = ErrCodeSuccess
	resp.ErrDesc = ErrDescSuccess
	resp.GateIp = gateIp
	resp.GatePort = gatePort
	resp.Token = token
	resp.RoleId = roleId
	c.JSON(http.StatusOK, resp)
}
