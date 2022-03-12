package timer

import (
	"math"
	"sync"
	"time"
)

const (
	// | 8bit | 8bit | 8bit | 8bit | 8bit | 8bit | 16bit |
	MaxLayer      = 7
	RootLayerBits = 16
	LayerBits     = 8

	DefaultTimeAccuracy = time.Millisecond * 10 // 10毫秒
)

type Option struct {
	TimeAccuracy time.Duration
}

type TimerCallBack func()

type TimerNode struct {
	expire uint64
	cb     TimerCallBack
	l      *TimerNodeList
	pre    *TimerNode
	next   *TimerNode
}

type TimerNodeList struct {
	root *TimerNode
}

func newTimerNodeList() *TimerNodeList {
	l := &TimerNodeList{}
	l.root = &TimerNode{}
	l.root.next = l.root
	l.root.pre = l.root
	return l
}

func (l *TimerNodeList) addTail(node *TimerNode) {
	node.l = l
	at := l.root.pre

	node.pre = at
	node.next = at.next
	node.pre.next = node
	node.next.pre = node
}

func (l *TimerNodeList) remove(node *TimerNode) bool {
	if node.l != l {
		return false
	}

	node.pre.next = node.next
	node.next.pre = node.pre
	node.next = nil // avoid memory leaks
	node.pre = nil  // avoid memory leaks
	node.l = nil

	return true
}

func (l *TimerNodeList) clear() {
	l.root.next = l.root
	l.root.pre = l.root
}

func (nl *TimerNodeList) forEach(f func(*TimerNode)) {
	if nl.root.next == nl.root {
		return
	}

	for n := nl.root.next; n != nl.root; n = n.next {
		f(n)
	}
}

type TimerWheel struct {
	timeAccuracy time.Duration

	mutex sync.Mutex

	layerMasks    []uint64
	layerMaxValue []uint64
	layerShift    []int

	jiffies      uint64             // since timer wheel start running
	allNodes     [][]*TimerNodeList // all timer nodes
	lastTickTime int64

	expiredList []*TimerNode

	stopChan chan bool
}

func NewTimerWheel(option *Option) *TimerWheel {
	tw := &TimerWheel{}

	if option != nil && option.TimeAccuracy > 0 {
		tw.timeAccuracy = option.TimeAccuracy
	} else {
		tw.timeAccuracy = DefaultTimeAccuracy
	}

	tw.layerMasks = make([]uint64, MaxLayer)
	tw.layerMaxValue = make([]uint64, MaxLayer)
	tw.layerShift = make([]int, MaxLayer)

	tw.layerMasks[0] = 1<<RootLayerBits - 1
	tw.layerMaxValue[0] = 1<<RootLayerBits - 1
	tw.layerShift[0] = 0
	for i := 1; i < MaxLayer; i++ {
		tw.layerMasks[i] = 1<<LayerBits - 1
		tw.layerMaxValue[i] = 1<<(RootLayerBits+i*LayerBits) - 1
		tw.layerShift[i] = RootLayerBits + (i-1)*LayerBits
	}

	for i := 0; i < MaxLayer; i++ {
		layerNodes := make([]*TimerNodeList, tw.layerMasks[i]+1)
		for j := 0; j <= int(tw.layerMasks[i]); j++ {
			layerNodes[j] = newTimerNodeList()
		}
		tw.allNodes = append(tw.allNodes, layerNodes)
	}

	tw.lastTickTime = time.Now().UnixNano() / int64(tw.timeAccuracy)

	tw.expiredList = []*TimerNode{}

	tw.stopChan = make(chan bool)
	return tw
}

func (tw *TimerWheel) Start() {
	go func() {
		tick := time.NewTicker(tw.timeAccuracy)
		defer tick.Stop()
		for {
			select {
			case <-tick.C:
				{
					tw.Tick()
				}
			case <-tw.stopChan:
				{
					return
				}
			}
		}
	}()
}

func (tw *TimerWheel) Stop() {
	select {
	case tw.stopChan <- true:
	default:
	}
}

// tick by wall clock
func (tw *TimerWheel) Tick() {
	currentTime := time.Now().UnixNano() / int64(tw.timeAccuracy)
	delta := currentTime - tw.lastTickTime
	if delta > 0 {
		tw.doTick(int(delta))
	}
	tw.lastTickTime = currentTime
}

// user called tick by elapsed time
func (tw *TimerWheel) TickElapsed(elaspeTime time.Duration) {
	delta := int64(elaspeTime) / int64(tw.timeAccuracy)
	if delta > 0 {
		tw.doTick(int(delta))
	}
}

func (tw *TimerWheel) AddTimer(duration time.Duration, cb TimerCallBack) *TimerNode {
	delta := uint64(duration / tw.timeAccuracy)
	if delta < 1 {
		// put it to the next root slot if expired now
		delta = 1
	}

	tw.mutex.Lock()

	if math.MaxUint64-delta < tw.jiffies {
		// actually a invalid expired time
		delta = math.MaxUint64 - tw.jiffies
	}

	expire := tw.jiffies + delta
	node := &TimerNode{
		expire: expire,
		cb:     cb,
	}

	tw.addNode(node)

	tw.mutex.Unlock()
	return node
}

func (tw *TimerWheel) RemoveTimer(n *TimerNode) (ret bool) {
	tw.mutex.Lock()

	if n.l != nil {
		ret = n.l.remove(n)
	}

	tw.mutex.Unlock()
	return
}

func (tw *TimerWheel) After(duration time.Duration) (chan bool, *TimerNode) {
	ch := make(chan bool)
	node := tw.AddTimer(duration, func() {
		select {
		case ch <- true:
		default:
		}
	})
	return ch, node
}

func (tw *TimerWheel) addNode(node *TimerNode) {
	delta := node.expire - tw.jiffies
	for i := 0; i < MaxLayer; i++ {
		if delta < tw.layerMaxValue[i] {
			idx := (node.expire >> tw.layerShift[i]) & tw.layerMasks[i]
			nodeList := tw.allNodes[i][idx]
			nodeList.addTail(node)
			break
		}
	}
}

func (tw *TimerWheel) cascade(layer int) {
	idx := (tw.jiffies >> tw.layerShift[layer]) & tw.layerMasks[layer]
	nodeList := tw.allNodes[layer][idx]
	tw.allNodes[layer][idx].clear()

	nodeList.forEach(func(node *TimerNode) {
		tw.addNode(node)
	})

}

func (tw *TimerWheel) doTick(delta int) {
	for i := 0; i < delta; i++ {
		tw.mutex.Lock()
		tw.jiffies++

		rootIdx := tw.jiffies & tw.layerMasks[0]
		if rootIdx == 0 {
			for layer := 1; layer < MaxLayer; layer++ {
				if (tw.jiffies>>tw.layerShift[layer])&tw.layerMasks[layer] == 0 {
					tw.cascade(layer + 1)
				} else {
					break
				}
			}
		}

		expireList := tw.allNodes[0][rootIdx]
		tw.expiredList = tw.expiredList[:0]
		expireList.forEach(func(n *TimerNode) {
			tw.expiredList = append(tw.expiredList, n)
		})

		tw.mutex.Unlock()

		for _, n := range tw.expiredList {
			n.cb()
		}
	}
}
