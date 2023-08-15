package workpool

import (
	"container/list"
	"sync"
	"sync/atomic"
)

type WorkerQueue struct {
	loopFlag atomic.Bool

	queue *list.List
	len   atomic.Int64
	lock  sync.Locker
	cond  *sync.Cond
}

func newWorkQueue() *WorkerQueue {
	l := NewSpinLock()
	//l := &sync.Mutex{}
	wq := &WorkerQueue{
		queue: list.New(),
		lock:  l,
		cond:  sync.NewCond(l),
	}
	return wq
}

func (wq *WorkerQueue) insert(work *Worker) {
	wq.lazyloop()

	wq.lock.Lock()
	defer wq.lock.Unlock()

	wq.queue.PushBack(work)
	wq.len.Add(1)
	wq.cond.Signal()
}

func (wq *WorkerQueue) length() int64 {
	return wq.len.Load()
}

func (wq *WorkerQueue) pop() *Worker {
	wq.lock.Lock()
	defer wq.lock.Unlock()

	for wq.queue.Len() == 0 {
		wq.cond.Wait()
	}

	elem := wq.queue.Front()
	wq.queue.Remove(elem)
	wq.len.Add(-1)
	return elem.Value.(*Worker)
}

func (wq *WorkerQueue) lazyloop() {
	if !wq.loopFlag.CompareAndSwap(false, true) {
		return
	}

	go func() {
		for {
			worker := wq.pop()
			worker.run()
		}
	}()
}
