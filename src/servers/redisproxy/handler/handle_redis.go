package handler

import (
	"encoding/json"
	"framework/log"
	"framework/proto/pb/ssredisproxy"
	"framework/rpc"
	"redisproxy/redismgr"
)

func (h *MessageHandler) HandleRpcReqRedisCmd(ctx rpc.Context,
	req *ssredisproxy.ReqRedisCmd, resp *ssredisproxy.RespRedisCmd) {
	reqJson, _ := json.Marshal(req)
	log.Debug("rcv ReqRedisCmd : %s", string(reqJson))
	defer func() {
		respJson, _ := json.Marshal(resp)
		log.Debug("response RespRedisCmd : %s", string(respJson))
	}()

	args := make([]any, len(req.GetArgs()))
	for i, v := range req.GetArgs() {
		args[i] = v
	}
	result, err := redismgr.Instance().DoCmd(args...)
	if err != nil {
		if redismgr.Instance().IsNilError(err) {
			resp.ErrCode = new(int32)
			*resp.ErrCode = int32(ssredisproxy.Errors_redis_nil_error)
			resp.ErrDesc = new(string)
			*resp.ErrDesc = err.Error()
			return
		}
		log.Error("redis do cmd error: %v", err)
		resp.ErrCode = new(int32)
		*resp.ErrCode = int32(ssredisproxy.Errors_execute_failed)
		resp.ErrDesc = new(string)
		*resp.ErrDesc = err.Error()
		return
	}
	data, _ := json.Marshal(result)
	resp.Result = new(string)
	*resp.Result = string(data)
}

// TODO implement
func (h *MessageHandler) HandleRpcReqRedisEval(ctx rpc.Context,
	req *ssredisproxy.ReqRedisEval, resp *ssredisproxy.RespRedisEval) {
}
