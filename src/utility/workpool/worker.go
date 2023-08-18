package workpool

import (
	"container/list"
	"sync"
	"sync/atomic"
)

type Worker struct {
	loopFlag atomic.Bool

	queue *list.List
	len   atomic.Int64
	lock  sync.Locker
	cond  *sync.Cond
}

func newWorker() *Worker {
	l := NewSpinLock()
	wq := &Worker{
		queue: list.New(),
		lock:  l,
		cond:  sync.NewCond(l),
	}

	//wq.lazyloop()
	return wq
}

func (wq *Worker) push(task TaskFunc) {
	wq.lazyloop()

	wq.lock.Lock()
	defer wq.lock.Unlock()

	wq.queue.PushBack(task)
	wq.len.Add(1)
	wq.cond.Signal()
}

func (wq *Worker) waitPop() TaskFunc {
	wq.lock.Lock()
	defer wq.lock.Unlock()

	for wq.queue.Len() == 0 {
		wq.cond.Wait()
	}

	elem := wq.queue.Front()
	wq.queue.Remove(elem)
	wq.len.Add(-1)
	return elem.Value.(TaskFunc)
}

func (wq *Worker) length() int64 {
	return wq.len.Load()
}

func (wq *Worker) lazyloop() {
	if !wq.loopFlag.CompareAndSwap(false, true) {
		return
	}

	go func() {
		for {
			taskFunc := wq.waitPop()
			taskFunc()
		}
	}()
}
