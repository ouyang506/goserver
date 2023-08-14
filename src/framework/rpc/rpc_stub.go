package rpc

import (
	"fmt"
	"framework/log"
	"framework/network"
	"sort"
	"sync"
	"time"
)

const (
	RpcQueueMax = 4096
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

	rpcChan chan (*RpcEntry)

	netconn network.Connection
	connMu  sync.RWMutex
}

func newRpcStub(netcore network.NetworkCore,
	serverType int, remoteIp string, remotePort int) *RpcStub {
	stub := &RpcStub{
		netcore:    netcore,
		ServerType: serverType,
		RemoteIP:   remoteIp,
		RemotePort: remotePort,
		rpcChan:    make(chan (*RpcEntry), RpcQueueMax),
	}
	return stub
}

func (stub *RpcStub) routeKey() string {
	return fmt.Sprintf("%s:%d", stub.RemoteIP, stub.RemotePort)
}

// connect rpc remote server
func (stub *RpcStub) lazyInitConn() {
	stub.connMu.Lock()
	defer stub.connMu.Unlock()

	if stub.netconn != nil {
		return
	}

	//初始化netconn进行网络连接
	netconn, err := stub.netcore.TcpConnect(stub.RemoteIP, stub.RemotePort, true)
	if err != nil {
		log.Error("stub try to connect error : %v, stub: %+v", err, stub)
	}
	stub.netconn = netconn

	go stub.loopSend()
}

func (stub *RpcStub) loopSend() {
	var current *RpcEntry = nil
	for {
		// netconn是否处于连接状态
		if stub.netconn.GetState() != network.ConnStateConnected {
			time.Sleep(200 * time.Millisecond)
			continue
		}

		var rpc *RpcEntry = nil
		if current != nil {
			rpc = current
		} else {
			rpc = stub.waitPopRpc()
		}

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

		err := stub.netcore.TcpSendMsg(stub.netconn.GetSessionId(), msg)
		if err != nil {
			log.Error("rpc stub loop send msg error: %v", err)
			continue
		}
		current = nil
	}
}

func (stub *RpcStub) pushRpc(rpc *RpcEntry) bool {
	// 收到待发送的rpc时，启动连接
	stub.lazyInitConn()

	select {
	case stub.rpcChan <- rpc:
		return true
	default:
		return false
	}
}

func (stub *RpcStub) waitPopRpc() *RpcEntry {
	for {
		rpc := <-stub.rpcChan

		// 队列里会有脏数据，需要清除掉已经超时的rpc
		if rpc.getTimeoutFlag() {
			continue
		}
		return rpc
	}
}

func (stub *RpcStub) close() {
	stub.connMu.Lock()
	defer stub.connMu.Unlock()

	if stub.netconn != nil {
		stub.netcore.TcpClose(stub.netconn.GetSessionId())
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

// 添加rpc到管道代理
func (mgr *RpcStubManger) addRpc(rpc *RpcEntry) bool {
	rpcStub := mgr.selectStub(rpc)
	if rpcStub == nil {
		log.Error("cannot find rpc stub, serverType : %v", rpc.TargetSvrType)
		return false
	}

	if !rpcStub.pushRpc(rpc) {
		log.Error("push rpc failed, callId : %v", rpc.CallId)
		return false
	}

	return true
}
