package lru

import "container/list"

type Callback func(interface{}) bool

type LRU struct {
	capacity int
	elemMap  map[interface{}]*list.Element
	elemList list.List
}

func NewLRU(capacity int) *LRU {
	if capacity < 1 {
		capacity = 1
	}
	l := &LRU{
		capacity: capacity,
		elemMap:  make(map[interface{}]*list.Element, capacity),
		elemList: *list.New(),
	}
	return l
}

func (l *LRU) Add(k interface{}, v interface{}) bool {
	if entry, ok := l.elemMap[k]; ok {
		l.elemList.MoveToFront(entry)
		entry.Value = v
		return false
	}

	elem := l.elemList.PushFront(v)
	l.elemMap[k] = elem

	if l.elemList.Len() > l.capacity {
		l.elemList.Remove(l.elemList.Back())
	}
	return true
}

func (l *LRU) Get(k interface{}) interface{} {
	if entry, ok := l.elemMap[k]; ok {
		l.elemList.MoveToFront(entry)
		return entry.Value
	}
	return nil
}

func (l *LRU) Peek(k interface{}) interface{} {
	if entry, ok := l.elemMap[k]; ok {
		return entry.Value
	}
	return nil
}

func (l *LRU) Contains(k interface{}) bool {
	_, ok := l.elemMap[k]
	return ok
}

func (l *LRU) Remove(k interface{}) bool {
	if entry, ok := l.elemMap[k]; ok {
		l.elemList.Remove(entry)
		delete(l.elemMap, k)
		return true
	}
	return false
}

func (l *LRU) Foreach(f Callback) {
	for elem := l.elemList.Front(); elem != nil; elem = elem.Next() {
		if !f(elem.Value) {
			break
		}
	}
}
