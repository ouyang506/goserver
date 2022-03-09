package rpc

import (
	"common/log"
	"sync/atomic"
)

var (
	nextSessionId = int64(0)
)

func genNextSessionId() int64 {
	return atomic.AddInt64(&nextSessionId, 1)
}

type RpcManager struct {
	rpcStubMgr    *RpcStubManger
	pendingRpcMap map[int64]*Rpc //TODO: 添加rpc过期删除
}

func NewRpcManager() *RpcManager {
	mgr := &RpcManager{}
	mgr.rpcStubMgr = NewRpcStubManager()
	mgr.pendingRpcMap = map[int64]*Rpc{}
	return mgr
}

func (mgr *RpcManager) GetStubMgr() *RpcStubManger {
	return mgr.rpcStubMgr
}

func (mgr *RpcManager) AddRpc(rpc *Rpc) bool {
	if rpc.SessionID == 0 {
		rpc.SessionID = genNextSessionId()
	}
	rpcStub := mgr.rpcStubMgr.FindStub(rpc)
	if rpcStub == nil {
		return false
	}
	ret := rpcStub.PushRpc(rpc)
	if !ret {
		log.Error("push rpc error")
		return false
	}

	mgr.pendingRpcMap[rpc.SessionID] = rpc
	return true
}

func (mgr *RpcManager) RcvRpcResponse(sessionId int64, resp []byte) {
	rpc, ok := mgr.pendingRpcMap[sessionId]
	if !ok {
		return
	}
	select {
	case rpc.Chan <- nil:
		{

		}
	default:
	}
}
