package rpc

import (
	"framework/log"
	"framework/network"
	"framework/proto/pb"
	"reflect"
)

// rpc网络事件回调
type RpcNetEventHandler interface {
	network.NetEventHandler
	SetOwner(*RpcManager)
	GetOwner() *RpcManager
}

type RpcNetEventHandlerBase struct {
	owner *RpcManager
}

func (h *RpcNetEventHandlerBase) SetOwner(*RpcManager) {

}

func (h *RpcNetEventHandlerBase) GetOwner() *RpcManager {
	return h.owner
}

func (h *RpcNetEventHandlerBase) OnAccept(c network.Connection) {
	peerHost, peerPort := c.GetPeerAddr()
	log.Info("NetEvent OnAccept, peerHost:%v, peerPort:%v", peerHost, peerPort)
}

func (h *RpcNetEventHandlerBase) OnConnect(c network.Connection, err error) {
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

func (h *RpcNetEventHandlerBase) OnClosed(c network.Connection) {
	peerHost, peerPort := c.GetPeerAddr()
	log.Info("NetEvent OnClosed, sessionId : %v, peerHost:%v, peerPort:%v", c.GetSessionId(), peerHost, peerPort)
}

func (e *RpcNetEventHandlerBase) OnRcvMsg(c network.Connection, msg interface{}) {
}

// 内部协议网络事件回调
type InnerNetEventHandler struct {
	RpcNetEventHandlerBase
}

func NewInnerNetEventHandler() *InnerNetEventHandler {
	return &InnerNetEventHandler{}
}

func (h *InnerNetEventHandler) OnRcvMsg(c network.Connection, msg interface{}) {
	rcvInnerMsg := msg.(*InnerMessage)
	msgId := rcvInnerMsg.Head.MsgID
	log.Debug("NetEvent OnRcvMsg, sessionId: %d, msgId: %d", c.GetSessionId(), msgId)
	switch {
	case rcvInnerMsg.Head.MsgID < 0:
		log.Error("receive wrong message id, sessionId: %d, msgId: %d", c.GetSessionId(), msgId)
		return
	case rcvInnerMsg.Head.MsgID == 0:
		h.GetOwner().OnRcvResponse(rcvInnerMsg.Head.CallId, rcvInnerMsg.PbMsg)
		return
	case rcvInnerMsg.Head.MsgID > 0:
		reqMsg := rcvInnerMsg.PbMsg
		_, respMsg := pb.GetProtoMsgById(msgId)
		method := h.GetOwner().GetMsgHandlerFunc(msgId)
		if method == nil {
			log.Error("rpc message handle function not found, sessionId : %d, msgId: %d ", c.GetSessionId(), msgId)
			return
		}

		method.Call([]reflect.Value{reflect.ValueOf(reqMsg), reflect.ValueOf(respMsg)})

		respInnerMsg := &InnerMessage{}
		respInnerMsg.Head.CallId = rcvInnerMsg.Head.CallId
		respInnerMsg.Head.MsgID = 0
		respInnerMsg.PbMsg = respMsg

		h.GetOwner().TcpSendMsg(c.GetSessionId(), respInnerMsg)
		return
	default:
		return
	}
}

// 外部协议网络事件回调
type OuterNetEventHandler struct {
	RpcNetEventHandlerBase
}

func NewOuterNetEventHandler() *OuterNetEventHandler {
	return &OuterNetEventHandler{}
}

func (h *OuterNetEventHandler) OnRcvMsg(c network.Connection, msg interface{}) {
	rcvOuterMsg := msg.(*OuterMessage)
	msgId := rcvOuterMsg.Head.MsgID
	log.Debug("NetEvent OnRcvMsg, sessionId: %d, msgId: %d", c.GetSessionId(), msgId)
	switch {
	case rcvOuterMsg.Head.MsgID < 0:
		log.Error("receive wrong message id, sessionId: %d, msgId: %d", c.GetSessionId(), msgId)
		return
	case rcvOuterMsg.Head.MsgID == 0:
		h.GetOwner().OnRcvResponse(rcvOuterMsg.Head.CallId, rcvOuterMsg.PbMsg)
		return
	case rcvOuterMsg.Head.MsgID > 0:
		reqMsg := rcvOuterMsg.PbMsg
		_, respMsg := pb.GetProtoMsgById(msgId)
		method := h.GetOwner().GetMsgHandlerFunc(msgId)
		if method == nil {
			log.Error("rpc message handle function not found, sessionId : %d, msgId: %d ", c.GetSessionId(), msgId)
			return
		}

		method.Call([]reflect.Value{reflect.ValueOf(reqMsg), reflect.ValueOf(respMsg)})

		respOuterMsg := &OuterMessage{}
		respOuterMsg.Head.CallId = rcvOuterMsg.Head.CallId
		respOuterMsg.Head.MsgID = 0
		respOuterMsg.PbMsg = respMsg

		h.GetOwner().TcpSendMsg(c.GetSessionId(), respOuterMsg)
		return
	default:
		return
	}
}
