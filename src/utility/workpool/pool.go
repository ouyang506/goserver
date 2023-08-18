package workpool

import "sync/atomic"

type TaskFunc func()

type Pool struct {
	size    int
	workers []*Worker

	i atomic.Int64
}

func NewPool(size int) *Pool {
	if size < 1 {
		size = 1
	}
	pool := &Pool{
		size: size,
	}

	for i := 0; i < size; i++ {
		pool.workers = append(pool.workers, newWorker())
	}
	return pool
}

func (pool *Pool) Submit(task func(), ops ...Option) {
	if task == nil {
		return
	}

	option := LoadOptions(ops...)

	var allocWorker *Worker = nil
	if option.workerHashKey != nil {
		allocWorker = pool.hashQueue(*option.workerHashKey)
	} else {
		i := pool.i.Add(1)
		//allocWorker = pool.minLoadQueue()
		allocWorker = pool.workers[i%int64(pool.size)]
	}

	allocWorker.push(task)
}

// 最小负载分配
func (pool *Pool) minLoadQueue() *Worker {
	index := 0
	minLoad := int64(-1)
	for i, workQueue := range pool.workers {
		size := workQueue.length()
		//已经没有任务在执行，直接返回
		if size == 0 {
			index = i
			break
		}

		if minLoad < 0 || size < minLoad {
			minLoad = size
			index = i
		}
	}
	return pool.workers[index]
}

// hash分配
func (pool *Pool) hashQueue(key uint64) *Worker {
	index := key % uint64(pool.size)
	return pool.workers[index]
}
