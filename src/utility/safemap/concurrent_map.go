package safemap

type HashFunc[K comparable] func(k K) int

type ConcurrentMap[K comparable, V any] struct {
	shardNum int
	shards   []*SafeMap[K, V]
	hashFunc HashFunc[K]
}

func NewConcurrentMap[K comparable, V any](shardNum int, hashFunc HashFunc[K]) *ConcurrentMap[K, V] {
	if shardNum <= 1 {
		shardNum = 1
	}
	m := &ConcurrentMap[K, V]{
		shardNum: shardNum,
		hashFunc: hashFunc,
	}

	m.shards = make([]*SafeMap[K, V], shardNum)
	for i := 0; i < shardNum; i++ {
		m.shards[i] = NewSafeMap[K, V]()
	}

	return m
}

func (cm *ConcurrentMap[K, V]) getShard(k K) int {
	hashValue := cm.hashFunc(k)
	if hashValue < 0 {
		hashValue = -hashValue
	}
	return hashValue % cm.shardNum
}

func (cm *ConcurrentMap[K, V]) Get(k K) (V, bool) {
	index := cm.getShard(k)
	return cm.shards[index].Get(k)
}

func (cm *ConcurrentMap[K, V]) Add(k K, v V) {
	index := cm.getShard(k)
	cm.shards[index].Add(k, v)
}

func (cm *ConcurrentMap[K, V]) Del(k K) {
	index := cm.getShard(k)
	cm.shards[index].Del(k)
}
