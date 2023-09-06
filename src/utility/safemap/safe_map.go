package safemap

import "sync"

type SafeMap[K comparable, V any] struct {
	l sync.RWMutex
	m map[K]V
}

func NewSafeMap[K comparable, V any]() *SafeMap[K, V] {
	sm := &SafeMap[K, V]{
		m: map[K]V{},
	}
	return sm
}

func (sm *SafeMap[K, V]) Get(k K) (V, bool) {
	sm.l.RLock()
	defer sm.l.RUnlock()
	v, ok := sm.m[k]
	return v, ok
}

func (sm *SafeMap[K, V]) Add(k K, v V) {
	sm.l.Lock()
	defer sm.l.Unlock()
	sm.m[k] = v
}

func (sm *SafeMap[K, V]) Del(k K) {
	sm.l.Lock()
	defer sm.l.Unlock()
	delete(sm.m, k)
}

func (sm *SafeMap[K, V]) Length() int {
	sm.l.RLock()
	defer sm.l.RUnlock()
	return len(sm.m)
}
