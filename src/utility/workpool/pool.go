package workpool

type Pool struct {
	size       int
	workQueues []*WorkerQueue
}

func NewPool(size int) *Pool {
	if size < 1 {
		size = 1
	}
	pool := &Pool{
		size: size,
	}

	for i := 0; i < size; i++ {
		pool.workQueues = append(pool.workQueues, newWorkQueue())
	}
	return pool
}

func (pool *Pool) Submit(task func(), ops ...Option) {
	if task == nil {
		return
	}

	option := LoadOptions(ops...)

	worker := newWorker(task)

	var allocQueue *WorkerQueue = nil
	if option.workerHashKey != nil {
		allocQueue = pool.hashQueue(*option.workerHashKey)
	} else {
		allocQueue = pool.minLoadQueue()
	}

	allocQueue.insert(worker)
}

// 最小负载分配
func (pool *Pool) minLoadQueue() *WorkerQueue {
	index := 0
	minLoad := -1
	for i, workQueue := range pool.workQueues {
		size := workQueue.length()
		if minLoad < 0 || size < minLoad {
			minLoad = size
			index = i
		}
	}
	return pool.workQueues[index]
}

// hash分配
func (pool *Pool) hashQueue(key uint64) *WorkerQueue {
	index := key % uint64(pool.size)
	return pool.workQueues[index]
}
