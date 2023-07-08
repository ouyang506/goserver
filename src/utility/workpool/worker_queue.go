package workpool

import (
	"common/utility/queue"
	"sync"
)

type WorkerQueue struct {
	queue *queue.LockFreeQueue
	cond  *sync.Cond
}

func newWorkQueue() *WorkerQueue {
	wq := &WorkerQueue{
		queue: queue.NewLockFreeQueue(),
		cond:  sync.NewCond(NewSpinLock()),
		//cond: sync.NewCond(&sync.Mutex{}),
	}

	go wq.loop()
	return wq
}

func (wq *WorkerQueue) insert(work *Worker) {
	wq.queue.Enqueue(work)
	wq.cond.Signal()
}

func (wq *WorkerQueue) pop() *Worker {
	elem := wq.queue.Dequeue()
	if elem == nil {
		return nil
	}
	return elem.(*Worker)
}

func (wq *WorkerQueue) length() int {
	return int(wq.queue.Length())
}

func (wq *WorkerQueue) loop() {
	for {
		worker := wq.pop()
		if worker != nil {
			worker.run()
		} else {
			wq.cond.L.Lock()
			wq.cond.Wait()
			wq.cond.L.Unlock()
		}
	}
}
