package rpc

import (
	"common/log"
	"common/network"
	"common/registry"
	"strconv"
	"strings"
)

const (
	NetConnAttrRpcStub = "RpcStub"
)

// rpc网络代理
type RpcStub struct {
	ServerType int
	InstanceID int
	RemoteIP   string
	RemotePort int

	pendingRpcQueue []*Rpc
	pendingRpcMap   map[int64]*Rpc
	netcore         network.NetworkCore
	netconn         network.Connection
}

func (stub *RpcStub) PushRpc(rpc *Rpc) {
	if _, ok := stub.pendingRpcMap[rpc.SessionID]; ok {
		return
	}
	stub.pendingRpcMap[rpc.SessionID] = rpc
	stub.pendingRpcQueue = append(stub.pendingRpcQueue, rpc)
}

func (stub *RpcStub) SendRpc() {
	if stub.netconn == nil || stub.netconn.GetConnState() != network.ConnStateConnected {
		//进行网络连接
		netconn, err := stub.netcore.TcpConnect(stub.RemoteIP, stub.RemotePort, true)
		if err != nil {
			log.Error("stub try to connect error : %v, stub: %+v", err, stub)
		} else {
			netconn.SetAttrib(NetConnAttrRpcStub, stub)
			stub.netconn = netconn
		}
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
	return true
}

func (typeStubs *RpcServerTypeStubs) delStub(instanceID int) bool {
	for i, v := range typeStubs.stubs {
		if v.InstanceID == instanceID {
			typeStubs.stubs = append(typeStubs.stubs[:i], typeStubs.stubs[i+1:]...)
			typeStubs.router.DelRoute(strconv.Itoa(instanceID))
			return true
		}
	}
	return false
}

func (typeStubs *RpcServerTypeStubs) getStub(instanceID int) *RpcStub {
	for _, v := range typeStubs.stubs {
		if v.InstanceID == instanceID {
			return v
		}
	}
	return nil
}

type RpcStubManger struct {
	netcore     network.NetworkCore
	typeStubMap map[int]*RpcServerTypeStubs
	registry    *registry.RegistryMgr
}

func NewRpcStubManager() *RpcStubManger {
	mgr := &RpcStubManger{}

	mgr.netcore = network.NewNetworkCore(network.WithLogger(log.GetLogger()),
		network.WithEventHandler(mgr),
		network.WithLoadBalance(network.NewLoadBalanceRoundRobin(0)),
		network.WithSocketSendBufferSize(10240),
		network.WithSocketRcvBufferSize(10240),
		network.WithSocketTcpNoDelay(true),
		network.WithFrameCodec(network.NewVariableFrameLenCodec()))

	mgr.typeStubMap = map[int]*RpcServerTypeStubs{}

	etcdReg := registry.NewEtcdRegistry(log.GetLogger(), []string{"127.0.0.1:2379"}, "", "")
	mgr.registry = registry.NewRegistryMgr(log.GetLogger(), etcdReg, mgr.registryCB)
	mgr.registry.DoRegister("1", "1", 10)

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
			stub := &RpcStub{
				ServerType: serverType,
				InstanceID: instanceID,
				RemoteIP:   ip,
				RemotePort: port,
			}
			mgr.AddStub(stub)
		}
	case registry.OperDelete:
		{
			mgr.DelStub(serverType, instanceID)
		}
	}
}

func (mgr *RpcStubManger) AddStub(stub *RpcStub) bool {
	serverTypeStubs, ok := mgr.typeStubMap[stub.ServerType]
	if !ok {
		serverTypeStubs = &RpcServerTypeStubs{
			serverType: stub.ServerType,
			router:     newConsistRouter(),
		}
		mgr.typeStubMap[stub.ServerType] = serverTypeStubs
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

	memberName := typeStubs.router.SelectServer(rpc.RouteKey).(string)
	if memberName == "" {
		return nil
	}

	instanceID, _ := strconv.Atoi(memberName)
	stub := typeStubs.getStub(instanceID)

	return stub
}

// 网络事件回调
func (mgr *RpcStubManger) OnAccept(c network.Connection) {
	peerHost, peerPort := c.GetPeerAddr()
	log.Debug("rpc stub manager OnAccept, peerHost:%v, peerPort:%v", peerHost, peerPort)
}

func (mgr *RpcStubManger) OnConnected(c network.Connection) {
	peerHost, peerPort := c.GetPeerAddr()
	log.Debug("rpc stub manager OnConnected, peerHost:%v, peerPort:%v", peerHost, peerPort)
}

func (mgr *RpcStubManger) OnConnectFailed(c network.Connection) {
	peerHost, peerPort := c.GetPeerAddr()
	log.Debug("rpc stub manager OnConnectFailed, peerHost:%v, peerPort:%v", peerHost, peerPort)
}

func (mgr *RpcStubManger) OnClosed(c network.Connection) {
	peerHost, peerPort := c.GetPeerAddr()
	log.Debug("rpc stub manager OnClosed, peerHost:%v, peerPort:%v", peerHost, peerPort)

}
