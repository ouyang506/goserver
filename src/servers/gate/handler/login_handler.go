package handler

import (
	"framework/proto/pb/cs"
)

func (h *MessageHandler) HandleReqLoginGate(req *cs.ReqLoginGate, resp *cs.RespLoginGate) {
	resp.Result = new(int32)
	*resp.Result = 100
}
