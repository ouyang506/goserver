package rpc

import (
	"encoding/binary"
	"errors"
	"framework/network"
	"framework/proto/pb"

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
	Head    InnerMessageHead
	PbMsg   proto.Message
	Content []byte
}

// 服务器内部解析器
// |CallId-8bytes|MsgID-4bytes|pbcontent-nbytes|
type InnerMessageCodec struct {
	rpcMgr *RpcManager
}

func NewInnerMessageCodec(rpcMgr *RpcManager) *InnerMessageCodec {
	return &InnerMessageCodec{
		rpcMgr: rpcMgr,
	}
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
	msg.Content = innerMsgBytes[skip:]
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
