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

type CreateAccountReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CreateAccountResp struct {
	ErrCode int    `json:"error_code"`
	ErrDesc string `json:"error_desc"`
}

type LoginAccountReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginAccountResp struct {
	ErrCode  int    `json:"error_code"`
	ErrDesc  string `json:"error_desc"`
	GateIp   string `json:"gate_ip"`
	GatePort int    `json:"gate_port"`
	Token    string `json:"token"`
	PlayerId int64  `json:"player_id"`
}

// 创建账号
func handlerCreateAccount(c *gin.Context) {
	req := &CreateAccountReq{}
	resp := &CreateAccountResp{}

	// unmarshal request
	err := c.BindJSON(req)
	if err != nil {
		resp.ErrCode = ErrCodeUnmarshalParam
		resp.ErrDesc = ErrDescUnmarshalParam
		c.JSON(http.StatusOK, resp)
		return
	}

	reqJson, _ := json.Marshal(req)
	log.Debug("rcv create account: %s", string(reqJson))
	defer func() {
		respJson, _ := json.Marshal(resp)
		log.Debug("response create account: %s", string(respJson))
	}()

	if req.Username == "" || len(req.Username) > 64 {
		resp.ErrCode = ErrCodeInvalidUsername
		resp.ErrDesc = ErrDescInvalidUsername
		c.JSON(http.StatusOK, resp)
		return
	}
	row, err := mysqlutil.QueryOne("select id from t_account where username=?", req.Username)
	if err != nil {
		log.Error("query account error: %v", err)
		resp.ErrCode = ErrCodeDBFailed
		resp.ErrDesc = ErrDescDBFailed
		c.JSON(http.StatusOK, resp)
		return
	}

	if row != nil {
		resp.ErrCode = ErrCodeUsernameExist
		resp.ErrDesc = ErrDescUsernameExist
		c.JSON(http.StatusOK, resp)
		return
	}

	_, _, err = mysqlutil.Execute("insert into t_account(username, passwd) values(?,?)",
		req.Username, req.Password)
	if err != nil {
		log.Error("insert account error: %v", err)
		resp.ErrCode = ErrCodeDBFailed
		resp.ErrDesc = ErrDescDBFailed
		c.JSON(http.StatusOK, resp)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// 用户名+密码登录
func handleLoginAccount(c *gin.Context) {
	req := &LoginAccountReq{}
	resp := &LoginAccountResp{}

	// unmarshal request
	err := c.BindJSON(req)
	if err != nil {
		resp.ErrCode = ErrCodeUnmarshalParam
		resp.ErrDesc = ErrDescUnmarshalParam
		c.JSON(http.StatusOK, resp)
		return
	}

	reqJson, _ := json.Marshal(req)
	log.Debug("rcv login account: %s", string(reqJson))
	defer func() {
		respJson, _ := json.Marshal(resp)
		log.Debug("response login account: %s", string(respJson))
	}()

	// check username and passwd
	row, err := mysqlutil.QueryOne("select passwd from t_account where username=?", req.Username)
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

	// query player id if has created the player
	row, err = mysqlutil.QueryOne("select id, nickname from t_player where account = ?", req.Username)
	if err != nil {
		log.Error("query player error: %v", err)
		resp.ErrCode = ErrCodeDBFailed
		resp.ErrDesc = ErrDescDBFailed
		c.JSON(http.StatusOK, resp)
		return
	}

	playerId := int64(0)
	if row != nil {
		playerId = row.FieldInt64("id")
	}

	token := ""
	gateIp := ""
	gatePort := 0

	if playerId > 0 {
		// 存在创角，选择一个gate地址供客户端连接
		gateIp, gatePort, err = fetchOneGate()
		if err != nil {
			log.Error("query gate addr failed: %v", err)
			resp.ErrCode = ErrCodeDBFailed
			resp.ErrDesc = ErrDescDBFailed
			c.JSON(http.StatusOK, resp)
			return
		}

		if gateIp == "" || gatePort == 0 {
			log.Error("available gate not found")
			resp.ErrCode = ErrCodeValidGateNotFound
			resp.ErrDesc = ErrDescValidGateNotFound
			c.JSON(http.StatusOK, resp)
			return
		}

		// 存在创角，生成登录gate的token
		token, err = genGateToken(playerId, gateIp, gatePort)
		if err != nil {
			log.Error("generate gate token failed failed, %v", err)
			resp.ErrCode = ErrCodeDBFailed
			resp.ErrDesc = ErrDescDBFailed
			c.JSON(http.StatusOK, resp)
			return
		}
	}

	resp.GateIp = gateIp
	resp.GatePort = gatePort
	resp.Token = token
	resp.PlayerId = playerId
	c.JSON(http.StatusOK, resp)
}

func fetchOneGate() (string, int, error) {
	addrInfos, err := redisutil.HGetAll(redisutil.RKeyGateOuterAddr)
	if err != nil {
		return "", 0, err
	}

	now := time.Now().Unix()
	validAddrs := [][2]string{}
	expiredArr := []string{} // 过期待删除的
	for k, v := range addrInfos {
		expired, _ := strconv.ParseInt(v, 10, 64)
		if expired == 0 || now >= expired {
			expiredArr = append(expiredArr, k)
			continue
		}
		splits := strings.SplitN(k, "_", 2)
		if len(splits) != 2 {
			continue
		}
		validAddrs = append(validAddrs, [2]string{splits[0], splits[1]})
	}
	// 触发删除非活跃的gate地址，单次至多5个
	if len(expiredArr) > 0 {
		if len(expiredArr) > 5 {
			expiredArr = expiredArr[:5]
		}
		redisutil.HDel(redisutil.RKeyGateOuterAddr, expiredArr...)
	}

	if len(validAddrs) == 0 {
		return "", 0, nil
	}

	randIdx := rand.Int() % len(validAddrs)
	gateIp := validAddrs[randIdx][0]
	gatePort, _ := strconv.Atoi(validAddrs[randIdx][1])
	return gateIp, gatePort, nil
}

func genGateToken(playerId int64, gateIp string, gatePort int) (string, error) {
	//TODO: random a string for token
	token := strconv.FormatInt(rand.Int63(), 32)
	strPlayerId := strconv.FormatInt(playerId, 10)
	rkey := fmt.Sprintf(redisutil.RKeyLoginGateToken, strPlayerId)
	rvalueMap := map[string]string{
		"token":     token,
		"gate_ip":   gateIp,
		"gate_port": strconv.Itoa(gatePort),
	}

	rvalue, _ := json.Marshal(rvalueMap)
	err := redisutil.SetEx(rkey, string(rvalue), 7*24*3600)
	if err != nil {
		return "", err
	}
	return token, nil
}
