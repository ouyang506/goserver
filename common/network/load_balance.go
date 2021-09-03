package network

import (
	"sync"
)

type LoadBalance interface {
	AllocConnection(int64)
	GetConnection(int64)
	RemoveConnection(int64)
}

type LoadBalanceRoundRobin struct {
	bucketCount int
	bucketIndex int
	bucketMap   map[int64]int
	bucketMutex sync.RWMutex
}

func NewLoadBalanceRoundRobin(bucketCount int) {
	lb := &LoadBalanceRoundRobin{}
	lb.bucketCount = bucketCount
	lb.bucketIndex = 0
	lb.bucketMap = make(map[int64]int)
	lb.bucketMutex = sync.RWMutex{}
}

func (lb *LoadBalanceRoundRobin) AllocConnection(sessionId int64) int {
	lb.bucketMutex.Lock()
	allocIndex := lb.bucketIndex
	lb.bucketMap[sessionId] = allocIndex
	lb.bucketIndex = (lb.bucketIndex + 1) % lb.bucketCount
	lb.bucketMutex.Unlock()
	return allocIndex
}

func (lb *LoadBalanceRoundRobin) GetConnection(sessionId int64) int {
	lb.bucketMutex.RLock()
	idx, ok := lb.bucketMap[sessionId]
	lb.bucketMutex.RUnlock()

	if !ok {
		return -1
	} else {
		return idx
	}
}

func (lb *LoadBalanceRoundRobin) RemoveConnection(sessionId int64) {
	lb.bucketMutex.Lock()
	delete(lb.bucketMap, sessionId)
	lb.bucketMutex.Unlock()
}
