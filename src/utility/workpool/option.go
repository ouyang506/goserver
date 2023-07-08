package workpool

type Option func(ops *Options)

type Options struct {
	workerHashKey *uint64 //固定分配到指定workqueue
}

func LoadOptions(options ...Option) *Options {
	ops := &Options{}
	for _, option := range options {
		option(ops)
	}
	return ops
}

func WithWorkerHashKey(workerHashKey uint64) Option {
	return func(ops *Options) {
		ops.workerHashKey = new(uint64)
		*ops.workerHashKey = workerHashKey
	}
}
