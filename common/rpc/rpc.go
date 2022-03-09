package rpc

import (
	"errors"
	"time"
)

var (
	rpcMgr *RpcManager = nil
)

var (
	PushRpcError error = errors.New("push rpc error")
)

// 使用rpc前需要先初始化rpc管理器
func InitRpc(mgr *RpcManager) {
	rpcMgr = mgr
}

type Rpc struct {
	SessionID     int64  // rpc请求唯一ID
	Oneway        bool   // 是否等待返回
	ErrCode       int32  // rpc返回的错误码
	TargetSvrType int    // 目标服务类型
	RouteKey      string // 路由key
	Request       []byte
	Response      []byte
	Chan          chan (error)
}

func CreateRpc() *Rpc {
	rpc := &Rpc{}
	return rpc
}

func (rpc *Rpc) Call() {
	if !rpc.Oneway {
		rpc.Chan = make(chan error)
	}

	ret := rpcMgr.AddRpc(rpc)
	if !ret {
		return
	}

	if !rpc.Oneway {
		tick := time.NewTicker(time.Second * 3)
		defer tick.Stop()
		select {
		case err := <-rpc.Chan:
			{
				if err != nil {
					rpc.ErrCode = 1
					return
				}
			}
		case <-tick.C:
			{

			}
		}
	}
}
