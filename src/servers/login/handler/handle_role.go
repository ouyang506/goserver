package handler

import (
	"common/mysqlutil"
	"encoding/json"
	"framework/log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CreateRoleReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}

type CreateRoleResp struct {
	ErrCode  int    `json:"error_code"`
	ErrDesc  string `json:"error_desc"`
	GateIp   string `json:"gate_ip"`
	GatePort int    `json:"gate_port"`
	Token    string `json:"token"`
	RoleId   int64  `json:"role_id"`
}

// 创建角色
func handleCreateRole(c *gin.Context) {
	req := &CreateRoleReq{}
	resp := &CreateRoleResp{}

	// unmarshal request
	err := c.BindJSON(req)
	if err != nil {
		resp.ErrCode = ErrCodeUnmarshalParam
		resp.ErrDesc = ErrDescUnmarshalParam
		c.JSON(http.StatusOK, resp)
		return
	}

	reqJson, _ := json.Marshal(req)
	log.Debug("rcv create role : %s", string(reqJson))
	defer func() {
		respJson, _ := json.Marshal(resp)
		log.Debug("response create role : %s", string(respJson))
	}()

	if req.Nickname == "" || len(req.Nickname) > 64 {
		resp.ErrCode = ErrCodeInvalidNickname
		resp.ErrDesc = ErrDescInvalidNickname
		c.JSON(http.StatusOK, resp)
		return
	}

	// check username and passwd
	row, err := mysqlutil.QueryOne("select passwd from account where username=?", req.Username)
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

	row, err = mysqlutil.QueryOne("select id from role where nickname=?", req.Nickname)
	if err != nil {
		log.Error("query role error: %v", err)
		resp.ErrCode = ErrCodeDBFailed
		resp.ErrDesc = ErrDescDBFailed
		c.JSON(http.StatusOK, resp)
		return
	}

	if row != nil {
		resp.ErrCode = ErrCodeNicknameExist
		resp.ErrDesc = ErrDescNicknameExist
		c.JSON(http.StatusOK, resp)
		return
	}

	lastInsertId, _, err := mysqlutil.Execute("insert into role(account,nickname) values(?,?)",
		req.Username, req.Nickname)
	if err != nil {
		log.Error("insert role error: %v", err)
		resp.ErrCode = ErrCodeDBFailed
		resp.ErrDesc = ErrDescDBFailed
		c.JSON(http.StatusOK, resp)
		return
	}

	roleId := lastInsertId

	token := ""
	gateIp := ""
	gatePort := 0
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
	token, err = genGateToken(roleId, gateIp, gatePort)
	if err != nil {
		log.Error("generate gate token failed, %v", err)
		resp.ErrCode = ErrCodeDBFailed
		resp.ErrDesc = ErrDescDBFailed
		c.JSON(http.StatusOK, resp)
		return
	}

	resp.GateIp = gateIp
	resp.GatePort = gatePort
	resp.Token = token
	resp.RoleId = roleId
	c.JSON(http.StatusOK, resp)
}
