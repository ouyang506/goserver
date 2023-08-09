package handler

import (
	"framework/log"
	"framework/proto/pb/ss"
)

func (h *MessageHandler) HandleRpcReqExecuteSql(req *ss.ReqExecuteSql, resp *ss.RespExecuteSql) {
	log.Debug("rcv ReqExecuteSql : %s", req.String())
	log.Debug("response RespExecuteSql : %s", resp.String())
}
