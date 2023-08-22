package handler

import (
	"encoding/json"
	"framework/log"
	"framework/proto/pb/ss"
	"redisproxy/redismgr"
)

func (h *MessageHandler) HandleRpcReqRedisCmd(req *ss.ReqRedisCmd, resp *ss.RespRedisCmd) {
	reqJson, _ := json.Marshal(req)
	log.Debug("rcv ReqRedisCmd : %s", string(reqJson))
	defer func() {
		respJson, _ := json.Marshal(resp)
		log.Debug("response RespRedisCmd : %s", string(respJson))
	}()

	redisMgr := redismgr.GetRedisMgr()
	args := make([]any, len(req.GetArgs()))
	for i, v := range req.GetArgs() {
		args[i] = v
	}
	result, err := redisMgr.DoCmd(args...)
	if err != nil {
		if redisMgr.IsNilError(err) {
			resp.ErrCode = new(int32)
			*resp.ErrCode = int32(ss.SsRedisProxyError_redis_nil_error)
			resp.ErrDesc = new(string)
			*resp.ErrDesc = err.Error()
			return
		}
		log.Error("redis do cmd error: %v", err)
		resp.ErrCode = new(int32)
		*resp.ErrCode = int32(ss.SsRedisProxyError_execute_failed)
		resp.ErrDesc = new(string)
		*resp.ErrDesc = err.Error()
		return
	}
	data, _ := json.Marshal(result)
	resp.Result = new(string)
	*resp.Result = string(data)
}

// TODO implement
func (h *MessageHandler) HandleRpcReqRedisEval(req *ss.ReqRedisEval, resp *ss.RespRedisEval) {
}
