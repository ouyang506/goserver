package rpc

import (
	"framework/log"
	"framework/network"
	"strconv"
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

func (stub *RpcStub) pushRpc(rpc *Rpc) bool {
	if len(stub.pendingRpcQueue) >= RpcQueueMax {
		return false
	}
	stub.pendingRpcQueue = append(stub.pendingRpcQueue, rpc)
	stub.trySendRpc()
	return true
}

func (stub *RpcStub) removeRpc(callId int64) bool {
	for i, v := range stub.pendingRpcQueue {
		if v.CallId == callId {
			stub.pendingRpcQueue = append(stub.pendingRpcQueue[:i], stub.pendingRpcQueue[i+1:]...)
			return true
		}
	}
	return false
}

func (stub *RpcStub) trySendRpc() {
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
			err := stub.netcore.TcpSendMsg(netconn.(network.Connection).GetSessionId(), rpc.ReqMsg)
			if err != nil {
				break
			}
		}
	}
}

func (stub *RpcStub) onConnected() {
	stub.trySendRpc()
}

func (stub *RpcStub) close() {
	netconn := stub.netconn.Load()
	if netconn != nil {
		connnSessionId := netconn.(network.Connection).GetSessionId()
		stub.netcore.TcpClose(connnSessionId)
	}
}

type RpcStubManger struct {
	rpcMgr      *RpcManager
	netcore     network.NetworkCore
	typeStubMap map[int]*RpcServerTypeStubs
}

type RpcServerTypeStubs struct {
	serverType int
	stubs      []*RpcStub
	router     RpcRouter
}

func newRpcStubManager(rpcMgr *RpcManager) *RpcStubManger {
	mgr := &RpcStubManger{
		rpcMgr: rpcMgr,
	}

	// 初始化网络
	codecs := []network.Codec{}
	codecs = append(codecs, NewInnerMessageCodec())
	codecs = append(codecs, network.NewVariableFrameLenCodec())

	mgr.netcore = network.NewNetworkCore(
		network.WithEventHandler(&RpcNetEvent{rpcMgr: rpcMgr}),
		network.WithLoadBalance(network.NewLoadBalanceRoundRobin(0)),
		network.WithSocketSendBufferSize(32*1024),
		network.WithSocketRcvBufferSize(32*1024),
		network.WithSocketTcpNoDelay(true),
		network.WithFrameCodecs(codecs))
	mgr.netcore.Start()

	mgr.typeStubMap = map[int]*RpcServerTypeStubs{}

	return mgr
}

// 添加一个代理管道
func (mgr *RpcStubManger) addStub(serverType int, instanceId int, remoteIp string, remotePort int) bool {
	typeStubs, ok := mgr.typeStubMap[serverType]
	if !ok {
		typeStubs = &RpcServerTypeStubs{
			serverType: serverType,
			router:     newConsistRouter(),
		}
		mgr.typeStubMap[serverType] = typeStubs
	}

	stub := &RpcStub{
		ServerType: serverType,
		InstanceID: instanceId,
		RemoteIP:   remoteIp,
		RemotePort: remotePort,
		netcore:    mgr.netcore,
	}

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
	// 按照InstanceID排序
	typeStubs.stubs = append(typeStubs.stubs[:index], append([]*RpcStub{stub}, typeStubs.stubs[index:]...)...)
	typeStubs.router.UpdateRoute(strconv.Itoa(stub.InstanceID), stub)

	log.Debug("add rpc stub [%v:%v][%v:%v]", stub.ServerType, stub.InstanceID, stub.RemoteIP, stub.RemotePort)
	return true

}

// 删除一个代理管道
func (mgr *RpcStubManger) delStub(serverType int, instanceId int) bool {
	typeStubs, ok := mgr.typeStubMap[serverType]
	if ok {
		for i, stub := range typeStubs.stubs {
			if stub.InstanceID == instanceId {
				typeStubs.stubs = append(typeStubs.stubs[:i], typeStubs.stubs[i+1:]...)
				typeStubs.router.DelRoute(strconv.Itoa(instanceId))
				stub.close()

				log.Debug("delete rpc stub [%v:%v][%v:%v]", stub.ServerType, stub.InstanceID, stub.RemoteIP, stub.RemotePort)
				return true
			}
		}
	}
	return false
}

// 查询代理管道
func (mgr *RpcStubManger) findStub(serverType int, instanceId int) *RpcStub {
	typeStubs, ok := mgr.typeStubMap[serverType]
	if ok {
		for _, stub := range typeStubs.stubs {
			if stub.InstanceID == instanceId {
				return stub
			}
		}
	}
	return nil
}

// 为rpc分配一个stub
func (mgr *RpcStubManger) selectStub(rpc *Rpc) *RpcStub {
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
