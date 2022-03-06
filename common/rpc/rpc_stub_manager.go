package rpc

import (
	"common/log"
	"common/network"
	"common/utility/consistent"
	"fmt"
	"strconv"
	"strings"
)

const (
	NetConnAttrRpcStub = "RpcStub"
)

// rpc网络代理
type RpcStub struct {
	ServerType       int // key = ServerType + ServerTypeInstID
	ServerTypeInstID int
	RemoteIP         string
	RemotePort       int
	netconn          network.Connection
}

type RpcStubManger struct {
	netcore         network.NetworkCore
	serverTypeStubs map[int][]*RpcStub

	consisRouterMap map[int]*consistent.Consistent // rpc一致性路由容器
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

	mgr.serverTypeStubs = map[int][]*RpcStub{}
	mgr.consisRouterMap = map[int]*consistent.Consistent{}

	return mgr
}

func (mgr *RpcStubManger) AddStub(stub *RpcStub) {
	typeStubs, ok := mgr.serverTypeStubs[stub.ServerType]
	if !ok {
		mgr.serverTypeStubs[stub.ServerType] = []*RpcStub{stub}
	} else {
		// 无法添加相同的stub
		for _, stubTmp := range typeStubs {
			if stubTmp.ServerType == stub.ServerType && stubTmp.ServerTypeInstID == stub.ServerTypeInstID {
				return
			}
		}
		mgr.serverTypeStubs[stub.ServerType] = append(typeStubs, stub)
	}

	// 进行网络连接
	netconn, err := mgr.netcore.TcpConnect(stub.RemoteIP, stub.RemotePort)
	if err != nil {
		log.Error("add stub try to connect error : %v, stub: %+v", err, stub)
	} else {
		netconn.SetAttrib(NetConnAttrRpcStub, stub)
		stub.netconn = netconn
	}
}

func (mgr *RpcStubManger) DelStub(stub *RpcStub) bool {
	typeStubs, ok := mgr.serverTypeStubs[stub.ServerType]
	if ok {
		for i, stubTmp := range typeStubs {
			if stubTmp.ServerType == stub.ServerType && stubTmp.ServerTypeInstID == stub.ServerTypeInstID {
				mgr.serverTypeStubs[stub.ServerType] = append(typeStubs[:i], typeStubs[i+1:]...)
				return true
			}
		}
	}
	return false
}

func (mgr *RpcStubManger) GetStub(serverType int, serverTypeInstID int) *RpcStub {
	typeStubs, ok := mgr.serverTypeStubs[serverType]
	if !ok {
		return nil
	}
	if len(typeStubs) <= 0 {
		return nil
	}

	for _, stub := range typeStubs {
		if stub.ServerType == serverType && stub.ServerTypeInstID == serverTypeInstID {
			return stub
		}
	}
	return nil
}

func (mgr *RpcStubManger) GetHashRouteStub(serverType int, hashKey string) *RpcStub {
	router, ok := mgr.consisRouterMap[serverType]
	if !ok {
		return nil
	}

	routeKey := router.Get(hashKey)
	if routeKey == "" {
		return nil
	}

	splitArr := strings.SplitN(routeKey, ":", 2)
	if len(splitArr) == 2 {
		serverType, _ := strconv.Atoi(splitArr[0])
		serverTypeInstID, _ := strconv.Atoi(splitArr[0])
		return mgr.GetStub(serverType, serverTypeInstID)
	}
	return nil
}

func (mgr *RpcStubManger) TcpSend(stub *RpcStub, request []byte) bool {
	if stub.netconn.GetConnState() != network.ConnStateConnected {
		return false
	}
	mgr.netcore.TcpSend(stub.netconn.GetSessionId(), request)
	return true
}

// 网络事件回调
func (mgr *RpcStubManger) OnAccept(c network.Connection) {
	peerHost, peerPort := c.GetPeerAddr()
	log.Debug("rpc stub manager OnAccept, peerHost:%v, peerPort:%v", peerHost, peerPort)
}

func (mgr *RpcStubManger) OnConnected(c network.Connection) {
	peerHost, peerPort := c.GetPeerAddr()
	log.Debug("rpc stub manager OnConnected, peerHost:%v, peerPort:%v", peerHost, peerPort)

	attrValue, ok := c.GetAttrib(NetConnAttrRpcStub)
	if ok && attrValue != nil {
		stub := attrValue.(*RpcStub)
		router, ok := mgr.consisRouterMap[stub.ServerType]
		if !ok {
			router = consistent.NewConsistent()
			mgr.consisRouterMap[stub.ServerType] = router
		}

		key := fmt.Sprintf("%d:%d", stub.ServerType, stub.ServerTypeInstID)
		router.Add(key)
	}
}

func (mgr *RpcStubManger) OnClosed(c network.Connection) {
	peerHost, peerPort := c.GetPeerAddr()
	log.Debug("rpc stub manager OnClosed, peerHost:%v, peerPort:%v", peerHost, peerPort)
	attrValue, ok := c.GetAttrib(NetConnAttrRpcStub)
	if ok && attrValue != nil {
		stub := attrValue.(*RpcStub)
		router, ok := mgr.consisRouterMap[stub.ServerType]
		if ok && router != nil {
			key := fmt.Sprintf("%d:%d", stub.ServerType, stub.ServerTypeInstID)
			router.Remove(key)
		}
	}
}
