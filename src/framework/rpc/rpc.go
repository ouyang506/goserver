package rpc

import (
	"strconv"
	"sync/atomic"
	"time"

	"framework/consts"
	"framework/network"
	"utility/timer"

	_ "framework/proto/pb"

	"google.golang.org/protobuf/proto"
)

type RpcModeType int

const (
	RpcModeInner RpcModeType = 1 //服务器内部rpc
	RpcModeOuter RpcModeType = 2 //客户端rpc
)

var (
	rpcMgr      *RpcManager = nil
	outerRpcMgr *RpcManager = nil
)

var (
	DefaultRpcTimeout = time.Second * 3
	nextRpcCallId     = int64(10000)
)

func GetRpcManager() *RpcManager {
	return rpcMgr
}

func GetOuterRpcManager() *RpcManager {
	return outerRpcMgr
}

func genNextRpcCallId() int64 {
	return atomic.AddInt64(&nextRpcCallId, 1)
}

// type RpcCallbackFunc func(err error, respInnerMsg *InnerMessage)

type RpcEntry struct {
	RpcMode         RpcModeType // rpc模式
	CallId          int64       // rpc请求唯一ID
	MsgId           int         // 消息ID
	TargetSvrType   int         // 目标服务类型
	TargetSvrInstId int         // 目标服务实例id
	RouteKey        string      // 路由key
	IsOneway        bool        // 是否单向通知

	Timeout   time.Duration // 超时时间
	WaitTimer *timer.Timer  // 超时定时器

	ReqMsg   proto.Message // 请求的msg数据
	RespMsg  proto.Message // 返回的msg数据
	RespChan chan (error)  // 收到对端返回或者超时通知
}

// 创建一个rpc
func createRpc(rpcMode RpcModeType, targetSvrType int, reqMsgId int, req proto.Message,
	resp proto.Message, options ...Option) *RpcEntry {

	ops := LoadOptions(options...)

	rpc := &RpcEntry{
		RpcMode:       rpcMode,
		CallId:        genNextRpcCallId(),
		MsgId:         reqMsgId,
		TargetSvrType: targetSvrType,
		ReqMsg:        req,
		RespMsg:       resp,
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

	return rpc
}

// rpc同步调用
func doCall(rpcMode RpcModeType, targetSvrType int,
	reqMsgId int, req proto.Message, resp proto.Message, options ...Option) (err error) {

	rpc := createRpc(rpcMode, targetSvrType, reqMsgId, req, resp, options...)
	ret := false
	if rpcMode == RpcModeInner {
		ret = rpcMgr.AddRpc(rpc)
	} else {
		ret = outerRpcMgr.AddRpc(rpc)
	}

	if !ret {
		err = ErrorAddRpc
		return
	}

	// wait for rpc response util timeout
	err = <-rpc.RespChan
	if err != nil {
		return
	}

	return
}

// 初始化服务器内部rpc管理器
func InitRpc(eventHandler network.NetEventHandler) {
	rpcMgr = NewRpcManager(RpcModeInner, eventHandler)
}

// 提供内部监听服务
func TcpListen(ip string, port int) {
	rpcMgr.rpcStubMgr.netcore.TcpListen(ip, port)
}

// rpc同步调用
func Call(targetSvrType int, reqMsgId int, req proto.Message,
	resp proto.Message, options ...Option) (err error) {
	return doCall(RpcModeInner, targetSvrType, reqMsgId, req, resp, options...)
}

// rpc通知
func Notify(targetSvrType int, reqMsgId int,
	req proto.Message, resp proto.Message, options ...Option) error {
	rpc := createRpc(RpcModeInner, targetSvrType, reqMsgId, req, resp, options...)
	rpc.IsOneway = true
	ret := rpcMgr.AddRpc(rpc)
	if !ret {
		return ErrorAddRpc
	}
	return nil
}

// 初始化客户端到服务器rpc管理器
func InitOuterRpc(eventHandler network.NetEventHandler) {
	outerRpcMgr = NewRpcManager(RpcModeOuter, eventHandler)
}

// 提供外部监听服务
func OuterTcpListen(ip string, port int) {
	outerRpcMgr.rpcStubMgr.netcore.TcpListen(ip, port)
}

// 外部rpc同步调用
func OuterCall(reqMsgId int, req proto.Message, resp proto.Message, options ...Option) (err error) {
	targetSvrType := consts.ServerTypeGate
	return doCall(RpcModeOuter, targetSvrType, reqMsgId, req, resp, options...)
}
