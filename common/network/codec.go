package network

import (
	"common/utility/ringbuffer"
	"errors"
)

const (
	MaxFrameLength     = 2 * 1024 * 1024 // 2M
	MaxFrameLengthSize = 4               // beyond 2M
)

type Codec interface {
	Encode([]byte, *ringbuffer.RingBuffer) error
	Decode(*ringbuffer.RingBuffer) ([]byte, error)
}

type FixedFrameLenCodec struct {
}

type VariableFrameLenCodec struct {
}

func (cc *VariableFrameLenCodec) Encode(buf []byte, ring *ringbuffer.RingBuffer) (err error) {
	length := len(buf)
	if length <= 0 {
		return errors.New("encode tcp frame length error")
	}

	if length > MaxFrameLength {
		return errors.New("encode tcp frame length greater than max frame length")
	}

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
	return
}

func (cc *VariableFrameLenCodec) Decode(ring *ringbuffer.RingBuffer) (msg []byte, err error) {
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
		return nil, nil
	}

	frameLengthSize := index + 1
	if frameLengthSize+frameLen > totalLen {
		return nil, nil
	}

	if frameLengthSize+frameLen <= headLen {
		msg = head[frameLengthSize : frameLengthSize+frameLen]
	} else {
		if frameLengthSize >= headLen {
			msg = tail[frameLengthSize-headLen : frameLengthSize-headLen+frameLen]
		} else {
			msg = make([]byte, frameLen)
			copy(msg, head[frameLengthSize:])
			copy(msg, tail[:frameLen+frameLengthSize-headLen])
		}
	}
	ring.Discard(frameLengthSize + frameLen)
	return
}
