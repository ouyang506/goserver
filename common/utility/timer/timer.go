package timer

import (
	"math"
	"time"
)

const (
	// | 8bit | 8bit | 8bit | 8bit | 8bit | 8bit | 16bit |
	MaxLayer      = 7
	RootLayerBits = 16
	LayerBits     = 8

	DefaultTimeAccuracy = time.Millisecond * 10 // 10毫秒
)

type TimerCallBack func()

type Option struct {
	TimeAccuracy time.Duration
}

type Node struct {
	expire uint64
	cb     TimerCallBack
	l      *NodeList
}

type NodeList struct {
	nodes []*Node
}

func (nl *NodeList) addTail(node *Node) {
	node.l = nl
	nl.nodes = append(nl.nodes, node)
}

func (nl *NodeList) remove(node *Node) bool {
	if node.l != nl {
		return false
	}
	for i, n := range nl.nodes {
		if n == node {
			nl.nodes = append(nl.nodes[0:i], nl.nodes[i+1:]...)
			return true
		}
	}

	return false
}

func (nl *NodeList) forEach(f func(*Node)) {
	for _, node := range nl.nodes {
		f(node)
	}
}

type WheelTimer struct {
	stopChannel  chan bool
	timeAccuracy time.Duration

	layerMasks    []uint64
	layerMaxValue []uint64
	layerShift    []int

	jiffies  uint64        // since wheel timer start
	allNodes [][]*NodeList // all timer nodes
}

func NewWheelTimer(option *Option) *WheelTimer {
	t := &WheelTimer{}

	if option != nil && option.TimeAccuracy > 0 {
		t.timeAccuracy = option.TimeAccuracy
	} else {
		t.timeAccuracy = DefaultTimeAccuracy
	}

	t.layerMasks = make([]uint64, MaxLayer)
	t.layerMaxValue = make([]uint64, MaxLayer)
	t.layerShift = make([]int, MaxLayer)

	t.layerMasks[0] = 1<<RootLayerBits - 1
	t.layerMaxValue[0] = 1<<RootLayerBits - 1
	t.layerShift[0] = 0
	for i := 1; i < MaxLayer; i++ {
		t.layerMasks[i] = 1<<LayerBits - 1
		t.layerMaxValue[i] = 1<<(RootLayerBits+i*LayerBits) - 1
		t.layerShift[i] = RootLayerBits + (i-1)*LayerBits
	}

	for i := 0; i < MaxLayer; i++ {
		layerNodes := make([]*NodeList, t.layerMasks[i]+1)
		for j := 0; j <= int(t.layerMasks[i]); j++ {
			layerNodes[j] = &NodeList{}
		}
		t.allNodes = append(t.allNodes, layerNodes)
	}

	return t
}

func (t *WheelTimer) AddTimer(duration time.Duration, cb TimerCallBack) *Node {
	delta := uint64(duration / t.timeAccuracy)
	if delta < 1 {
		// put it to the next root slot if expired now
		delta = 1
	}

	if math.MaxUint64-delta < t.jiffies {
		// actually a invalid expired time
		delta = math.MaxUint64 - t.jiffies
	}

	expire := t.jiffies + delta
	node := &Node{
		expire: expire,
		cb:     cb,
	}

	t.addNode(node)

	return node
}

func (t *WheelTimer) RemoveTimer(n *Node) bool {
	nodeList := n.l
	if nodeList == nil {
		return false
	}
	return nodeList.remove(n)
}

func (t *WheelTimer) addNode(node *Node) {
	delta := node.expire - t.jiffies
	for i := 0; i < MaxLayer; i++ {
		if delta < t.layerMaxValue[i] {
			idx := (node.expire >> t.layerShift[i]) & t.layerMasks[i]
			nodeList := t.allNodes[i][idx]
			nodeList.addTail(node)
			break
		}
	}
}

func (t *WheelTimer) cascade(layer int) {
	idx := (t.jiffies >> t.layerShift[layer]) & t.layerMasks[layer]
	nodeList := t.allNodes[layer][idx]
	t.allNodes[layer][idx] = &NodeList{}

	nodeList.forEach(func(node *Node) {
		t.addNode(node)
	})

}

func (t *WheelTimer) tick(delta int) {
	for i := 0; i < delta; i++ {
		t.jiffies++

		rootIdx := t.jiffies & t.layerMasks[0]
		if rootIdx == 0 {
			for layer := 1; layer < MaxLayer; layer++ {
				if (t.jiffies>>t.layerShift[layer])&t.layerMasks[layer] == 0 {
					t.cascade(layer + 1)
				} else {
					break
				}
			}
		}

		expireList := t.allNodes[0][rootIdx]
		expireList.forEach(func(n *Node) {
			n.cb()
		})
	}

}

func (t *WheelTimer) Run() {
	lastTime := time.Now().UnixNano() / int64(t.timeAccuracy)
	ticker := time.NewTicker(t.timeAccuracy)
	for {
		select {
		case <-ticker.C:
			{
				currentTime := time.Now().UnixNano() / int64(t.timeAccuracy)
				delta := currentTime - lastTime
				if delta > 0 {
					t.tick(int(delta))
				}
				lastTime = currentTime
			}
		case <-t.stopChannel:
			return
		}
	}
}
