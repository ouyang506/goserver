package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AccountLoginReq struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

type AccountLoginResp struct {
	ErrCode  int    `json:"error_code"`
	ErrDesc  string `json:"error_desc"`
	Token    string `json:"token"`
	GateIp   string `json:"gate_ip"`
	GatePort int    `json:"gate_port"`
}

// 用户名+密码登录
func handlerAccountLogin(c *gin.Context) {
	req := &AccountLoginReq{}
	resp := &AccountLoginResp{}
	err := c.BindJSON(req)
	if err != nil {
		resp.ErrCode = ErrCodeUnmarshalParam
		resp.ErrDesc = ErrDescUnmarshalParam
		c.JSON(http.StatusOK, resp)
		return
	}

	if req.UserName != "admin" || req.Password != "123456" {
		resp.ErrCode = ErrCodeInvalidUserOrPasswd
		resp.ErrDesc = ErrDescInvalidUserOrPasswd
		c.JSON(http.StatusOK, resp)
		return
	}
	resp.ErrCode = ErrCodeSuccess
	resp.ErrDesc = ErrDescSuccess
	resp.GateIp = "127.0.0.1"
	resp.GatePort = 5000
	resp.Token = "token_test"
	c.JSON(http.StatusOK, resp)
}
