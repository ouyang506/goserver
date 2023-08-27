package handler

import (
	"common/redisutil"
	"encoding/json"
	"fmt"
	"framework/log"
	"framework/proto/pb"
	"framework/proto/pb/cs"
	"gate/configmgr"
	"strconv"
)

func (h *MessageHandler) HandleRpcReqLoginGate(req *cs.ReqLoginGate, resp *cs.RespLoginGate) {
	reqJson, _ := json.Marshal(req)
	log.Debug("rcv ReqLoginGate: %s", string(reqJson))
	defer func() {
		respJson, _ := json.Marshal(resp)
		log.Debug("response RespLoginGate: %s", string(respJson))
	}()

	roleId := req.GetRoleId()
	token := req.GetToken()
	rkey := fmt.Sprintf(redisutil.RKeyLoginGateToken, token)
	rvalue, err := redisutil.Get(rkey)
	if err != nil {
		log.Error("query login token from redis error: %v", err)
		resp.ErrCode = new(int32)
		*resp.ErrCode = int32(pb.ERROR_CODE_DB_FAILED)
		resp.ErrDesc = new(string)
		*resp.ErrDesc = "query db failed"
		return
	}

	if rvalue == "" {
		log.Error("token not found, roleId= %v", roleId)
		resp.ErrCode = new(int32)
		*resp.ErrCode = int32(pb.ERROR_CODE_CHECK_LOGIN_TOKEN_FAILED)
		resp.ErrDesc = new(string)
		*resp.ErrDesc = "check token failed"
		return
	}

	rvalueMap := make(map[string]string)
	err = json.Unmarshal([]byte(rvalue), &rvalueMap)
	if err != nil {
		log.Error("unmarshal login gate value error: %v", err)
		resp.ErrCode = new(int32)
		*resp.ErrCode = int32(pb.ERROR_CODE_UNKOWN)
		resp.ErrDesc = new(string)
		*resp.ErrDesc = "parse token data failed"
		return
	}

	strRoleId, _ := rvalueMap["role_id"]
	if strRoleId != strconv.FormatInt(roleId, 10) {
		log.Error("token not match, value= %v", rvalue)
		resp.ErrCode = new(int32)
		*resp.ErrCode = int32(pb.ERROR_CODE_CHECK_LOGIN_TOKEN_FAILED)
		resp.ErrDesc = new(string)
		*resp.ErrDesc = "check token failed"
		return
	}

	//need check gate addr?
	conf := configmgr.Instance().GetConfig()
	outerIp := conf.Outer.OuterIp
	outerPort := conf.Outer.Port
	strGateIp, _ := rvalueMap["gate_ip"]
	strGatePort, _ := rvalueMap["gate_port"]
	if strGateIp != outerIp || strGatePort != strconv.Itoa(outerPort) {
		log.Error("gate endpoint not match, value= %v", rvalue)
		resp.ErrCode = new(int32)
		*resp.ErrCode = int32(pb.ERROR_CODE_CHECK_LOGIN_GATE_ENDPOINT_FAILED)
		resp.ErrDesc = new(string)
		*resp.ErrDesc = "check gate endpoint failed"
		return
	}
}
