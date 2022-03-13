package rpc

import (
	"common/log"
	"common/network"
	"common/pbmsg"
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

	codecs := []network.Codec{}
	codecs = append(codecs, NewInnerMessageCodec())
	codecs = append(codecs, network.NewVariableFrameLenCodec())

	mgr.netcore = network.NewNetworkCore(
		network.WithEventHandler(mgr),
		network.WithLoadBalance(network.NewLoadBalanceRoundRobin(0)),
		network.WithSocketSendBufferSize(10240),
		network.WithSocketRcvBufferSize(10240),
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
	// 按照instanceID排序
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
		stub.(*RpcStub).trySendRpc()
	}
}

func (mgr *RpcStubManger) OnClosed(c network.Connection) {
	peerHost, peerPort := c.GetPeerAddr()
	log.Info("rpc stub manager OnClosed, sessionId : %v, peerHost:%v, peerPort:%v", c.GetSessionId(), peerHost, peerPort)
}

func (mgr *RpcStubManger) OnRcvMsg(c network.Connection, msg interface{}) {
	rcvInnerMsg := msg.(*InnerMessage)
	log.Debug("rpc stub manager OnRcvMsg, sessionId : %v, msg: %+v", c.GetSessionId(), rcvInnerMsg)
	mgr.rpcMgr.OnRcvResponse(rcvInnerMsg.Head.CallId, rcvInnerMsg)

	// for test response
	if rcvInnerMsg.Head.MsgID == int(pbmsg.MsgID_login_gate_req) {
		respInnerMsg := &InnerMessage{}
		respInnerMsg.Head.MsgID = int(pbmsg.MsgID_login_gate_resp)
		pb := &pbmsg.LoginGateRespT{}
		respInnerMsg.PbMsg = pb
		pb.Result = new(int32)
		*pb.Result = 8888
		mgr.netcore.TcpSendMsg(c.GetSessionId(), respInnerMsg)
	}
}
