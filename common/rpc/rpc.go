package rpc

import (
	"strconv"
	"sync/atomic"
	"time"

	"common/utility/timer"

	"google.golang.org/protobuf/proto"
)

var (
	rpcMgr *RpcManager = nil
)

var (
	DefaultRpcTimeout = time.Second * 3
	nextRpcCallId     = int64(100)
)

func genNextRpcCallId() int64 {
	return atomic.AddInt64(&nextRpcCallId, 1)
}

// 初始化全局rpc管理器
func InitRpc() {
	InitMsgMapping()
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
	WaitTimer     *timer.Timer  // 定时器
}

func createRpc(targetSvrType int, reqMsgId int, req proto.Message, options ...Option) *Rpc {
	ops := LoadOptions(options...)

	rpc := &Rpc{
		CallId:   genNextRpcCallId(),
		RespChan: make(chan error),
	}

	if ops.RpcTimout > 0 {
		rpc.Timeout = ops.RpcTimout
	} else {
		rpc.Timeout = DefaultRpcTimeout
	}

	if ops.RouteKey != "" {
		rpc.RouteKey = ops.RouteKey
	} else {
		rpc.RouteKey = strconv.FormatInt(rpc.CallId, 10)
	}

	rpc.ReqMsg = &InnerMessage{
		Head:  InnerMessageHead{CallId: rpc.CallId, MsgID: reqMsgId},
		PbMsg: req,
	}

	return rpc
}

// rpc同步调用
func Call(targetSvrType int, reqMsgId int, req proto.Message, options ...Option) (resp proto.Message, err error) {
	rpc := createRpc(targetSvrType, reqMsgId, req, options...)
	ret := rpcMgr.AddRpc(rpc)
	if !ret {
		err = ErrorAddRpc
		return
	}

	// wait for rpc response util timeout
	err = <-rpc.RespChan
	if err != nil {
		return
	}

	respInnerMsg, ok := rpc.RespMsg.(*InnerMessage)
	if !ok {
		err = ErrorRpcRespMsgType
		return
	}
	resp = respInnerMsg.PbMsg
	return
}

// rpc通知
func Notify(targetSvrType int, reqMsgId int, req proto.Message, options ...Option) error {
	rpc := createRpc(targetSvrType, reqMsgId, req, options...)
	ret := rpcMgr.AddRpc(rpc)
	if !ret {
		return ErrorAddRpc
	}
	return nil
}
