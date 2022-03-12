package rpc

import (
	"common/log"
	"common/network"
	"common/registry"
	"strconv"
	"strings"
	"sync/atomic"
)

const (
	AttrRpcStub = "AttrRpcStub"

	RpcQueueMax = 5000
)

// rpc网络代理
type RpcStub struct {
	ServerType int
	InstanceID int
	RemoteIP   string
	RemotePort int

	pendingRpcQueue []*Rpc

	netcore       network.NetworkCore
	netconn       atomic.Value //network.Connection
	netconnInited int32
}

func (stub *RpcStub) PushRpc(rpc *Rpc) bool {
	if len(stub.pendingRpcQueue) >= RpcQueueMax {
		return false
	}
	stub.pendingRpcQueue = append(stub.pendingRpcQueue, rpc)
	stub.TrySendRpc()
	return true
}

func (stub *RpcStub) RemoveRpc(callId int64) bool {
	for i, v := range stub.pendingRpcQueue {
		if v.CallId == callId {
			stub.pendingRpcQueue = append(stub.pendingRpcQueue[:i], stub.pendingRpcQueue[i+1:]...)
			return true
		}
	}
	return false
}

func (stub *RpcStub) TrySendRpc() {
	//初始化netconn进行网络连接
	if atomic.CompareAndSwapInt32(&stub.netconnInited, 0, 1) {
		atrrib := map[interface{}]interface{}{}
		atrrib[AttrRpcStub] = stub
		netconn, err := stub.netcore.TcpConnect(stub.RemoteIP, stub.RemotePort, true, atrrib)
		if err != nil {
			log.Error("stub try to connect error : %v, stub: %+v", err, stub)
		}
		stub.netconn.Store(netconn)
	}

	// netconn处于连接状态
	netconn := stub.netconn.Load()
	if netconn != nil && netconn.(network.Connection).GetConnState() == network.ConnStateConnected {
		for {
			if len(stub.pendingRpcQueue) <= 0 {
				break
			}
			rpc := stub.pendingRpcQueue[0]
			stub.pendingRpcQueue = stub.pendingRpcQueue[1:]
			err := stub.netcore.TcpSend(netconn.(network.Connection).GetSessionId(), rpc.Request)
			if err != nil {
				break
			}
		}
	}
}

func (stub *RpcStub) Close() {
	netconn := stub.netconn.Load()
	if netconn != nil {
		connnSessionId := netconn.(network.Connection).GetSessionId()
		stub.netcore.TcpClose(connnSessionId)
	}
}

type RpcServerTypeStubs struct {
	serverType int
	stubs      []*RpcStub
	router     RpcRouter
}

func (typeStubs *RpcServerTypeStubs) addStub(stub *RpcStub) bool {
	index := 0
	for i, v := range typeStubs.stubs {
		if v.InstanceID == stub.InstanceID {
			return false
		}
		if v.InstanceID > stub.InstanceID {
			index = i
			break
		}
	}
	// 按照instanceID排序
	typeStubs.stubs = append(typeStubs.stubs[:index], append([]*RpcStub{stub}, typeStubs.stubs[index:]...)...)
	typeStubs.router.UpdateRoute(strconv.Itoa(stub.InstanceID), stub)

	log.Debug("add rpc stub [%v:%v][%v:%v]", stub.ServerType, stub.InstanceID, stub.RemoteIP, stub.RemotePort)
	return true
}

func (typeStubs *RpcServerTypeStubs) delStub(instanceID int) bool {
	for i, v := range typeStubs.stubs {
		if v.InstanceID == instanceID {
			typeStubs.stubs = append(typeStubs.stubs[:i], typeStubs.stubs[i+1:]...)
			typeStubs.router.DelRoute(strconv.Itoa(instanceID))

			log.Debug("delete rpc stub [%v:%v][%v:%v]", v.ServerType, v.InstanceID, v.RemoteIP, v.RemotePort)
			return true
		}
	}
	return false
}

// func (typeStubs *RpcServerTypeStubs) getStub(instanceID int) *RpcStub {
// 	for _, v := range typeStubs.stubs {
// 		if v.InstanceID == instanceID {
// 			return v
// 		}
// 	}
// 	return nil
// }

type RpcStubManger struct {
	netcore     network.NetworkCore
	typeStubMap map[int]*RpcServerTypeStubs
	registry    *registry.RegistryMgr
}

func NewRpcStubManager() *RpcStubManger {
	mgr := &RpcStubManger{}

	mgr.netcore = network.NewNetworkCore(
		network.WithEventHandler(mgr),
		network.WithLoadBalance(network.NewLoadBalanceRoundRobin(0)),
		network.WithSocketSendBufferSize(10240),
		network.WithSocketRcvBufferSize(10240),
		network.WithSocketTcpNoDelay(true),
		network.WithFrameCodec(network.NewVariableFrameLenCodec()))
	mgr.netcore.Start()

	mgr.typeStubMap = map[int]*RpcServerTypeStubs{}

	etcdReg := registry.NewEtcdRegistry(log.GetLogger(), []string{"127.0.0.1:2379"}, "", "")
	mgr.registry = registry.NewRegistryMgr(log.GetLogger(), etcdReg, mgr.registryCB)

	return mgr
}

func (mgr *RpcStubManger) registryCB(oper registry.OperType, key string, value string) {
	serverType := 0
	instanceID := 0
	splitKeyArr := strings.SplitN(key, ":", 2)
	if len(splitKeyArr) >= 2 {
		serverType, _ = strconv.Atoi(splitKeyArr[0])
		instanceID, _ = strconv.Atoi(splitKeyArr[1])
	}

	switch oper {
	case registry.OperUpdate:
		{
			ip := ""
			port := 0
			splitValueArr := strings.SplitN(value, ":", 2)
			if len(splitValueArr) >= 2 {
				ip = splitKeyArr[0]
				port, _ = strconv.Atoi(splitKeyArr[1])
			}
			if ip == "" || port <= 0 {
				return
			}
			mgr.AddStub(serverType, instanceID, ip, port)
		}
	case registry.OperDelete:
		{
			mgr.DelStub(serverType, instanceID)
		}
	}
}

func (mgr *RpcStubManger) AddStub(serverType int, instanceId int, remoteIp string, remotePort int) bool {
	serverTypeStubs, ok := mgr.typeStubMap[serverType]
	if !ok {
		serverTypeStubs = &RpcServerTypeStubs{
			serverType: serverType,
			router:     newConsistRouter(),
		}
		mgr.typeStubMap[serverType] = serverTypeStubs
	}

	stub := &RpcStub{
		ServerType: serverType,
		InstanceID: instanceId,
		RemoteIP:   remoteIp,
		RemotePort: remotePort,
		netcore:    mgr.netcore,
	}
	return serverTypeStubs.addStub(stub)
}

func (mgr *RpcStubManger) DelStub(serverType int, instanceId int) bool {
	typeStubs, ok := mgr.typeStubMap[serverType]
	if ok {
		return typeStubs.delStub(instanceId)
	}
	return false
}

func (mgr *RpcStubManger) FindStub(rpc *Rpc) *RpcStub {
	typeStubs, ok := mgr.typeStubMap[rpc.TargetSvrType]
	if !ok {
		return nil
	}

	stub := typeStubs.router.SelectServer(rpc.RouteKey)
	if stub == nil {
		return nil
	}

	return stub.(*RpcStub)
}

// 网络事件回调
func (mgr *RpcStubManger) OnAccept(c network.Connection) {
	peerHost, peerPort := c.GetPeerAddr()
	log.Info("rpc stub manager OnAccept, peerHost:%v, peerPort:%v", peerHost, peerPort)
}

func (mgr *RpcStubManger) OnConnect(c network.Connection, err error) {
	peerHost, peerPort := c.GetPeerAddr()
	if err != nil {
		log.Info("rpc stub manager OnConnectFailed, sessionId: %v, peerHost:%v, peerPort:%v", c.GetSessionId(), peerHost, peerPort)
	} else {
		log.Info("rpc stub manager OnConnected, sessionId: %v, peerHost:%v, peerPort:%v,", c.GetSessionId(), peerHost, peerPort)
		stub, ok := c.GetAttrib(AttrRpcStub)
		if !ok || stub == nil {
			return
		}
		stub.(*RpcStub).TrySendRpc()
	}
}

func (mgr *RpcStubManger) OnClosed(c network.Connection) {
	peerHost, peerPort := c.GetPeerAddr()
	log.Info("rpc stub manager OnClosed, peerHost:%v, peerPort:%v", peerHost, peerPort)
}
