package handler

import (
	"common/redisutil"
	"encoding/json"
	"fmt"
	"framework/log"
	"framework/proto/pb"
	"framework/proto/pb/cs"
	"framework/rpc"
	"gate/configmgr"
	"gate/netmgr"
	"gate/playermgr"
	"strconv"
)

func (h *MessageHandler) HandleRpcReqLoginGate(ctx rpc.Context, req *cs.ReqLoginGate, resp *cs.RespLoginGate) {
	reqJson, _ := json.Marshal(req)
	log.Debug("rcv ReqLoginGate: %s", string(reqJson))
	defer func() {
		respJson, _ := json.Marshal(resp)
		log.Debug("response RespLoginGate: %s", string(respJson))
	}()

	playerId := req.GetPlayerId()
	strPlayerId := strconv.FormatInt(playerId, 10)
	token := req.GetToken()
	rkey := fmt.Sprintf(redisutil.RKeyLoginGateToken, strPlayerId)
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
		log.Error("token not found, playerId= %v", playerId)
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

	savedToken, ok := rvalueMap["token"]
	if !ok || token != savedToken {
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
	strGateIp := rvalueMap["gate_ip"]
	strGatePort := rvalueMap["gate_port"]
	if strGateIp != outerIp || strGatePort != strconv.Itoa(outerPort) {
		log.Error("gate endpoint not match, value= %v", rvalue)
		resp.ErrCode = new(int32)
		*resp.ErrCode = int32(pb.ERROR_CODE_CHECK_LOGIN_GATE_ENDPOINT_FAILED)
		resp.ErrDesc = new(string)
		*resp.ErrDesc = "check gate endpoint failed"
		return
	}

	conn := ctx.GetNetConn()
	if conn == nil {
		log.Error("get connection nil, playerId=%v", playerId)
		resp.ErrCode = new(int32)
		*resp.ErrCode = int32(pb.ERROR_CODE_UNKOWN)
		resp.ErrDesc = new(string)
		*resp.ErrDesc = "unkown error"
	}
	conn.SetAttrib(netmgr.NetAttrPlayer{}, playerId)
	log.Info("player login gate, playerId = %v, sessionId = %v", playerId, conn.GetSessionId())

	player := playermgr.NewPlayer(playerId)
	player.SetNetSessionId(conn.GetSessionId())
	oldPlayer := playermgr.Instance().LoadAndStorePlayer(playerId, player)
	if oldPlayer != nil {
		log.Info("remove old player and close the connection, playerId=%d, oldNetSessionId=%v",
			playerId, oldPlayer.GetNetSessionId())
		conn.DelAttrib(netmgr.NetAttrPlayer{})
		rpc.TcpClose(rpc.RpcModeOuter, conn.GetSessionId())
	}

	playermgr.Instance().OnLogin(player)
}
