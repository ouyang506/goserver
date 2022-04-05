package handler

import (
	"common/pbmsg"
)

func HandleLoginGateReq(req *pbmsg.LoginGateReqT, resp *pbmsg.LoginGateRespT) {
	resp.Result = new(int32)
	*resp.Result = 100
	return
}
