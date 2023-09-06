package safemap

type HashFunc[K comparable] func(k K) uint32

type ConcurrentMap[K comparable, V any] struct {
	shardNum uint32
	shards   []*SafeMap[K, V]
	hashFunc HashFunc[K]
}

func NewConcurrentMap[K comparable, V any](shardNum uint32, hashFunc HashFunc[K]) *ConcurrentMap[K, V] {
	if shardNum <= 1 {
		shardNum = 1
	}
	m := &ConcurrentMap[K, V]{
		shardNum: shardNum,
		hashFunc: hashFunc,
	}

	m.shards = make([]*SafeMap[K, V], shardNum)
	for i := uint32(0); i < shardNum; i++ {
		m.shards[i] = NewSafeMap[K, V]()
	}

	return m
}

func (cm *ConcurrentMap[K, V]) getShard(k K) int {
	hashValue := cm.hashFunc(k)

	if hashValue < cm.shardNum {
		return int(hashValue)
	}
	return int(hashValue % cm.shardNum)
}

func (cm *ConcurrentMap[K, V]) Get(k K) (V, bool) {
	index := cm.getShard(k)
	return cm.shards[index].Get(k)
}

func (cm *ConcurrentMap[K, V]) Set(k K, v V) {
	index := cm.getShard(k)
	cm.shards[index].Set(k, v)
}

func (cm *ConcurrentMap[K, V]) Del(k K) {
	index := cm.getShard(k)
	cm.shards[index].Del(k)
}
