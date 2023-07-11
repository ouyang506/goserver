package rpc

import (
	"encoding/binary"
	"errors"
	"fmt"
	"framework/log"
	"framework/network"

	"google.golang.org/protobuf/proto"
)

// 服务器内部协议头
type InnerMessageHead struct {
	CallId int64 // keep reponse callid equals to request callid
	MsgID  int   // 大于0表示rpc请求消息，等于0表示rpc返回的消息
	Guid   int64 // 玩家id等
}

// 服务器内部协议
type InnerMessage struct {
	Head  InnerMessageHead
	PbMsg proto.Message
}

// 服务器内部解析器
// |CallId-8bytes|MsgID-4bytes|pbcontent-nbytes|
type InnerMessageCodec struct {
}

func NewInnerMessageCodec() *InnerMessageCodec {
	return &InnerMessageCodec{}
}
func (cc *InnerMessageCodec) Encode(c network.Connection, in interface{}) (interface{}, bool, error) {
	innerMsg := in.(*InnerMessage)

	out := make([]byte, 12)
	skip := 0
	binary.LittleEndian.PutUint64(out[skip:], uint64(innerMsg.Head.CallId))
	skip += 8
	binary.LittleEndian.PutUint32(out[skip:], uint32(innerMsg.Head.MsgID))
	skip += 4

	content, err := proto.Marshal(innerMsg.PbMsg)
	if err != nil {
		return nil, false, err
	}
	out = append(out, content...)

	return out, true, nil
}

func (cc *InnerMessageCodec) Decode(c network.Connection, in interface{}) (interface{}, bool, error) {
	innerMsgBytes := in.([]byte)
	if len(innerMsgBytes) < 12 {
		return nil, false, errors.New("unmarshal inner message msg id error")
	}

	skip := 0
	msg := &InnerMessage{}
	msg.Head.CallId = int64(binary.LittleEndian.Uint64(innerMsgBytes))
	skip += 8
	msg.Head.MsgID = int(binary.LittleEndian.Uint32(innerMsgBytes[skip:]))
	skip += 4

	var pb proto.Message = nil
	// switch msg.Head.MsgID {
	// case int(pbmsg.MsgID_login_gate_req):
	// 	pb = &pbmsg.LoginGateReqT{}
	// case int(pbmsg.MsgID_login_gate_resp):
	// 	pb = &pbmsg.LoginGateRespT{}
	// default:
	// }

	if pb == nil {
		return nil, false, fmt.Errorf("can not find pb message by id : %v", msg.Head.MsgID)
	}

	err := proto.Unmarshal(innerMsgBytes[skip:], pb)
	if err != nil {
		return nil, false, err
	}
	msg.PbMsg = pb
	return msg, true, nil
}

// 网络事件回调
type InnerNetEvent struct {
	rpcMgr *RpcManager
}

func NewInnerNetEvent(rpcMgr *RpcManager) *InnerNetEvent {
	return &InnerNetEvent{
		rpcMgr: rpcMgr,
	}
}

func (e *InnerNetEvent) OnAccept(c network.Connection) {
	peerHost, peerPort := c.GetPeerAddr()
	log.Info("NetEvent OnAccept, peerHost:%v, peerPort:%v", peerHost, peerPort)
}

func (e *InnerNetEvent) OnConnect(c network.Connection, err error) {
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

func (e *InnerNetEvent) OnClosed(c network.Connection) {
	peerHost, peerPort := c.GetPeerAddr()
	log.Info("NetEvent OnClosed, sessionId : %v, peerHost:%v, peerPort:%v", c.GetSessionId(), peerHost, peerPort)
}

func (e *InnerNetEvent) OnRcvMsg(c network.Connection, msg interface{}) {
	rcvInnerMsg := msg.(*InnerMessage)
	log.Debug("NetEvent OnRcvMsg, sessionId : %v, msg: %+v", c.GetSessionId(), rcvInnerMsg)

	if rcvInnerMsg.Head.MsgID == 0 {

		e.rpcMgr.OnRcvResponse(rcvInnerMsg.Head.CallId, rcvInnerMsg)

	} //else {

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
	//}
}
