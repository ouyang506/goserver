package rpc

import (
	"common/log"
	"common/utility/timer"
	"time"
)

type PendingRpcEntry struct {
	rpc     *Rpc
	rpcStub *RpcStub
}

type RpcManager struct {
	rpcStubMgr *RpcStubManger

	pendingRpcMap map[int64]*PendingRpcEntry
	timerMgr      *timer.TimerWheel
}

func NewRpcManager() *RpcManager {
	mgr := &RpcManager{}
	mgr.rpcStubMgr = NewRpcStubManager()
	mgr.pendingRpcMap = map[int64]*PendingRpcEntry{}
	mgr.timerMgr = timer.NewTimerWheel(&timer.Option{TimeAccuracy: time.Millisecond * 200})
	mgr.timerMgr.Start()
	return mgr
}

func (mgr *RpcManager) GetStubMgr() *RpcStubManger {
	return mgr.rpcStubMgr
}

func (mgr *RpcManager) AddRpc(rpc *Rpc) bool {
	if rpc.CallId == 0 {
		rpc.CallId = genNextRpcCallId()
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

	pendingRpcEntry := &PendingRpcEntry{
		rpc:     rpc,
		rpcStub: rpcStub,
	}
	mgr.pendingRpcMap[rpc.CallId] = pendingRpcEntry

	mgr.timerMgr.AddTimer(rpc.Timeout, func() {
		log.Debug("rcp timeout : CallId : %v", rpc.CallId)
		mgr.RemoveRpc(rpc.CallId)
		select {
		case rpc.Chan <- nil:
		default:
		}
	})

	return true
}

func (mgr *RpcManager) RemoveRpc(sessionId int64) bool {
	pendingRpcEntry, ok := mgr.pendingRpcMap[sessionId]
	if !ok {
		return false
	}

	if pendingRpcEntry.rpcStub == nil {
		return false
	}

	if !pendingRpcEntry.rpcStub.RemoveRpc(sessionId) {
		return false
	}

	return true
}

func (mgr *RpcManager) RcvRpcResponse(sessionId int64, resp []byte) {
	pendingRpcEntry, ok := mgr.pendingRpcMap[sessionId]
	if !ok {
		return
	}
	select {
	case pendingRpcEntry.rpc.Chan <- nil:
		{

		}
	default:
	}
}
