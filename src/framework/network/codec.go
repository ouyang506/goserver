package network

import (
	"errors"
)

const (
	MaxFrameLength     = 2 * 1024 * 1024 // 2M
	MaxFrameLengthSize = 4               // beyond 2M
)

// codec会放在网络协程中处理，提升性能
// codec队列由有序数组构成，第一个codec的输出作为第二个codec的输入
// encode从codec_1 -> codec_2 .. -> codec_N, decode倒序从codec_N -> .. codec_2 -> codec_1
// encode队列最终将msg->conn.sendbuff，decode最终将conn.rcvbuff->msg
// bChain返回false时停止该链式过程
type Codec interface {
	Encode(c Connection, in interface{}) (out interface{}, bChain bool, err error)
	Decode(c Connection, in interface{}) (out interface{}, bChain bool, err error)
}

// MessageFrameLen固定长度解析（4字节)
type FixedFrameLenCodec struct {
}

// MessageFrameLen可变长度解析（类utf8编码）
type VariableFrameLenCodec struct {
}

func NewVariableFrameLenCodec() *VariableFrameLenCodec {
	return &VariableFrameLenCodec{}
}

func (cc *VariableFrameLenCodec) Encode(c Connection, in interface{}) (ignore_out interface{}, bChain bool, err error) {
	buf := in.([]byte)

	length := len(buf)
	if length <= 0 {
		err = errors.New("encode tcp frame length error")
		return
	}

	if length > MaxFrameLength {
		err = errors.New("encode tcp frame length greater than max frame length")
		return
	}
	ring := c.GetSendBuff()
	// TODO : need check full ?
	for {
		b := byte(length & 0x7F)
		length = length >> 7
		if length > 0 {
			ring.WriteByte(byte(0x80) | b)
		} else {
			ring.WriteByte(b)
			break
		}
	}
	ring.Write(buf)
	bChain = true
	return
}

func (cc *VariableFrameLenCodec) Decode(c Connection, in interface{}) (out interface{}, bChain bool, err error) {
	ring := c.GetRcvBuff()

	head, tail := ring.PeekAll()
	headLen := len(head)
	tailLen := len(tail)
	totalLen := headLen + tailLen

	frameLen := 0
	b := byte(0)
	index := 0
	for {
		if index >= totalLen {
			return
		}

		if index+1 > MaxFrameLengthSize {
			err = errors.New("decode length beyond max length size")
			return
		}

		if index < headLen {
			b = head[index]
		} else {
			b = tail[index-headLen]
		}
		frameLen = (int(b)&0x7F)<<(7*index) | frameLen

		if frameLen > MaxFrameLength {
			err = errors.New("decode length beyond max length")
			return
		}

		if b&0x80 <= 0 {
			break
		}
		index++
	}

	if frameLen <= 0 {
		err = nil
		bChain = false
		return
	}

	frameLengthSize := index + 1
	if frameLengthSize+frameLen > totalLen {
		err = nil
		bChain = false
		return
	}

	var content []byte = nil
	if frameLengthSize+frameLen <= headLen {
		content = head[frameLengthSize : frameLengthSize+frameLen]
	} else {
		if frameLengthSize >= headLen {
			content = tail[frameLengthSize-headLen : frameLengthSize-headLen+frameLen]
		} else {
			content = make([]byte, frameLen)
			copy(content, head[frameLengthSize:])
			copy(content[headLen-frameLengthSize:], tail[:frameLen+frameLengthSize-headLen])
		}
	}
	ring.Discard(frameLengthSize + frameLen)
	out = content
	bChain = true
	return
}
