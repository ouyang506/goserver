package rpc

import (
	"framework/log"
	"framework/network"
	"framework/proto/pb"
	"reflect"
	"runtime"
	"utility/workpool"
)

// rpc网络事件回调
type RpcNetEventHandler interface {
	network.NetEventHandler
	SetOwner(*RpcManager)
	GetOwner() *RpcManager
}

type RpcNetEventHandlerBase struct {
	owner          *RpcManager
	processMsgPool *workpool.Pool
}

func (h *RpcNetEventHandlerBase) SetOwner(owner *RpcManager) {
	h.owner = owner
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
		log.Error("NetEvent connect remote endpoint error, error: %v, sessionId: %v, peerHost:%v, peerPort:%v", err, c.GetSessionId(), peerHost, peerPort)
	} else {
		log.Info("NetEvent connect remote endpoint success, sessionId: %v, peerHost:%v, peerPort:%v,", c.GetSessionId(), peerHost, peerPort)
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
	return &InnerNetEventHandler{
		RpcNetEventHandlerBase: RpcNetEventHandlerBase{
			processMsgPool: workpool.NewPool(1024),
		},
	}
}

func (h *InnerNetEventHandler) OnRcvMsg(c network.Connection, msg interface{}) {
	rcvInnerMsg := msg.(*InnerMessage)
	msgId := rcvInnerMsg.MsgID
	// log.Debug("NetEvent OnRcvMsg, sessionId: %d, callId : %v, msgId: %d",
	// 	c.GetSessionId(), rcvInnerMsg.CallId, msgId)

	switch {
	case msgId < 0:
		log.Error("receive wrong message id, sessionId: %d, msgId: %d", c.GetSessionId(), msgId)
		return
	case msgId == 0:
		h.GetOwner().OnRcvResponse(rcvInnerMsg.CallId, rcvInnerMsg.Content)
		return
	case msgId > 0:
		reqMsg := rcvInnerMsg.Content
		_, respMsg := pb.GetProtoMsgById(msgId)
		method := h.GetOwner().GetMsgHandlerFunc(msgId)
		if method == nil {
			log.Error("rpc message handle function not found, sessionId : %d, msgId: %d ", c.GetSessionId(), msgId)
			return
		}

		var ops []workpool.Option
		if rcvInnerMsg.Guid > 0 {
			ops = append(ops, workpool.WithWorkerHashKey(uint64(rcvInnerMsg.Guid)))
		}
		h.processMsgPool.Submit(func() {
			defer func() {
				if r := recover(); r != nil {
					buff := make([]byte, 4096)
					n := runtime.Stack(buff, false)
					log.Error("handler rcv msg panic, msgId: %v, stack : %s", msgId, string(buff[:n]))
				}
			}()
			ctx := createContext(h.owner.Context()).SetNetConn(c)
			if respMsg == nil {
				method.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(reqMsg)})
			} else {
				method.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(reqMsg), reflect.ValueOf(respMsg)})

				respInnerMsg := &InnerMessage{}
				respInnerMsg.CallId = rcvInnerMsg.CallId
				respInnerMsg.MsgID = 0
				respInnerMsg.Guid = rcvInnerMsg.Guid
				respInnerMsg.Content = respMsg

				h.GetOwner().TcpSendMsg(c.GetSessionId(), respInnerMsg)
			}
		}, ops...)

	default:
		return
	}
}

// 外部协议网络事件回调
type OuterNetEventHandler struct {
	RpcNetEventHandlerBase
}

func NewOuterNetEventHandler() *OuterNetEventHandler {
	return &OuterNetEventHandler{
		RpcNetEventHandlerBase: RpcNetEventHandlerBase{
			processMsgPool: workpool.NewPool(1024),
		},
	}
}

func (h *OuterNetEventHandler) OnRcvMsg(c network.Connection, msg interface{}) {
	rcvOuterMsg := msg.(*OuterMessage)
	msgId := rcvOuterMsg.MsgID
	log.Debug("NetEvent OnRcvMsg, sessionId: %d, msgId: %d", c.GetSessionId(), msgId)
	switch {
	case rcvOuterMsg.MsgID < 0:
		log.Error("receive wrong message id, sessionId: %d, msgId: %d", c.GetSessionId(), msgId)
		return
	case msgId == 0:
		h.GetOwner().OnRcvResponse(rcvOuterMsg.CallId, rcvOuterMsg.Content)
		return
	case rcvOuterMsg.MsgID > 0:
		reqMsg := rcvOuterMsg.Content
		_, respMsg := pb.GetProtoMsgById(msgId)
		method := h.GetOwner().GetMsgHandlerFunc(msgId)
		if method == nil {
			log.Error("rpc message handle function not found, sessionId : %d, msgId: %d ", c.GetSessionId(), msgId)
			return
		}

		h.processMsgPool.Submit(func() {
			defer func() {
				if r := recover(); r != nil {
					buff := make([]byte, 4096)
					n := runtime.Stack(buff, false)
					log.Error("handler rcv msg panic, msgId: %v, stack : %s", msgId, string(buff[:n]))
				}
			}()

			ctx := createContext(h.owner.Context()).SetNetConn(c)
			if respMsg == nil {
				method.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(reqMsg)})
			} else {
				method.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(reqMsg), reflect.ValueOf(respMsg)})

				respOuterMsg := &OuterMessage{}
				respOuterMsg.CallId = rcvOuterMsg.CallId
				respOuterMsg.MsgID = 0
				respOuterMsg.Content = respMsg

				h.GetOwner().TcpSendMsg(c.GetSessionId(), respOuterMsg)
				return
			}
		}, workpool.WithWorkerHashKey(uint64(c.GetSessionId())))

	default:
		return
	}
}
