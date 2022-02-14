package consistent

import (
	"fmt"
	"hash/crc32"
	"sort"
	"sync"
)

const DefaultNumberOfReplicas = 20

type Uint32Slice []uint32

func (x Uint32Slice) Len() int           { return len(x) }
func (x Uint32Slice) Less(i, j int) bool { return x[i] < x[j] }
func (x Uint32Slice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

type Consistent struct {
	sync.RWMutex
	numberOfReplicas uint32
	circle           map[uint32]string
	members          map[string]bool
	sortedHash       Uint32Slice
}

func NewConsistent() *Consistent {
	c := &Consistent{}
	c.numberOfReplicas = DefaultNumberOfReplicas
	c.circle = map[uint32]string{}
	c.members = map[string]bool{}
	c.sortedHash = []uint32{}
	return c
}

func (c *Consistent) Add(member string) {
	c.Lock()
	c.add(member)
	c.Unlock()
}

func (c *Consistent) Remove(member string) {
	c.Lock()
	c.remove(member)
	c.Unlock()
}

func (c *Consistent) Get(name string) string {
	c.RLock()
	member := c.get(name)
	c.RUnlock()
	return member
}

func (c *Consistent) Members() []string {
	c.RLock()
	arr := []string{}
	for k := range c.members {
		arr = append(arr, k)
	}
	c.RUnlock()
	return arr
}

func (c *Consistent) add(member string) {
	for i := uint32(0); i < c.numberOfReplicas; i++ {
		replKey := c.replMemberKey(member, i)
		c.circle[c.hash(replKey)] = member
	}
	c.members[member] = true
	c.updateSortedHash()
}

func (c *Consistent) remove(member string) {
	for i := uint32(0); i < c.numberOfReplicas; i++ {
		replKey := c.replMemberKey(member, i)
		delete(c.circle, c.hash(replKey))
	}
	delete(c.members, member)
	c.updateSortedHash()
}

func (c *Consistent) get(name string) string {
	if len(c.circle) <= 0 {
		return ""
	}

	hashValue := c.hash(name)
	index := sort.Search(len(c.sortedHash), func(x int) bool {
		return c.sortedHash[x] >= hashValue
	})
	if index >= len(c.sortedHash) {
		index = 0
	}
	memberKey := c.sortedHash[index]
	return c.circle[memberKey]
}

func (c *Consistent) updateSortedHash() {
	arr := c.sortedHash[:0]
	for k := range c.circle {
		arr = append(arr, k)
	}

	sort.Sort(arr)
	c.sortedHash = arr
}

func (c *Consistent) hash(key string) uint32 {
	return crc32.ChecksumIEEE([]byte(key))
}

func (c *Consistent) replMemberKey(member string, index uint32) string {
	return fmt.Sprintf("%s:%d", member, index)
}
