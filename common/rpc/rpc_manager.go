package rpc

import (
	"common/log"
	"common/utility/timer"
	"errors"
	"sync"
	"time"
)

var (
	ErrorAddRpc     = errors.New("add rpc error")
	ErrorRpcTimeOut = errors.New("rpc time out error")
)

// RPC管理器(thread safe)
type RpcManager struct {
	mutex      sync.Mutex
	rpcStubMgr *RpcStubManger

	pendingRpcMap map[int64]*PendingRpcEntry
	timerMgr      *timer.TimerWheel
}

type PendingRpcEntry struct {
	rpc     *Rpc
	rpcStub *RpcStub
}

func NewRpcManager() *RpcManager {
	mgr := &RpcManager{}
	mgr.rpcStubMgr = newRpcStubManager(mgr)
	mgr.pendingRpcMap = map[int64]*PendingRpcEntry{}
	mgr.timerMgr = timer.NewTimerWheel(&timer.Option{TimeAccuracy: time.Millisecond * 200})
	mgr.timerMgr.Start()
	return mgr
}

// 添加一个代理管道
func (mgr *RpcManager) AddStub(serverType int, instanceId int, remoteIp string, remotePort int) bool {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	return mgr.rpcStubMgr.addStub(serverType, instanceId, remoteIp, remotePort)
}

// 删除一个代理管道
func (mgr *RpcManager) DelStub(serverType int, instanceId int) bool {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	return mgr.rpcStubMgr.delStub(serverType, instanceId)
}

func (mgr *RpcManager) AddRpc(rpc *Rpc) bool {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	if rpc.CallId == 0 {
		rpc.CallId = genNextRpcCallId()
	}

	rpcStub := mgr.rpcStubMgr.selectStub(rpc)
	if rpcStub == nil {
		return false
	}
	ret := rpcStub.pushRpc(rpc)
	if !ret {
		log.Error("push rpc error, callId : %v", rpc.CallId)
		return false
	}

	pendingRpcEntry := &PendingRpcEntry{
		rpc:     rpc,
		rpcStub: rpcStub,
	}
	mgr.pendingRpcMap[rpc.CallId] = pendingRpcEntry

	mgr.timerMgr.AddTimer(rpc.Timeout, func() {
		log.Debug("rpc timeout : CallId : %v", rpc.CallId)
		mgr.RemoveRpc(rpc.CallId)
		select {
		case rpc.RespChan <- ErrorRpcTimeOut:
		default:
		}
	})

	return true
}

func (mgr *RpcManager) RemoveRpc(sessionId int64) bool {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	pendingRpcEntry, ok := mgr.pendingRpcMap[sessionId]
	if !ok {
		return false
	}

	if pendingRpcEntry.rpcStub == nil {
		return false
	}

	if !pendingRpcEntry.rpcStub.removeRpc(sessionId) {
		return false
	}

	return true
}

// 收到rpc返回后回调
func (mgr *RpcManager) OnRcvResponse(sessionId int64, respMsg interface{}) {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	pendingRpcEntry, ok := mgr.pendingRpcMap[sessionId]
	if !ok {
		return
	}

	pendingRpcEntry.rpc.RespMsg = respMsg
	select {
	case pendingRpcEntry.rpc.RespChan <- nil:
	default:
	}

	pendingRpcEntry.rpcStub.removeRpc(sessionId)
}
