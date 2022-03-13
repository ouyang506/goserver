package rpc

import (
	"encoding/binary"
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"

	"common/network"
	"common/pbmsg"
)

// 服务器内部协议头
type InnerMessageHead struct {
	CallId int64 // keep reponse callid equals to request callid
	MsgID  int
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
	switch msg.Head.MsgID {
	case int(pbmsg.MsgID_login_gate_req):
		pb = &pbmsg.LoginGateReqT{}
	case int(pbmsg.MsgID_login_gate_resp):
		pb = &pbmsg.LoginGateRespT{}
	default:
	}

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
