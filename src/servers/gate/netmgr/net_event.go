package netmgr

import (
	"framework/log"
	"framework/network"
	"framework/proto/pb"
	"framework/rpc"
	"reflect"
)

// 网络事件回调
type NetMessageEvent struct {
	rpcMgr *rpc.RpcManager
}

func NewNetMessageEvent() *NetMessageEvent {
	return &NetMessageEvent{}
}
func (e *NetMessageEvent) SetRpcMgr(rpcMgr *rpc.RpcManager) {
	e.rpcMgr = rpcMgr
}

func (e *NetMessageEvent) OnAccept(c network.Connection) {
	//
}

func (e *NetMessageEvent) OnConnect(c network.Connection, err error) {
	//
}

func (e *NetMessageEvent) OnClosed(c network.Connection) {
	//
}

func (e *NetMessageEvent) OnRcvMsg(c network.Connection, msg interface{}) {
	rcvOuterMsg := msg.(*rpc.OuterMessage)
	log.Debug("NetEvent OnRcvMsg, sessionId : %v, msg: %+v", c.GetSessionId(), rcvOuterMsg)

	msgId := rcvOuterMsg.Head.MsgID
	if msgId > 0 {

		log.Debug("handle message, id: %d", msgId)
		reqMsg := rcvOuterMsg.PbMsg
		_, respMsg := pb.GetProtoMsgById(msgId)
		method := e.rpcMgr.GetMsgHandlerFunc(msgId)
		if method == nil {
			//log error
			return
		}

		method.Call([]reflect.Value{reflect.ValueOf(reqMsg), reflect.ValueOf(respMsg)})

		respOuterMsg := &rpc.OuterMessage{}
		respOuterMsg.Head.CallId = rcvOuterMsg.Head.CallId
		respOuterMsg.Head.MsgID = 0
		respOuterMsg.PbMsg = respMsg

		e.rpcMgr.TcpSendMsg(c.GetSessionId(), respOuterMsg)

	}
}
