package robot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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

func httpLogin(url string, username string, password string) (error, *AccountLoginResp) {
	loginReq := AccountLoginReq{}
	loginReq.UserName = username
	loginReq.Password = password
	data, _ := json.Marshal(loginReq)
	r, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err, nil
	}

	r.Header.Set("Content-Type", "application/json")

	c := http.Client{}
	resp, err := c.Do(r)
	if err != nil {
		return err, nil
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return err, nil
	}
	loginResp := &AccountLoginResp{}
	err = json.Unmarshal(respData, loginResp)
	if err != nil {
		return err, nil
	}

	if loginResp.ErrCode != 0 {
		return fmt.Errorf("login server response error, error code: %d, error desc: %s",
			loginResp.ErrCode, loginResp.ErrDesc), nil
	}

	if loginResp.GateIp == "" || loginResp.GatePort == 0 {
		return errors.New("login server response gate addr error"), nil
	}

	return nil, loginResp
}
