package rpc

import (
	"framework/log"
	"framework/network"
	"framework/proto/pb"
	"reflect"
	"strings"
	"sync"
	"time"
	"utility/timer"

	"google.golang.org/protobuf/proto"
)

// RPC管理器(thread safe)
type RpcManager struct {
	mutex      sync.Mutex
	rpcStubMgr *RpcStubManger

	pendingRpcMap map[int64]*PendingRpcEntry
	timerMgr      *timer.TimerWheel

	msgHandleMap map[int]reflect.Value
}

type PendingRpcEntry struct {
	rpc     *RpcEntry
	rpcStub *RpcStub
}

func NewRpcManager(mode RpcModeType, userNetEventHandler network.NetEventHandler,
	msgHandler any) *RpcManager {

	mgr := &RpcManager{}

	codecs := []network.Codec{}
	handlers := []network.NetEventHandler{}

	switch mode {
	case RpcModeOuter:
		codecs = append(codecs, NewOuterMessageCodec(mgr))
		codecs = append(codecs, network.NewVariableFrameLenCodec())

		outerEventHandler := NewOuterNetEvent(mgr)
		handlers = append(handlers, outerEventHandler)
		handlers = append(handlers, userNetEventHandler)
	case RpcModeInner:
		fallthrough
	default:
		codecs = append(codecs, NewInnerMessageCodec(mgr))
		codecs = append(codecs, network.NewVariableFrameLenCodec())

		innerEventHandler := NewInnerNetEvent(mgr)
		handlers = append(handlers, innerEventHandler)
		handlers = append(handlers, userNetEventHandler)
	}

	mgr.rpcStubMgr = newRpcStubManager(mgr, codecs, handlers)
	mgr.pendingRpcMap = map[int64]*PendingRpcEntry{}
	mgr.timerMgr = timer.NewTimerWheel(&timer.Option{TimeAccuracy: time.Millisecond * 200})
	mgr.timerMgr.Start()

	mgr.SetMsgHandler(msgHandler)
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
func (mgr *RpcManager) AddRpc(rpc *RpcEntry) bool {
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

// 通过callId获取rpc信息
func (mgr *RpcManager) GetRpc(callId int64) *RpcEntry {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()
	pending, ok := mgr.pendingRpcMap[callId]
	if !ok {
		return nil
	}
	return pending.rpc
}

// 收到rpc返回后回调
func (mgr *RpcManager) OnRcvResponse(sessionId int64, respMsg proto.Message) {
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
}

// 通过网络连接的sessionId发送消息
func (mgr *RpcManager) TcpSendMsg(connSessionId int64, msg any) error {
	err := mgr.rpcStubMgr.netcore.TcpSendMsg(connSessionId, msg)
	if err != nil {
		return err
	}
	return nil
}

func (mgr *RpcManager) SetMsgHandler(handler any) {
	methodMap := make(map[int]reflect.Value)
	refType := reflect.TypeOf(handler)
	refValue := reflect.ValueOf(handler)
	methodCount := refType.NumMethod()
	for i := 0; i < methodCount; i++ {
		methodName := refType.Method(i).Name
		if !strings.HasPrefix(methodName, RpcHandlerMethodPrefix) {
			continue
		}
		reqMsgName := methodName[len(RpcHandlerMethodPrefix):]
		reqMsgId := pb.GetMsgIdByName(reqMsgName)
		methodMap[reqMsgId] = refValue.Method(i)
	}

	mgr.msgHandleMap = methodMap
}

func (mgr *RpcManager) GetMsgHandlerFunc(msgId int) *reflect.Value {
	method, ok := mgr.msgHandleMap[msgId]
	if !ok {
		return nil
	}
	return &method
}
