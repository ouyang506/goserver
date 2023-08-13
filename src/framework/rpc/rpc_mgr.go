package rpc

import (
	"framework/log"
	"framework/network"
	"framework/proto/pb"
	"framework/registry"
	"reflect"
	"strings"
	"sync"
	"time"
	"utility/timer"

	"google.golang.org/protobuf/proto"
)

// RPC管理器(thread safe)
type RpcManager struct {
	netcore network.NetworkCore

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

func NewRpcManager(mode RpcModeType, msgHandler any,
	options ...Option) *RpcManager {

	mgr := &RpcManager{}

	ops := LoadOptions(options...)

	codecs := []network.Codec{}
	handlers := []network.NetEventHandler{}

	switch mode {
	case RpcModeOuter:
		codecs = append(codecs, NewOuterMessageCodec(mgr))
		codecs = append(codecs, network.NewVariableFrameLenCodec())
		if ops.NetEventHandler != nil {
			handlers = append(handlers, ops.NetEventHandler)
			ops.NetEventHandler.SetOwner(mgr)
		} else {
			outerEventHandler := NewOuterNetEventHandler()
			outerEventHandler.SetOwner(mgr)
			handlers = append(handlers, outerEventHandler)
		}
	case RpcModeInner:
		codecs = append(codecs, NewInnerMessageCodec(mgr))
		codecs = append(codecs, network.NewVariableFrameLenCodec())

		if ops.NetEventHandler != nil {
			handlers = append(handlers, ops.NetEventHandler)
			ops.NetEventHandler.SetOwner(mgr)
		} else {
			outerEventHandler := NewInnerNetEventHandler()
			outerEventHandler.SetOwner(mgr)
			handlers = append(handlers, outerEventHandler)
		}
	}

	// 初始化网络
	mgr.netcore = network.NewNetworkCore(
		network.WithEventHandlers(handlers),
		network.WithLoadBalance(network.NewLoadBalanceRoundRobin(0)),
		network.WithSocketSendBufferSize(32*1024),
		network.WithSocketRcvBufferSize(32*1024),
		network.WithSocketTcpNoDelay(true),
		network.WithFrameCodecs(codecs))
	mgr.netcore.Start()

	mgr.rpcStubMgr = newRpcStubManager(mgr.netcore)

	mgr.pendingRpcMap = map[int64]*PendingRpcEntry{}
	mgr.timerMgr = timer.NewTimerWheel(&timer.Option{TimeAccuracy: time.Millisecond * 200})
	mgr.timerMgr.Start()

	mgr.initMsgHandler(msgHandler)
	return mgr
}

// 设置消息委托处理器
// @param handler 结构体指针
func (mgr *RpcManager) initMsgHandler(handler any) {
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
		if reqMsgId == 0 {
			log.Warn("cannot find the message id by handle method, methodName = %v", methodName)
			continue
		}
		methodMap[reqMsgId] = refValue.Method(i)
	}
	mgr.msgHandleMap = methodMap
}

// 通过消息获取处理函数
func (mgr *RpcManager) GetMsgHandlerFunc(msgId int) *reflect.Value {
	method, ok := mgr.msgHandleMap[msgId]
	if !ok {
		return nil
	}
	return &method
}

// 监听端口
func (mgr *RpcManager) TcpListen(ip string, port int) error {
	return mgr.rpcStubMgr.netcore.TcpListen(ip, port)
}

// 发送网络消息
func (mgr *RpcManager) TcpSendMsg(connSessionId int64, msg any) error {
	err := mgr.rpcStubMgr.netcore.TcpSendMsg(connSessionId, msg)
	if err != nil {
		return err
	}
	return nil
}

// 监视注册中心服务
func (mgr *RpcManager) FetchWatchService(regMgr registry.Registry) {
	cb := func(oper registry.WatchEventType, skey registry.ServiceKey) {
		serverType, ip, port := skey.ServerType, skey.IP, skey.Port
		if serverType == 0 || ip == "" || port == 0 {
			log.Error("handler registry callback error, operate type: %v, key : %v", oper, skey)
			return
		}
		if oper == registry.WatchEventTypeAdd {
			mgr.AddStub(serverType, ip, port)
		} else if oper == registry.WatchEventTypeDelete {
			mgr.DelStub(serverType, ip, port)
		}
	}

	regMgr.FetchAndWatchService(cb)
}

// 添加一个代理管道(注册中心回调调用)
func (mgr *RpcManager) AddStub(serverType int, remoteIp string, remotePort int) bool {
	return mgr.rpcStubMgr.addStub(serverType, remoteIp, remotePort)
}

// 删除一个代理管道(注册中心回调调用)
func (mgr *RpcManager) DelStub(serverType int, remoteIp string, remotePort int) bool {
	return mgr.rpcStubMgr.delStub(serverType, remoteIp, remotePort)
}

// 删除目标类型的全部代理管道
func (mgr *RpcManager) DelStubsByType(serverType int) bool {
	return mgr.rpcStubMgr.delStubsByType(serverType)
}

// 查询代理管道
func (mgr *RpcManager) FindStub(serverType int, remoteIP string, remotePort int) *RpcStub {
	return mgr.rpcStubMgr.findStub(serverType, remoteIP, remotePort)
}

// 添加一个rpc到管理器
func (mgr *RpcManager) AddRpc(rpc *RpcEntry) bool {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	if rpc.TargetSvrType <= 0 {
		log.Error("target server type error, TargetSvrType : %v", rpc.TargetSvrType)
		return false
	}

	rpcStub := mgr.rpcStubMgr.selectStub(rpc)
	if rpcStub == nil {
		log.Error("cannot find rpc stub, serverType : %v", rpc.TargetSvrType)
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
