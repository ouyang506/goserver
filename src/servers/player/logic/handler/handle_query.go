package handler

import (
	"encoding/json"
	"framework/log"
	"framework/proto/pb/cs"
	"framework/rpc"
)

func (h *MessageHandler) HandleRpcReqQueryPlayer(ctx rpc.Context, req *cs.ReqQueryPlayer, resp *cs.RespQueryPlayer) {
	defer func() {
		respJson, _ := json.Marshal(resp)
		reqJson, _ := json.Marshal(req)
		log.Debug("rcv ReqQueryPlayer: %s, response RespQueryPlayer: %s",
			string(reqJson), string(respJson))
	}()

	resp.ErrCode = new(int32)
	*resp.ErrCode = 0
	resp.ErrDesc = new(string)
	*resp.ErrDesc = "OK"
}
