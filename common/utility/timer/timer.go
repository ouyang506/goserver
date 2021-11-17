package main

import (
	"sync/atomic"
	"time"
)

const (
	DefaultTimeUnit = time.Millisecond * 10 // 10毫秒
	// | 8bit | 8bit | 8bit | 8bit | 8bit | 8bit | 16bit |
	Layer         = 7
	RootLayerBits = 16
	LayerBits     = 8
)

type TimerId uint64
type TimerCallBack func()

type Node struct {
	id TimerId
	cb TimerCallBack
}

type WheelTimer struct {
	autoIncId     int64
	layerSlotCnts []int
	layerIndex    []int
	jiffies       int64
	nodes         [][]*Node
}

func NewWheelTimer() *WheelTimer {
	wt := &WheelTimer{}

	wt.layerSlotCnts = make([]int, Layer)
	wt.layerSlotCnts[0] = 1 << RootLayerBits
	for i := 1; i <= Layer; i++ {
		wt.layerSlotCnts[i] = 1 << LayerBits
	}

	wt.layerIndex = make([]int, Layer)
	return wt
}

func (t *WheelTimer) AddTimer(duration time.Duration, cb TimerCallBack) TimerId {
	id := atomic.AddInt64(&t.autoIncId, 1)
	node := &Node{
		id: TimerId(id),
		cb: cb,
	}

	expire := t.jiffies + int64(duration)/int64(DefaultTimeUnit)

	if expire-t.jiffies < int64(t.layerSlotCnts[0]) {
		idx := (expire - t.jiffies + int64(t.layerIndex[0])) % int64(t.layerSlotCnts[0])
		t.nodes[idx] = append(t.nodes[idx], node)
	} else {
		for {

		}
	}

	return 0
}

func main() {

}
