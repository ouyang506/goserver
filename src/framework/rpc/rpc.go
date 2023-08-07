package rpc

import (
	"errors"
	"reflect"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"framework/consts"
	"framework/network"
	"utility/timer"

	"framework/proto/pb"
	_ "framework/proto/pb"

	"google.golang.org/protobuf/proto"
)

var (
	ErrorRpcMgrNotFound = errors.New("rpc manager not found")
	ErrorAddRpc         = errors.New("add rpc error")
	ErrorRpcTimeOut     = errors.New("rpc time out error")
	ErrorRpcRespMsgType = errors.New("rpc response msg type error")
)

const RpcHandlerMethodPrefix = "HandleRpc"

type RpcModeType int

const (
	DefaultRpcMode RpcModeType = RpcModeInner //默认为服务器内部rpc

	RpcModeInner RpcModeType = 1 //服务器内部rpc
	RpcModeOuter RpcModeType = 2 //客户端rpc
)

var (
	DefaultRpcTimeout = time.Second * 3
	nextRpcCallId     = int64(10000)

	rpcMgrMap sync.Map = sync.Map{} //RpcModeType->*RpcManager
)

type RpcEntry struct {
	RpcMode       RpcModeType // rpc模式
	CallId        int64       // rpc请求唯一ID
	MsgId         int         // 消息ID
	TargetSvrType int         // 目标服务类型
	RouteKey      string      // 路由key
	IsOneway      bool        // 是否单向通知

	Timeout   time.Duration // 超时时间
	WaitTimer *timer.Timer  // 超时定时器

	ReqMsg   proto.Message // 请求的msg数据
	RespMsg  proto.Message // 返回的msg数据
	RespChan chan (error)  // 收到对端返回或者超时通知
}

// 初始化rpc管理器
func InitRpc(mode RpcModeType, eventHandler network.NetEventHandler, msgHandler any) {
	rpcMgr := NewRpcManager(mode, eventHandler, msgHandler)
	AddRpcManager(mode, rpcMgr)
}

// 提供监听服务
func TcpListen(mode RpcModeType, ip string, port int) error {
	rpcMgr := GetRpcManager(mode)
	if rpcMgr == nil {
		return ErrorRpcMgrNotFound
	}
	return rpcMgr.rpcStubMgr.netcore.TcpListen(ip, port)
}

// rpc同步调用
func Call(targetSvrType int, req proto.Message, resp proto.Message, options ...Option) error {
	rpcMode := DefaultRpcMode
	return doCall(rpcMode, targetSvrType, req, resp, options...)
}

// rpc通知
func Notify(targetSvrType int, req proto.Message, options ...Option) error {
	rpcMode := DefaultRpcMode
	rpc := createRpc(RpcModeInner, targetSvrType, req, nil, options...)
	rpc.IsOneway = true
	rpcMgr := GetRpcManager(rpcMode)
	if rpcMgr == nil {
		return ErrorRpcMgrNotFound
	}
	ret := rpcMgr.AddRpc(rpc)
	if !ret {
		return ErrorAddRpc
	}
	return nil
}

// 外部rpc同步调用
func OuterCall(req proto.Message, resp proto.Message, options ...Option) (err error) {
	targetSvrType := consts.ServerTypeGate
	return doCall(RpcModeOuter, targetSvrType, req, resp, options...)
}

// rpc同步调用
func doCall(rpcMode RpcModeType, targetSvrType int,
	req proto.Message, resp proto.Message, options ...Option) (err error) {

	rpc := createRpc(rpcMode, targetSvrType, req, resp, options...)
	ret := false
	rpcMgr := GetRpcManager(rpcMode)
	if rpcMgr == nil {
		return ErrorRpcMgrNotFound
	}
	ret = rpcMgr.AddRpc(rpc)
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

func genNextRpcCallId() int64 {
	return atomic.AddInt64(&nextRpcCallId, 1)
}

func createRpc(rpcMode RpcModeType, targetSvrType int, req proto.Message,
	resp proto.Message, options ...Option) *RpcEntry {

	ops := LoadOptions(options...)
	reqMsgId := pb.GetMsgIdByName(reflect.TypeOf(req).Elem().Name())

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

func GetRpcManager(mode RpcModeType) *RpcManager {
	rpcMgr, ok := rpcMgrMap.Load(mode)
	if !ok {
		return nil
	}
	return rpcMgr.(*RpcManager)
}

func AddRpcManager(mode RpcModeType, rpcMgr *RpcManager) {
	rpcMgrMap.Store(mode, rpcMgr)
}
