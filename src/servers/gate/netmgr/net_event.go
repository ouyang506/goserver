package netmgr

import (
	"framework/log"
	"framework/network"
	"gate/handler"
	"reflect"

	"google.golang.org/protobuf/proto"
)

var (
	ConnAttrPlayer = new(struct{})
)

type NetEvent struct {
	netmgr *NetMgr
}

func NewNetEvent(netmgr *NetMgr) *NetEvent {
	return &NetEvent{
		netmgr: netmgr,
	}
}

func (e *NetEvent) OnAccept(c network.Connection) {
	peerHost, peerPort := c.GetPeerAddr()
	log.Info("NetEvent OnAccept, peerHost:%v, peerPort:%v", peerHost, peerPort)
}

func (e *NetEvent) OnConnect(c network.Connection, err error) {
	peerHost, peerPort := c.GetPeerAddr()
	if err != nil {
		log.Info("NetEvent OnConnectFailed, sessionId: %v, peerHost:%v, peerPort:%v", c.GetSessionId(), peerHost, peerPort)
	} else {
		log.Info("NetEvent OnConnected, sessionId: %v, peerHost:%v, peerPort:%v,", c.GetSessionId(), peerHost, peerPort)
	}
	c.SetAttrib(&ConnAttrPlayer, 0)
}

func (e *NetEvent) OnClosed(c network.Connection) {
	peerHost, peerPort := c.GetPeerAddr()
	log.Info("NetEvent OnClosed, sessionId : %v, peerHost:%v, peerPort:%v", c.GetSessionId(), peerHost, peerPort)
	c.SetAttrib(&ConnAttrPlayer, 0)
}

func (e *NetEvent) OnRcvMsg(c network.Connection, msg interface{}) {
	log.Debug("NetEvent OnRcvMsg, sessionId : %v, msg: %+v", c.GetSessionId(), msg)

	rcvOuterMsg := msg.(*OuterMessage)

	msgId := rcvOuterMsg.Head.MsgID
	if msgId >= 1000 && msgId <= 1999 {
		//gate handler

		handler := handler.MessageHandler{}
		method, reqValue, respValue := handler.GetMsgHandlerById(msgId)
		if method == nil {
			log.Error("gate handle message func not found, msg id: %d", msgId)
			return
		}

		req := reqValue.Interface().(proto.Message)
		err := proto.Unmarshal(rcvOuterMsg.Content, req)
		if err != nil {
			log.Error("decode outer message error: %v, msg id: %v", err, msgId)
			return
		}

		in := []reflect.Value{reqValue, respValue}
		method.Call(in)

		// handler := GetProtoMsgHandler(rcvOuterMsg.Head.MsgID)

		// resp := GetProtoMsgById(rcvInnerMsg.Head.MsgID)
		// if resp == nil {
		// 	log.Error("get resp msg error, msg_id : %v", rcvInnerMsg.Head.MsgID)
		// 	return
		// }
		// in := []reflect.Value{reflect.ValueOf(rcvInnerMsg.PbMsg), reflect.ValueOf(resp)}
		// handler.Call(in)

		respOuterMsg := &OuterMessage{}
		respOuterMsg.Head.CallId = rcvOuterMsg.Head.CallId
		respOuterMsg.Head.MsgID = 0
		respContent, _ := proto.Marshal(respValue.Interface().(proto.Message))
		respOuterMsg.Content = respContent

		e.netmgr.netcore.TcpSendMsg(c.GetSessionId(), respOuterMsg)

	} else {
		// route to backend
	}
}
