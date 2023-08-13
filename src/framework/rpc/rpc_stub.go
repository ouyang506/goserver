package rpc

import (
	"fmt"
	"framework/log"
	"framework/network"
	"sort"
	"sync"
	"sync/atomic"
)

const (
	AttrRpcStub = "AttrRpcStub"

	RpcQueueMax = 5000
)

type StubSortByIpPort []*RpcStub

func (s StubSortByIpPort) Len() int {
	return len(s)
}
func (s StubSortByIpPort) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s StubSortByIpPort) Less(i, j int) bool {
	return s[i].routeKey() < s[j].routeKey()
}

// rpc网络代理
type RpcStub struct {
	netcore network.NetworkCore

	ServerType int
	RemoteIP   string
	RemotePort int

	pendingRpcQueue []*RpcEntry

	netconn atomic.Value //network.Connection
}

func newRpcStub(netcore network.NetworkCore,
	serverType int, remoteIp string, remotePort int) *RpcStub {
	stub := &RpcStub{
		netcore:         netcore,
		ServerType:      serverType,
		RemoteIP:        remoteIp,
		RemotePort:      remotePort,
		pendingRpcQueue: make([]*RpcEntry, 16),
	}
	stub.init()
	return stub
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

func (stub *RpcStub) routeKey() string {
	return fmt.Sprintf("%s:%d", stub.RemoteIP, stub.RemotePort)
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
					CallId:  rpc.CallId,
					MsgID:   rpc.MsgId,
					Guid:    0,
					Content: rpc.ReqMsg,
				}
			} else if rpc.RpcMode == RpcModeOuter {
				msg = &OuterMessage{
					CallId:  rpc.CallId,
					MsgID:   rpc.MsgId,
					Content: rpc.ReqMsg,
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

// 网络代理管道管理器
type RpcStubManger struct {
	netcore network.NetworkCore

	stubs   map[int]*ServerTypeStubs
	stubsMu sync.RWMutex
}

type ServerTypeStubs struct {
	serverType int
	stubs      []*RpcStub
	router     RpcRouter
}

func newRpcStubManager(netcore network.NetworkCore) *RpcStubManger {
	mgr := &RpcStubManger{
		netcore: netcore,
		stubs:   map[int]*ServerTypeStubs{},
	}
	return mgr
}

// 添加一个代理管道
func (mgr *RpcStubManger) addStub(serverType int, remoteIp string, remotePort int) bool {
	mgr.stubsMu.Lock()
	defer mgr.stubsMu.Unlock()

	typeStubs, ok := mgr.stubs[serverType]
	if !ok {
		typeStubs = &ServerTypeStubs{
			serverType: serverType,
			router:     newConsistRouter(),
		}
		mgr.stubs[serverType] = typeStubs
	}
	stub := newRpcStub(mgr.netcore, serverType, remoteIp, remotePort)

	typeStubs.stubs = append(typeStubs.stubs, stub)
	sort.Sort(StubSortByIpPort(typeStubs.stubs))

	typeStubs.router.UpdateRoute(stub.routeKey(), stub)
	log.Info("add rpc stub [%v][%v:%v]", stub.ServerType, stub.RemoteIP, stub.RemotePort)
	return true
}

// 删除一个代理管道
func (mgr *RpcStubManger) delStub(serverType int, remoteIp string, remotePort int) bool {
	mgr.stubsMu.Lock()
	defer mgr.stubsMu.Unlock()

	typeStubs, ok := mgr.stubs[serverType]
	if ok {
		for i, stub := range typeStubs.stubs {
			if stub.RemoteIP == remoteIp && stub.RemotePort == remotePort {
				typeStubs.stubs = append(typeStubs.stubs[:i], typeStubs.stubs[i+1:]...)
				typeStubs.router.DelRoute(stub.routeKey())
				stub.close()
				log.Info("delete rpc stub [%v][%v:%v]", stub.ServerType, stub.RemoteIP, stub.RemotePort)
				return true
			}
		}
	}
	return false
}

// 删除一个类型的代理管道
func (mgr *RpcStubManger) delStubsByType(serverType int) bool {
	mgr.stubsMu.Lock()
	defer mgr.stubsMu.Unlock()

	typeStubs, ok := mgr.stubs[serverType]
	if !ok {
		return false
	}
	if len(typeStubs.stubs) <= 0 {
		return false
	}
	for _, stub := range typeStubs.stubs {
		typeStubs.router.DelRoute(stub.routeKey())
		stub.close()
		log.Debug("delete rpc stub [%v][%v:%v]", stub.ServerType, stub.RemoteIP, stub.RemotePort)
	}

	delete(mgr.stubs, serverType)
	return true
}

// 查询代理管道
func (mgr *RpcStubManger) findStub(serverType int, remoteIP string, remotePort int) *RpcStub {
	mgr.stubsMu.RLock()
	defer mgr.stubsMu.RUnlock()

	typeStubs, ok := mgr.stubs[serverType]
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
	mgr.stubsMu.RLock()
	defer mgr.stubsMu.RUnlock()

	typeStubs, ok := mgr.stubs[rpc.TargetSvrType]
	if !ok {
		return nil
	}

	stub := typeStubs.router.SelectRoute(rpc.RouteKey)
	if stub == nil {
		return nil
	}

	return stub.(*RpcStub)
}
