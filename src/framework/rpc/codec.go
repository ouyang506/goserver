package rpc

import (
	"encoding/binary"
	"errors"
	"fmt"
	"framework/log"
	"framework/network"
	"framework/proto/pb"

	"google.golang.org/protobuf/proto"
)

// 服务器内部协议
type InnerMessage struct {
	CallId  int64 // keep reponse call id equals to request call id
	MsgID   int   // 大于0表示rpc请求消息，等于0表示rpc返回的消息
	Guid    int64 // 玩家id等
	Content proto.Message
}

// 服务器内部协议解析器
type InnerMessageCodec struct {
	rpcMgr *RpcManager
}

func NewInnerMessageCodec(rpcMgr *RpcManager) *InnerMessageCodec {
	return &InnerMessageCodec{rpcMgr: rpcMgr}
}

func (cc *InnerMessageCodec) Encode(c network.Connection, in interface{}) (interface{}, bool, error) {
	innerMsg := in.(*InnerMessage)

	out := make([]byte, 12)
	skip := 0
	binary.LittleEndian.PutUint64(out[skip:], uint64(innerMsg.CallId))
	skip += 8
	binary.LittleEndian.PutUint32(out[skip:], uint32(innerMsg.MsgID))
	skip += 4

	if innerMsg.Content != nil {
		contentBytes, _ := proto.Marshal(innerMsg.Content)
		out = append(out, contentBytes...)
	}

	return out, true, nil
}

func (cc *InnerMessageCodec) Decode(c network.Connection, in interface{}) (interface{}, bool, error) {
	innerMsgBytes := in.([]byte)
	if len(innerMsgBytes) < 12 {
		return nil, false, errors.New("inner message length error")
	}

	skip := 0
	innerMsg := &InnerMessage{}
	innerMsg.CallId = int64(binary.LittleEndian.Uint64(innerMsgBytes))
	skip += 8
	innerMsg.MsgID = int(binary.LittleEndian.Uint32(innerMsgBytes[skip:]))
	skip += 4

	content := innerMsgBytes[skip:]
	switch {
	case innerMsg.MsgID == 0: //rpc response
		rpcEntry := cc.rpcMgr.GetRpc(innerMsg.CallId)
		if rpcEntry == nil {
			// rpc has been timeout
			log.Error("rpc has been removed, callId = %v", innerMsg.CallId)
			return nil, false, nil
		}
		err := proto.Unmarshal(content, rpcEntry.RespMsg)
		if err != nil {
			// ignore the invalid response
			log.Error("unmarshal rpc response error: %v, callId: %v, reqMsgId: %v",
				err, innerMsg.CallId, rpcEntry.MsgId)
			return nil, false, nil
		}
		innerMsg.Content = rpcEntry.RespMsg

	case innerMsg.MsgID > 0: //rpc request
		reqMsg, _ := pb.GetProtoMsgById(innerMsg.MsgID)
		if reqMsg == nil {
			return nil, false, fmt.Errorf("invalid message id %v", innerMsg.MsgID)
		}

		err := proto.Unmarshal(content, reqMsg)
		if err != nil {
			return nil, false, fmt.Errorf("unmarshal message error: %v, msgId: %v", err, innerMsg.MsgID)
		}
		innerMsg.Content = reqMsg
	default:
		return nil, false, fmt.Errorf("invalid message id %v", innerMsg.MsgID)
	}

	return innerMsg, true, nil
}

// 客户端协议
type OuterMessage struct {
	CallId  int64 // keep reponse call id equals to request call id
	MsgID   int   // 大于0表示rpc请求消息，等于0表示rpc返回的消息
	Content proto.Message
}

// 客户端协议解析器
type OuterMessageCodec struct {
	rpcMgr *RpcManager
}

func NewOuterMessageCodec(rpcMgr *RpcManager) *OuterMessageCodec {
	return &OuterMessageCodec{rpcMgr: rpcMgr}
}
func (cc *OuterMessageCodec) Encode(c network.Connection, in interface{}) (interface{}, bool, error) {
	outerMsg := in.(*OuterMessage)

	out := make([]byte, 12)
	skip := 0
	binary.LittleEndian.PutUint64(out[skip:], uint64(outerMsg.CallId))
	skip += 8
	binary.LittleEndian.PutUint32(out[skip:], uint32(outerMsg.MsgID))
	skip += 4

	if outerMsg.Content != nil {
		contentBytes, _ := proto.Marshal(outerMsg.Content)
		out = append(out, contentBytes...)
	}

	return out, true, nil
}

func (cc *OuterMessageCodec) Decode(c network.Connection, in interface{}) (interface{}, bool, error) {
	outerMsgBytes := in.([]byte)
	if len(outerMsgBytes) < 12 {
		return nil, false, errors.New("outer message length error")
	}

	skip := 0
	msg := &OuterMessage{}
	msg.CallId = int64(binary.LittleEndian.Uint64(outerMsgBytes))
	skip += 8
	msg.MsgID = int(binary.LittleEndian.Uint32(outerMsgBytes[skip:]))
	skip += 4

	content := outerMsgBytes[skip:]
	switch {
	case msg.MsgID == 0: //rpc response
		rpcEntry := cc.rpcMgr.GetRpc(msg.CallId)
		if rpcEntry != nil {
			// rpc has been timeout
			return nil, false, nil
		}
		err := proto.Unmarshal(content, rpcEntry.RespMsg)
		if err != nil {
			// ignore the invalid response
			log.Error("unmarshal rpc response error: %v, callId: %v, reqMsgId: %v",
				err, msg.CallId, rpcEntry.MsgId)
			return nil, false, nil
		}
		msg.Content = rpcEntry.RespMsg

	case msg.MsgID > 0: //rpc request
		reqMsg, _ := pb.GetProtoMsgById(msg.MsgID)
		if reqMsg == nil {
			return nil, false, fmt.Errorf("invalid message id %v", msg.MsgID)
		}

		err := proto.Unmarshal(content, reqMsg)
		if err != nil {
			return nil, false, fmt.Errorf("unmarshal message error: %v, msgId: %v", err, msg.MsgID)
		}
		msg.Content = reqMsg
	default:
		return nil, false, fmt.Errorf("invalid message id %v", msg.MsgID)
	}

	return msg, true, nil
}
