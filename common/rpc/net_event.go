package rpc

import (
	"common/log"
	"common/network"
	"reflect"
)

// 网络事件回调
type RpcNetEvent struct {
	rpcMgr *RpcManager
}

func (e *RpcNetEvent) OnAccept(c network.Connection) {
	peerHost, peerPort := c.GetPeerAddr()
	log.Info("NetEvent OnAccept, peerHost:%v, peerPort:%v", peerHost, peerPort)
}

func (e *RpcNetEvent) OnConnect(c network.Connection, err error) {
	peerHost, peerPort := c.GetPeerAddr()
	if err != nil {
		log.Info("rpc stub manager OnConnectFailed, sessionId: %v, peerHost:%v, peerPort:%v", c.GetSessionId(), peerHost, peerPort)
	} else {
		log.Info("rpc stub manager OnConnected, sessionId: %v, peerHost:%v, peerPort:%v,", c.GetSessionId(), peerHost, peerPort)
		stub, ok := c.GetAttrib(AttrRpcStub)
		if !ok || stub == nil {
			return
		}
		stub.(*RpcStub).onConnected()
	}
}

func (e *RpcNetEvent) OnClosed(c network.Connection) {
	peerHost, peerPort := c.GetPeerAddr()
	log.Info("NetEvent OnClosed, sessionId : %v, peerHost:%v, peerPort:%v", c.GetSessionId(), peerHost, peerPort)
}

func (e *RpcNetEvent) OnRcvMsg(c network.Connection, msg interface{}) {
	rcvInnerMsg := msg.(*InnerMessage)
	log.Debug("NetEvent OnRcvMsg, sessionId : %v, msg: %+v", c.GetSessionId(), rcvInnerMsg)

	if rcvInnerMsg.Head.MsgID == 0 {

		e.rpcMgr.OnRcvResponse(rcvInnerMsg.Head.CallId, rcvInnerMsg)

	} else {

		handler := GetProtoMsgHandler(rcvInnerMsg.Head.MsgID)
		// if handler.Kind() != reflect.Func {
		// 	log.Error("get handler error, msg_id : %v", rcvInnerMsg.Head.MsgID)
		// 	return
		// }
		resp := GetProtoMsgById(rcvInnerMsg.Head.MsgID)
		if resp == nil {
			log.Error("get resp msg error, msg_id : %v", rcvInnerMsg.Head.MsgID)
			return
		}
		in := []reflect.Value{reflect.ValueOf(rcvInnerMsg.PbMsg), reflect.ValueOf(resp)}
		handler.Call(in)

		respInnerMsg := &InnerMessage{}
		respInnerMsg.Head.CallId = rcvInnerMsg.Head.CallId
		respInnerMsg.Head.MsgID = 0
		respInnerMsg.PbMsg = resp

		e.rpcMgr.rpcStubMgr.netcore.TcpSendMsg(c.GetSessionId(), respInnerMsg)
	}
}
