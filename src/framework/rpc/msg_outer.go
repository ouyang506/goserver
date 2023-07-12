package rpc

import (
	"encoding/binary"
	"errors"

	"framework/log"
	"framework/network"
	"framework/proto/pb"

	"google.golang.org/protobuf/proto"
)

// 客户端协议头
type OuterMessageHead struct {
	CallId int64 // keep reponse callid equals to request callid
	MsgID  int   // 大于0表示rpc请求消息，等于0表示rpc返回的消息
}

// 客户端协议
type OuterMessage struct {
	Head    OuterMessageHead
	PbMsg   proto.Message
	Content []byte
}

// 客户端协议解析器
// |CallId-8bytes|MsgID-4bytes|pbcontent-nbytes|
type OuterMessageCodec struct {
	rpcMgr *RpcManager
}

func NewOuterMessageCodec(rpcMgr *RpcManager) *OuterMessageCodec {
	return &OuterMessageCodec{
		rpcMgr: rpcMgr,
	}
}
func (cc *OuterMessageCodec) Encode(c network.Connection, in interface{}) (interface{}, bool, error) {
	outerMsg := in.(*OuterMessage)

	out := make([]byte, 12)
	skip := 0
	binary.LittleEndian.PutUint64(out[skip:], uint64(outerMsg.Head.CallId))
	skip += 8
	binary.LittleEndian.PutUint32(out[skip:], uint32(outerMsg.Head.MsgID))
	skip += 4

	content, _ := proto.Marshal(outerMsg.PbMsg)

	out = append(out, content...)

	return out, true, nil
}

func (cc *OuterMessageCodec) Decode(c network.Connection, in interface{}) (interface{}, bool, error) {
	outerMsgBytes := in.([]byte)
	if len(outerMsgBytes) < 12 {
		return nil, false, errors.New("unmarshal outer message msg id error")
	}

	skip := 0
	msg := &OuterMessage{}
	msg.Head.CallId = int64(binary.LittleEndian.Uint64(outerMsgBytes))
	skip += 8
	msg.Head.MsgID = int(binary.LittleEndian.Uint32(outerMsgBytes[skip:]))
	skip += 4
	msg.Content = outerMsgBytes[skip:]
	if msg.Head.MsgID == 0 {
		//rpc response
		rpcEntry := cc.rpcMgr.GetRpc(msg.Head.CallId)
		//respMsg := proto.Clone(rpcEntry.RespMsg)
		err := proto.Unmarshal(msg.Content, rpcEntry.RespMsg)
		if err != nil {
			//log error
		} else {
			msg.PbMsg = rpcEntry.RespMsg
		}
	} else if msg.Head.MsgID > 0 {
		reqMsg, _ := pb.GetProtoMsgById(msg.Head.MsgID)
		if reqMsg == nil {
			// log error
		}
		err := proto.Unmarshal(msg.Content, reqMsg)
		if err != nil {
			//log error
		} else {
			msg.PbMsg = reqMsg
		}
	} else {
		// message id error
	}

	return msg, true, nil
}

// 网络事件回调
type OuterNetEvent struct {
	rpcMgr *RpcManager
}

func NewOuterNetEvent(rpcMgr *RpcManager) *OuterNetEvent {
	return &OuterNetEvent{
		rpcMgr: rpcMgr,
	}
}

func (e *OuterNetEvent) OnAccept(c network.Connection) {
	peerHost, peerPort := c.GetPeerAddr()
	log.Info("NetEvent OnAccept, peerHost:%v, peerPort:%v", peerHost, peerPort)
}

func (e *OuterNetEvent) OnConnect(c network.Connection, err error) {
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

func (e *OuterNetEvent) OnClosed(c network.Connection) {
	peerHost, peerPort := c.GetPeerAddr()
	log.Info("NetEvent OnClosed, sessionId : %v, peerHost:%v, peerPort:%v", c.GetSessionId(), peerHost, peerPort)
}

func (e *OuterNetEvent) OnRcvMsg(c network.Connection, msg interface{}) {
	rcvOuterMsg := msg.(*OuterMessage)
	log.Debug("NetEvent OnRcvMsg, sessionId : %v, msg: %+v", c.GetSessionId(), rcvOuterMsg)

	if rcvOuterMsg.Head.MsgID == 0 {
		e.rpcMgr.OnRcvResponse(rcvOuterMsg.Head.CallId, rcvOuterMsg.PbMsg)
	}
}
