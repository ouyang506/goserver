package network

import (
	"common/utility/ringbuffer"
	"errors"
)

const (
	max_frame_len = 2 * 1024 * 1024 // 2M
)

type TcpCodec struct {
}

// Encode encodes frames upon server responses into TCP stream.
func (cc *TcpCodec) Encode(buf []byte, ring *ringbuffer.RingBuffer) (err error) {
	length := len(buf)
	if length <= 0 {
		return errors.New("encode tcp frame length error")
	}

	if length > max_frame_len {
		return errors.New("encode tcp frame length greater than max_frame_len")
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

// Decode decodes frames from TCP stream via specific implementation.
func (cc *TcpCodec) Decode(c gnet.Conn) ([]byte, error) {

	in := c.Read()
	if len(in) <= 0 {
		return nil, errors.New("decode tcp frame length error")
	}

	frameLen, shift, succ := readLength(in)
	if !succ {
		return nil, nil
	}

	if frameLen <= 0 {
		return nil, errors.New("decode tcp frame parse length error")
	}

	if len(in) < shift+frameLen {
		return nil, nil
	}

	fullMessage := make([]byte, frameLen)
	copy(fullMessage, in[shift:shift+frameLen])
	c.ShiftN(shift + frameLen)
	return fullMessage, nil
}

func writeLength(length int) (out []byte) {
	out = make([]byte, 0, 6)
	for {
		b := byte(length & 0x7F)
		length = length >> 7

		if length > 0 {
			out = append(out, byte(0x80)|b)
		} else {
			out = append(out, b)
			break
		}
	}
	return
}

func readLength(buf []byte) (out int, shift int, succ bool) {
	out = 0
	shift = 0
	succ = false
	for i, x := range buf {
		shift++
		out = (int(x)&0x7F)<<(7*i) | out

		if out > max_frame_len {
			succ = false
			return
		}

		if x&0x80 <= 0 {
			succ = false
			break
		}
	}

	return
}
