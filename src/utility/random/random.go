package random

import (
	"crypto/rand"
	"encoding/binary"
)

func Uint32() uint32 {
	b := make([]byte, 4)
	rand.Read(b)
	value := binary.BigEndian.Uint32(b)
	return value
}

// @return 返回值范围[0-0x7FFFFFFF],不返回负数方便使用
func Int32() int32 {
	b := make([]byte, 4)
	rand.Read(b)
	b[0] = b[0] & 0x7f
	value := binary.BigEndian.Uint32(b)
	return int32(value)
}

func Uint64() uint64 {
	b := make([]byte, 8)
	rand.Read(b)
	value := binary.BigEndian.Uint64(b)
	return value
}

func Int64() int64 {
	b := make([]byte, 8)
	rand.Read(b)
	b[0] = b[0] & 0x7f
	value := binary.BigEndian.Uint64(b)
	return int64(value)
}

// @return 前闭后开取值[min, max)
func RangeInt32(min, max int32) int32 {
	if min == max {
		return min
	}
	if min > max {
		min, max = max, min
	}
	return min + Int32()%(max-min)
}

// @return 前闭后开取值[min, max)
func RangeUint32(min, max uint32) uint32 {
	if min == max {
		return min
	}
	if min > max {
		min, max = max, min
	}
	return min + Uint32()%(max-min)
}

// @return 前闭后开取值[min, max)
func RangeInt64(min, max int64) int64 {
	if min == max {
		return min
	}
	if min > max {
		min, max = max, min
	}
	return min + Int64()%(max-min)
}

// @return 前闭后开取值[min, max)
func RangeUint64(min, max uint64) uint64 {
	if min == max {
		return min
	}
	if min > max {
		min, max = max, min
	}
	return min + Uint64()%(max-min)
}
