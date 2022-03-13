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

// 初始化全局rpc管理器
func InitRpc() {
	rpcMgr = NewRpcManager()
}

func TcpListen(ip string, port int) {
	rpcMgr.rpcStubMgr.netcore.TcpListen(ip, port)
}

func GetRpcManager() *RpcManager {
	return rpcMgr
}

type Rpc struct {
	CallId        int64         // rpc请求唯一ID
	Timeout       time.Duration // 超时时间
	TargetSvrType int           // 目标服务类型
	RouteKey      string        // 路由key
	ReqMsg        interface{}   // 请求的msg数据
	RespMsg       interface{}   // 返回的msg数据
	RespChan      chan (error)  // 收到对端返回或者超时通知
}

func CreateRpc() *Rpc {
	rpc := &Rpc{
		CallId:   genNextRpcCallId(),
		RespChan: make(chan error),
	}
	if rpc.Timeout <= 0 {
		rpc.Timeout = DefaultRpcTimeout
	}
	return rpc
}

func (rpc *Rpc) Call() error {
	ret := rpcMgr.AddRpc(rpc)
	if !ret {
		return ErrorAddRpc
	}

	err := <-rpc.RespChan
	if err != nil {
		return err
	}

	return nil
}

func (rpc *Rpc) Notify() error {
	ret := rpcMgr.AddRpc(rpc)
	if !ret {
		return ErrorAddRpc
	}
	return nil
}
