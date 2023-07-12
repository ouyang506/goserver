package handler

import (
	"framework/log"
	"framework/proto/pb/cs"
)

func (h *MessageHandler) HandleRpcReqLoginGate(req *cs.ReqLoginGate, resp *cs.RespLoginGate) {
	log.Debug("rcv ReqLoginGate : %s", req.String())
	resp.Result = new(int32)
	*resp.Result = 100
	log.Debug("response RespLoginGate : %s", resp.String())
}
