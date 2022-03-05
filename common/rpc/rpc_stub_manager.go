package rpc

import (
	"common/log"
	"common/network"
)

// rpc网络代理
type RpcStub struct {
	ServerType       int
	ServerTypeInstID int
	RemoteIP         string
	RemotePort       int
	NetSessionId     int64
}

type IdToStubMap map[int]*RpcStub

type RpcStubManger struct {
	serverTypeStubs map[int]IdToStubMap
	netcore         network.NetworkCore
}

func NewRpcStubManager() *RpcStubManger {
	mgr := &RpcStubManger{}
	mgr.serverTypeStubs = make(map[int]IdToStubMap)

	mgr.netcore = network.NewNetworkCore(network.WithLogger(log.GetLogger()),
		network.WithEventHandler(&StubNetEventHandler{rpcStubMgr: mgr}),
		network.WithLoadBalance(network.NewLoadBalanceRoundRobin(0)),
		network.WithSocketSendBufferSize(10240),
		network.WithSocketRcvBufferSize(10240),
		network.WithSocketTcpNoDelay(true),
		network.WithFrameCodec(network.NewVariableFrameLenCodec()))

	return mgr
}

func (mgr *RpcStubManger) AddStub(stub *RpcStub) {
	typeStubs, ok := mgr.serverTypeStubs[stub.ServerType]
	if !ok {
		mgr.serverTypeStubs[stub.ServerType] = make(IdToStubMap)
		mgr.serverTypeStubs[stub.ServerType][stub.ServerTypeInstID] = stub
	} else {
		typeStubs[stub.ServerTypeInstID] = stub
	}
}

func (mgr *RpcStubManger) DelStub(stub *RpcStub) {
	typeStubs, ok := mgr.serverTypeStubs[stub.ServerType]
	if ok {
		delete(typeStubs, stub.ServerTypeInstID)
	}
}

func (mgr *RpcStubManger) GetStub(serverType int) *RpcStub {
	typeStubs, ok := mgr.serverTypeStubs[serverType]
	if !ok {
		return nil
	}
	if len(typeStubs) <= 0 {
		return nil
	}

	for _, v := range typeStubs {
		return v
	}

	return nil
}

type StubNetEventHandler struct {
	rpcStubMgr *RpcStubManger
}

func (h *StubNetEventHandler) OnAccept(c network.Connection) {
	peerHost, peerPort := c.GetPeerAddr()
	log.Debug("StubNetEventHandler OnAccept, peerHost:%v, peerPort:%v", peerHost, peerPort)
}

func (h *StubNetEventHandler) OnConnected(c network.Connection) {
	peerHost, peerPort := c.GetPeerAddr()
	log.Debug("StubNetEventHandler OnConnected, peerHost:%v, peerPort:%v", peerHost, peerPort)
}

func (h *StubNetEventHandler) OnClosed(c network.Connection) {
	peerHost, peerPort := c.GetPeerAddr()
	log.Debug("StubNetEventHandler OnClosed, peerHost:%v, peerPort:%v", peerHost, peerPort)
}
