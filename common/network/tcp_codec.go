package network

import (
	"errors"

	"goserver/common/network/gnet"
)

const (
	max_frame_len = 100000
)

type TcpCodec struct {
}

// Encode encodes frames upon server responses into TCP stream.
func (cc *TcpCodec) Encode(c gnet.Conn, buf []byte) (out []byte, err error) {
	length := len(buf)
	if length <= 0 {
		return nil, errors.New("encode tcp frame length error")
	}

	if length > max_frame_len {
		return nil, errors.New("encode tcp frame length greater than max_frame_len")
	}

	out = writeLength(length)

	out = append(out, buf...)
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
