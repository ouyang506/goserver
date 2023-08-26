package robotmgr

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type CreateAccountReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CreateAccountResp struct {
	ErrCode int    `json:"error_code"`
	ErrDesc string `json:"error_desc"`
}

func createAccount(url string, username string, password string) (*CreateAccountResp, error) {
	createReq := CreateAccountReq{}
	createReq.Username = username
	createReq.Password = password
	data, _ := json.Marshal(createReq)
	r, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	r.Header.Set("Content-Type", "application/json")
	c := http.Client{}
	resp, err := c.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	createResp := &CreateAccountResp{}
	err = json.Unmarshal(respData, createResp)
	if err != nil {
		return nil, err
	}

	return createResp, nil
}

type LoginAccountReq struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

type LoginAccountResp struct {
	ErrCode  int    `json:"error_code"`
	ErrDesc  string `json:"error_desc"`
	GateIp   string `json:"gate_ip"`
	GatePort int    `json:"gate_port"`
	Token    string `json:"token"`
	RoleId   int64  `json:"role_id"`
}

func loginAccount(url string, username string, password string) (*LoginAccountResp, error) {
	loginReq := LoginAccountReq{}
	loginReq.UserName = username
	loginReq.Password = password
	data, _ := json.Marshal(loginReq)
	r, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	r.Header.Set("Content-Type", "application/json")
	c := http.Client{}
	resp, err := c.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	loginResp := &LoginAccountResp{}
	err = json.Unmarshal(respData, loginResp)
	if err != nil {
		return nil, err
	}

	return loginResp, nil
}

type CreateRoleReq struct {
	UserName string `json:"username"`
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

func createRole(url string, username string, password string, nickname string) (*CreateRoleResp, error) {
	createReq := CreateRoleReq{}
	createReq.UserName = username
	createReq.Password = password
	createReq.Nickname = nickname
	data, _ := json.Marshal(createReq)
	r, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	r.Header.Set("Content-Type", "application/json")
	c := http.Client{}
	resp, err := c.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	createResp := &CreateRoleResp{}
	err = json.Unmarshal(respData, createResp)
	if err != nil {
		return nil, err
	}

	return createResp, nil
}
