package rpc

import (
	"strconv"
	"sync/atomic"
	"time"

	"utility/timer"

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

type RpcCallbackFunc func(err error, respInnerMsg *InnerMessage)

type Rpc struct {
	CallId          int64  // rpc请求唯一ID
	TargetSvrType   int    // 目标服务类型
	TargetSvrInstId int    // 目标服务实例id
	RouteKey        string // 路由key
	IsOneway        bool   // 是否单向通知

	IsAsync  bool            // 是否是异步调用
	Callback RpcCallbackFunc // 异步调用回调函数

	Timeout   time.Duration // 超时时间
	WaitTimer *timer.Timer  // 超时定时器

	ReqMsg   interface{}  // 请求的msg数据
	RespMsg  interface{}  // 返回的msg数据
	RespChan chan (error) // 收到对端返回或者超时通知
}

// 创建一个rpc
func createRpc(targetSvrType int, reqMsgId int, req proto.Message, options ...Option) *Rpc {
	ops := LoadOptions(options...)

	rpc := &Rpc{
		CallId:        genNextRpcCallId(),
		TargetSvrType: targetSvrType,
		RespChan:      make(chan error),
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

// rpc异步调用
// callback会在其他协程中调用，不要阻塞
func AsyncCall(targetSvrType int, reqMsgId int, req proto.Message,
	callback RpcCallbackFunc, options ...Option) (err error) {

	rpc := createRpc(targetSvrType, reqMsgId, req, options...)
	rpc.IsAsync = true
	rpc.Callback = callback
	ret := rpcMgr.AddRpc(rpc)
	if !ret {
		err = ErrorAddRpc
		return
	}
	return
}

// rpc通知
func Notify(targetSvrType int, reqMsgId int, req proto.Message, options ...Option) error {
	rpc := createRpc(targetSvrType, reqMsgId, req, options...)
	rpc.IsOneway = true
	ret := rpcMgr.AddRpc(rpc)
	if !ret {
		return ErrorAddRpc
	}
	return nil
}

// 广播
func BroadcastMsg(targetSvrType int, reqMsgId int, req proto.Message, options ...Option) error {

	return nil
}

// // 转发消息
// func TransferMsg(targetSvrType int, reqMsgId int, req proto.Message, options ...Option) error {
// 	return Notify(targetSvrType, reqMsgId, req, options...)
// }

// 发送消息给指定服务实例
func SendMsgToServer(targetSvrType int, targetSvrInstId int, msgId int, req proto.Message, options ...Option) error {
	rpc := createRpc(targetSvrType, msgId, req, options...)
	rpc.TargetSvrInstId = targetSvrInstId
	ret := rpcMgr.AddRpc(rpc)
	if !ret {
		return ErrorAddRpc
	}
	return nil
}

// 通过网络连接的sessionId发送消息
func SendMsgByConnId(connSessionId int64, msgId int, msg proto.Message) error {
	return GetRpcManager().SendMsgByConnId(connSessionId, msgId, msg)
}
