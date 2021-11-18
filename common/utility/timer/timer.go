package timer

import (
	"time"
)

const (
	DefaultTimeUnit = time.Millisecond * 10 // 10毫秒

	// | 8bit | 8bit | 8bit | 8bit | 8bit | 8bit | 16bit |
	Layer         = 7
	RootLayerBits = 16
	LayerBits     = 8
)

type TimerCallBack func()

type Node struct {
	expire int64
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

func (nl *NodeList) clear() {
	nl.nodes = nl.nodes[:0]
}

func (nl *NodeList) forEach(f func(*Node)) {
	for _, node := range nl.nodes {
		f(node)
	}
}

type WheelTimer struct {
	stopChannel chan int

	layerSlotCnts []int
	layerMasks    []int
	layerShift    []int

	jiffies    int64 // since wheel timer start
	layerIndex []int

	allNodes [][]*NodeList // all timer nodes
}

func NewWheelTimer() *WheelTimer {
	t := &WheelTimer{}

	t.layerSlotCnts = make([]int, Layer)
	t.layerMasks = make([]int, Layer)
	t.layerShift = make([]int, Layer)

	t.layerSlotCnts[0] = 1 << RootLayerBits
	t.layerMasks[0] = t.layerSlotCnts[0] - 1
	t.layerShift[0] = 0
	for i := 1; i <= Layer; i++ {
		t.layerSlotCnts[i] = 1 << LayerBits
		t.layerMasks[i] = (t.layerSlotCnts[i] - 1) << (RootLayerBits + i*LayerBits)
		t.layerShift[i] = RootLayerBits + i*LayerBits
	}

	t.layerIndex = make([]int, Layer)
	return t
}

func (t *WheelTimer) AddTimer(duration time.Duration, cb TimerCallBack) *Node {

	expire := t.jiffies + int64(duration)/int64(DefaultTimeUnit)

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
	expire := node.expire
	delta := expire - t.jiffies
	for i := Layer - 1; i >= 0; i-- {
		mask := int64(t.layerMasks[i])
		if delta&mask > 0 {
			idx := (delta & mask) >> t.layerShift[i]
			nodeList := t.allNodes[i][idx]
			nodeList.addTail(node)
			break
		}
	}
}

func (t *WheelTimer) cascade(layer int) {
	nodeList := t.allNodes[layer][0]
	nodeList.forEach(func(node *Node) {
		t.addNode(node)
	})
	nodeList.clear()
}

func (t *WheelTimer) tick(delta int) {
	for i := 0; i < delta; i++ {
		t.jiffies++
		for l := 0; l < Layer-1; l++ {
			if t.jiffies&int64(t.layerMasks[l]) == 0 {
				t.cascade(l + 1)
			}
		}

		rootIdx := t.jiffies & int64(t.layerMasks[0])
		expireList := t.allNodes[0][rootIdx]
		expireList.forEach(func(n *Node) {
			n.cb()
		})
	}

}

func (t *WheelTimer) Run() {
	lastTime := time.Now().UnixNano()
	ticker := time.NewTicker(time.Duration(DefaultTimeUnit))
	for {
		select {
		case <-ticker.C:
			{
				currentTime := time.Now().UnixNano()
				delta := (currentTime - lastTime) / int64(DefaultTimeUnit)
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
