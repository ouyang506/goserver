package rpc

import (
	"time"
)

type Option func(ops *Options)

type Options struct {
	RpcTimout       time.Duration      // rpc超时时间
	RouteKey        string             // rpc路由key
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
		ops.RpcTimout = timeout
	}
}

func WithRouteKey(routeKey string) Option {
	return func(ops *Options) {
		ops.RouteKey = routeKey
	}
}

func WithNetEventHandler(h RpcNetEventHandler) Option {
	return func(ops *Options) {
		ops.NetEventHandler = h
	}
}
