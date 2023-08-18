package rpc

import (
	"framework/log"
	"framework/network"
	"framework/proto/pb"
	"framework/registry"
	"reflect"
	"strings"
	"sync"

	"google.golang.org/protobuf/proto"
)

// RPC管理器(thread safe)
type RpcManager struct {
	netcore      network.NetworkCore
	rpcStubMgr   *RpcStubManger
	msgHandleMap map[int]reflect.Value

	rpcMap sync.Map //rpcCallId->*RpcEntry
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
		network.WithSocketSendBufferSize(1024),
		network.WithSocketRcvBufferSize(1024),
		network.WithSocketTcpNoDelay(true),
		network.WithFrameCodecs(codecs))
	mgr.netcore.Start()

	mgr.rpcStubMgr = newRpcStubManager(mgr.netcore)

	mgr.rpcMap = sync.Map{}

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
	// 需要先store，否则可能response已经收到，根据callId获取rpc失败
	mgr.rpcMap.Store(rpc.CallId, rpc)
	if !mgr.rpcStubMgr.addRpc(rpc) {
		log.Error("add rpc failed, target server type: %v, callId: %v", rpc.TargetSvrType, rpc.CallId)
		mgr.rpcMap.Delete(rpc.CallId)
		return false
	}

	// not need waiting response
	if rpc.IsOneway {
		return true
	}

	return true
}

func (mgr *RpcManager) RemoveRpc(callId int64) bool {
	_, ok := mgr.rpcMap.LoadAndDelete(callId)
	return ok
}

// 通过callId获取rpc信息
func (mgr *RpcManager) GetRpc(callId int64) *RpcEntry {
	rpc, ok := mgr.rpcMap.Load(callId)
	if !ok {
		return nil
	}
	return rpc.(*RpcEntry)
}

// 收到rpc返回后回调
func (mgr *RpcManager) OnRcvResponse(callId int64, respMsg proto.Message) {
	rpc := mgr.GetRpc(callId)
	if rpc == nil {
		log.Info("rpc not found, callId: %v", callId)
		return
	}

	select {
	case rpc.NotifyChan <- struct{}{}:
		//log.Debug("push msg to resp channel ok, callId = %v", callId)
		break
	default:
		log.Error("push msg to resp channel error,len=%v, callId = %v", len(rpc.NotifyChan), callId)
	}
}
