package network

import "common/log"

type Option func(ops *Options)

type Options struct {
	logger               log.Logger
	numLoops             int
	loadBalance          LoadBalance
	eventHandler         NetEventHandler
	socketSendBufferSize int
	socketRcvBufferSize  int
	socketTcpNoDelay     bool
	codec                Codec
}

func loadOptions(op []Option) *Options {
	ops := &Options{}
	for _, f := range op {
		f(ops)
	}
	return ops
}

func WithLogger(logger log.Logger) Option {
	return func(ops *Options) {
		ops.logger = logger
	}
}

func WithNumLoop(numberLoops int) Option {
	return func(ops *Options) {
		ops.numLoops = numberLoops
	}
}

func WithLoadBalance(loadBalance LoadBalance) Option {
	return func(ops *Options) {
		ops.loadBalance = loadBalance
	}
}

func WithEventHandler(eventHandler NetEventHandler) Option {
	return func(ops *Options) {
		ops.eventHandler = eventHandler
	}
}

func WithSocketSendBufferSize(sendBufSize int) Option {
	return func(ops *Options) {
		ops.socketSendBufferSize = sendBufSize
	}
}

func WithSocketRcvBufferSize(rcvBufSize int) Option {
	return func(ops *Options) {
		ops.socketRcvBufferSize = rcvBufSize
	}
}

func WithSocketTcpNoDelay(tcpNoDelay bool) Option {
	return func(ops *Options) {
		ops.socketTcpNoDelay = tcpNoDelay
	}
}

func WithFrameCodec(codec Codec) Option {
	return func(ops *Options) {
		ops.codec = codec
	}
}
