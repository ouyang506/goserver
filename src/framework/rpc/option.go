package rpc

import (
	"time"
)

type Option func(ops *Options)

type Options struct {
	RpcTimout       *time.Duration     // rpc超时时间
	RpcRouteType    *RpcRouteType      // 路由规则
	NetEventHandler RpcNetEventHandler // 自定义网络事件处理
}

func LoadOptions(options ...Option) *Options {
	ops := &Options{}
	for _, option := range options {
		option(ops)
	}
	return ops
}

func WithTimeout(timeout time.Duration) Option {
	return func(ops *Options) {
		ops.RpcTimout = &timeout
	}
}

func WithRouteType(routeType RpcRouteType) Option {
	return func(ops *Options) {
		ops.RpcRouteType = &routeType
	}
}

func WithNetEventHandler(h RpcNetEventHandler) Option {
	return func(ops *Options) {
		ops.NetEventHandler = h
	}
}
