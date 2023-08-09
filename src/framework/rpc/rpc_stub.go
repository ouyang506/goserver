package rpc

import (
	"fmt"
	"framework/log"
	"framework/network"
	"sort"
	"sync/atomic"
)

const (
	AttrRpcStub = "AttrRpcStub"

	RpcQueueMax = 5000
)

// rpc网络代理
type RpcStub struct {
	ServerType int
	RemoteIP   string
	RemotePort int

	pendingRpcQueue []*RpcEntry

	netcore network.NetworkCore
	netconn atomic.Value //network.Connection
}

type StubSortByIpPort []*RpcStub

func (s StubSortByIpPort) Len() int {
	return len(s)
}
func (s StubSortByIpPort) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s StubSortByIpPort) Less(i, j int) bool {
	return s[i].key() < s[j].key()
}

func (stub *RpcStub) key() string {
	return fmt.Sprintf("%s:%d", stub.RemoteIP, stub.RemotePort)
}
func (stub *RpcStub) init() {
	//初始化netconn进行网络连接
	atrrib := map[interface{}]interface{}{}
	atrrib[AttrRpcStub] = stub
	netconn, err := stub.netcore.TcpConnect(stub.RemoteIP, stub.RemotePort, true, atrrib)
	if err != nil {
		log.Error("stub try to connect error : %v, stub: %+v", err, stub)
	}
	stub.netconn.Store(netconn)
}

func (stub *RpcStub) pushRpc(rpc *RpcEntry) bool {
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
	// netconn处于连接状态
	netconn := stub.netconn.Load()
	if netconn != nil && netconn.(network.Connection).GetConnState() == network.ConnStateConnected {
		for {
			if len(stub.pendingRpcQueue) <= 0 {
				break
			}
			rpc := stub.pendingRpcQueue[0]
			stub.pendingRpcQueue = stub.pendingRpcQueue[1:]
			var msg any = nil
			if rpc.RpcMode == RpcModeInner {
				msg = &InnerMessage{
					Head: InnerMessageHead{
						CallId: rpc.CallId,
						MsgID:  rpc.MsgId,
						Guid:   0,
					},
					PbMsg: rpc.ReqMsg,
				}
			} else if rpc.RpcMode == RpcModeOuter {
				msg = &OuterMessage{
					Head: OuterMessageHead{
						CallId: rpc.CallId,
						MsgID:  rpc.MsgId,
					},
					PbMsg: rpc.ReqMsg,
				}

			}
			err := stub.netcore.TcpSendMsg(netconn.(network.Connection).GetSessionId(), msg)
			if err != nil {
				break
			}
		}
	}
}

func (stub *RpcStub) onConnected() {
	go stub.trySendRpc()
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

func newRpcStubManager(rpcMgr *RpcManager, codecs []network.Codec, eventHandlers []network.NetEventHandler) *RpcStubManger {
	mgr := &RpcStubManger{
		rpcMgr: rpcMgr,
	}

	// 初始化网络
	mgr.netcore = network.NewNetworkCore(
		network.WithEventHandlers(eventHandlers),
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
func (mgr *RpcStubManger) addStub(serverType int, remoteIp string, remotePort int) bool {
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
		RemoteIP:   remoteIp,
		RemotePort: remotePort,
		netcore:    mgr.netcore,
	}
	stub.init()

	typeStubs.stubs = append(typeStubs.stubs, stub)
	sort.Sort(StubSortByIpPort(typeStubs.stubs))

	typeStubs.router.UpdateRoute(stub.key(), stub)

	log.Debug("add rpc stub [%v][%v:%v]", stub.ServerType, stub.RemoteIP, stub.RemotePort)
	return true
}

// 删除一个代理管道
func (mgr *RpcStubManger) delStub(serverType int, remoteIp string, remotePort int) bool {
	typeStubs, ok := mgr.typeStubMap[serverType]
	if ok {
		for i, stub := range typeStubs.stubs {
			if stub.RemoteIP == remoteIp && stub.RemotePort == remotePort {
				typeStubs.stubs = append(typeStubs.stubs[:i], typeStubs.stubs[i+1:]...)
				typeStubs.router.DelRoute(stub.key())
				stub.close()

				log.Debug("delete rpc stub [%v][%v:%v]", stub.ServerType, stub.RemoteIP, stub.RemotePort)
				return true
			}
		}
	}
	return false
}

// 删除一个类型的代理管道
func (mgr *RpcStubManger) delTypeStubs(serverType int) bool {
	typeStubs, ok := mgr.typeStubMap[serverType]
	if !ok {
		return false
	}
	if len(typeStubs.stubs) <= 0 {
		return false
	}
	for _, stub := range typeStubs.stubs {
		typeStubs.router.DelRoute(stub.key())
		stub.close()

		log.Debug("delete rpc stub [%v][%v:%v]", stub.ServerType, stub.RemoteIP, stub.RemotePort)
	}

	delete(mgr.typeStubMap, serverType)
	return true
}

// 查询代理管道
func (mgr *RpcStubManger) findStub(serverType int, remoteIP string, remotePort int) *RpcStub {
	typeStubs, ok := mgr.typeStubMap[serverType]
	if ok {
		for _, stub := range typeStubs.stubs {
			if stub.RemoteIP == remoteIP && stub.RemotePort == remotePort {
				return stub
			}
		}
	}
	return nil
}

// 为rpc分配一个stub
func (mgr *RpcStubManger) selectStub(rpc *RpcEntry) *RpcStub {
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
