package netmgr

import (
	"framework/log"
	"framework/network"
	"framework/proto/pb"
	"framework/rpc"

	"google.golang.org/protobuf/proto"
)

// 网络事件回调
type NetMessageEvent struct {
}

func NewNetMessageEvent() *NetMessageEvent {
	return &NetMessageEvent{}
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
		arr, ok := pb.MsgId2Type[msgId]
		if !ok {
			// drop the illegal request message
			return
		}
		// req := reflect.New(reflect.TypeOf(arr[0]).Elem())
		// resp := reflect.New(reflect.TypeOf(arr[1]).Elem())
		req := arr[0].Interface()
		proto.Unmarshal(rcvOuterMsg.Content, req)

		// handler := GetProtoMsgHandler(rcvInnerMsg.Head.MsgID)
		// // if handler.Kind() != reflect.Func {
		// // 	log.Error("get handler error, msg_id : %v", rcvInnerMsg.Head.MsgID)
		// // 	return
		// // }
		// resp := GetProtoMsgById(rcvInnerMsg.Head.MsgID)
		// if resp == nil {
		// 	log.Error("get resp msg error, msg_id : %v", rcvInnerMsg.Head.MsgID)
		// 	return
		// }
		// in := []reflect.Value{reflect.ValueOf(rcvInnerMsg.PbMsg), reflect.ValueOf(resp)}
		// handler.Call(in)

		// respInnerMsg := &InnerMessage{}
		// respInnerMsg.Head.CallId = rcvInnerMsg.Head.CallId
		// respInnerMsg.Head.MsgID = 0
		// respInnerMsg.PbMsg = resp

		// e.rpcMgr.rpcStubMgr.netcore.TcpSendMsg(c.GetSessionId(), respInnerMsg)

	}
}
