package netmgr

import (
	"encoding/binary"
	"errors"

	"framework/network"
)

// 客户端协议头
type OuterMessageHead struct {
	CallId int64 // keep reponse callid equals to request callid
	MsgID  int   // 大于0表示rpc请求消息，等于0表示rpc返回的消息
}

// 客户端协议
type OuterMessage struct {
	Head    OuterMessageHead
	Content []byte
}

// 客户端协议解析器
// |CallId-8bytes|MsgID-4bytes|pbcontent-nbytes|
type OuterMessageCodec struct {
}

func NewOuterMessageCodec() *OuterMessageCodec {
	return &OuterMessageCodec{}
}
func (cc *OuterMessageCodec) Encode(c network.Connection, in interface{}) (interface{}, bool, error) {
	outerMsg := in.(*OuterMessage)

	out := make([]byte, 12)
	skip := 0
	binary.LittleEndian.PutUint64(out[skip:], uint64(outerMsg.Head.CallId))
	skip += 8
	binary.LittleEndian.PutUint32(out[skip:], uint32(outerMsg.Head.MsgID))
	skip += 4

	out = append(out, outerMsg.Content...)

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

	return msg, true, nil
}
