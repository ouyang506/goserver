package random

import (
	"crypto/rand"
	"encoding/binary"
)

func RandUint32() uint32 {
	b := make([]byte, 4)
	rand.Read(b)
	value := binary.BigEndian.Uint32(b)
	return value
}

// @return 返回值范围[0-0x7FFFFFFF],不返回负数方便使用
func RandInt32() int32 {
	b := make([]byte, 4)
	rand.Read(b)
	b[0] = b[0] & 0x7f
	value := binary.BigEndian.Uint32(b)
	return int32(value)
}

func RandUint64() uint64 {
	b := make([]byte, 8)
	rand.Read(b)
	value := binary.BigEndian.Uint64(b)
	return value
}

func RandInt64() int64 {
	b := make([]byte, 8)
	rand.Read(b)
	b[0] = b[0] & 0x7f
	value := binary.BigEndian.Uint64(b)
	return int64(value)
}

// @return 前闭后开取值[min, max)
func RandRangeInt32(min, max int32) int32 {
	if min == max {
		return min
	}
	if min > max {
		min, max = max, min
	}
	return min + RandInt32()%(max-min)
}

// @return 前闭后开取值[min, max)
func RandRangeUint32(min, max uint32) uint32 {
	if min == max {
		return min
	}
	if min > max {
		min, max = max, min
	}
	return min + RandUint32()%(max-min)
}

// @return 前闭后开取值[min, max)
func RandRangeInt64(min, max int64) int64 {
	if min == max {
		return min
	}
	if min > max {
		min, max = max, min
	}
	return min + RandInt64()%(max-min)
}

// @return 前闭后开取值[min, max)
func RandRangeUint64(min, max uint64) uint64 {
	if min == max {
		return min
	}
	if min > max {
		min, max = max, min
	}
	return min + RandUint64()%(max-min)
}
