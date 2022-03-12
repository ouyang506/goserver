package rpc

import (
	"sync/atomic"
	"time"
)

var (
	rpcMgr *RpcManager = nil
)

var (
	DefaultRpcTimeout = time.Second * 3
	nextRpcCallId     = int64(0)
)

func genNextRpcCallId() int64 {
	return atomic.AddInt64(&nextRpcCallId, 1)
}

// 使用rpc前需要先初始化rpc管理器
func InitRpc(mgr *RpcManager) {
	rpcMgr = mgr
}

type Rpc struct {
	CallId        int64         // rpc请求唯一ID
	Timeout       time.Duration // 超时时间
	TargetSvrType int           // 目标服务类型
	RouteKey      string        // 路由key
	Request       []byte        // 请求的数据
	Response      []byte        // 返回的数据
	Chan          chan (error)
}

func CreateRpc() *Rpc {
	rpc := &Rpc{
		CallId: genNextRpcCallId(),
		Chan:   make(chan error),
	}
	if rpc.Timeout <= 0 {
		rpc.Timeout = DefaultRpcTimeout
	}
	return rpc
}

func (rpc *Rpc) Call() {
	ret := rpcMgr.AddRpc(rpc)
	if !ret {
		return
	}

	<-rpc.Chan
}

func (rpc *Rpc) Notify() {
	rpcMgr.AddRpc(rpc)
}
