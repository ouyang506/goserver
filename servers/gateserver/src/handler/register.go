package handler

import (
	"common/pbmsg"
	"common/rpc"
)

func RegHandler() {
	rpc.RegMsgHandler(pbmsg.MsgID_login_gate_req, HandleLoginGateReq)
}
