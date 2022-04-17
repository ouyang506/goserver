package rpc

import (
	"common/log"
	"common/utility/timer"
	"errors"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"
)

var (
	ErrorAddRpc         = errors.New("add rpc error")
	ErrorRpcTimeOut     = errors.New("rpc time out error")
	ErrorRpcRespMsgType = errors.New("rpc response msg type error")
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

// 添加一个代理管道(注册中心调用)
func (mgr *RpcManager) AddStub(serverType int, instanceId int, remoteIp string, remotePort int) bool {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	return mgr.rpcStubMgr.addStub(serverType, instanceId, remoteIp, remotePort)
}

// 删除一个代理管道(注册中心调用)
func (mgr *RpcManager) DelStub(serverType int, instanceId int) bool {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	return mgr.rpcStubMgr.delStub(serverType, instanceId)
}

// 添加一个rpc到管理器
func (mgr *RpcManager) AddRpc(rpc *Rpc) bool {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	var rpcStub *RpcStub = nil
	if rpc.TargetSvrType <= 0 {
		log.Error("target server type error, TargetSvrType : %v", rpc.TargetSvrType)
		return false
	}

	if rpc.TargetSvrInstId > 0 {
		rpcStub = mgr.rpcStubMgr.findStub(rpc.TargetSvrType, rpc.TargetSvrInstId)
	} else {
		rpcStub = mgr.rpcStubMgr.selectStub(rpc)
	}
	if rpcStub == nil {
		log.Error("cannot find rpc stub, serverType : %v, instId : %v", rpc.TargetSvrType, rpc.TargetSvrInstId)
		return false
	}

	ret := rpcStub.pushRpc(rpc)
	if !ret {
		log.Error("push rpc error, callId : %v", rpc.CallId)
		return false
	}

	// not need waiting response
	if rpc.IsOneway {
		return true
	}

	pendingRpcEntry := &PendingRpcEntry{
		rpc:     rpc,
		rpcStub: rpcStub,
	}
	mgr.pendingRpcMap[rpc.CallId] = pendingRpcEntry

	rpc.WaitTimer = mgr.timerMgr.AddTimer(rpc.Timeout, func() {
		log.Debug("rpc timeout : CallId : %v", rpc.CallId)
		mgr.RemoveRpc(rpc.CallId)
		select {
		case rpc.RespChan <- ErrorRpcTimeOut:
		default:
		}
		// 异步调用回调
		if rpc.IsAsync && rpc.Callback != nil {
			rpc.Callback(ErrorRpcTimeOut, nil)
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

	if pendingRpcEntry.rpc.WaitTimer != nil {
		mgr.timerMgr.RemoveTimer(pendingRpcEntry.rpc.WaitTimer)
	}

	// 异步rpc回调
	if pendingRpcEntry.rpc.IsAsync && pendingRpcEntry.rpc.Callback != nil {
		respInnerMsg, ok := respMsg.(*InnerMessage)
		if !ok {
			pendingRpcEntry.rpc.Callback(ErrorRpcRespMsgType, nil)
			return
		}
		pendingRpcEntry.rpc.Callback(nil, respInnerMsg)
	}
}

// 通过网络连接的sessionId发送消息
func (mgr *RpcManager) SendMsgByConnId(connSessionId int64, msgId int, msg proto.Message) error {
	innerMsg := &InnerMessage{
		Head:  InnerMessageHead{CallId: genNextRpcCallId(), MsgID: msgId},
		PbMsg: msg,
	}

	err := rpcMgr.rpcStubMgr.netcore.TcpSendMsg(connSessionId, innerMsg)
	if err != nil {
		return err
	}
	return nil
}
