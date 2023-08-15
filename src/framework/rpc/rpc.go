package rpc

import (
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"framework/consts"
	"framework/registry"

	"framework/proto/pb"

	"google.golang.org/protobuf/proto"
)

var (
	ErrorRpcMgrNotFound = errors.New("rpc manager not found")
	ErrorAddRpc         = errors.New("add rpc failed")
	ErrorRpcTimeOut     = errors.New("rpc time out error")
	ErrorRpcRespMsgType = errors.New("rpc response msg type error")
)

const RpcHandlerMethodPrefix = "HandleRpc"

type RpcModeType int

const (
	DefaultRpcMode RpcModeType = RpcModeInner //默认为服务器内部rpc
	RpcModeInner   RpcModeType = 1            //服务器内部rpc
	RpcModeOuter   RpcModeType = 2            //客户端rpc
)

var (
	DefaultRpcTimeout = time.Second * 5
	nextRpcCallId     = int64(0)

	rpcMgrMap sync.Map = sync.Map{} //RpcModeType->*RpcManager
)

type RpcEntry struct {
	RpcMode       RpcModeType // rpc模式
	TargetSvrType int         // 目标服务类型
	IsOneway      bool        // 是否单向通知

	CallId    int64        // rpc请求唯一ID
	MsgId     int          // 消息ID
	Guid      int64        // InnerMsg传输guid,同时作为路由的key
	RouteType RpcRouteType // 路由方式

	ReqMsg   proto.Message        // 请求的msg数据
	RespMsg  proto.Message        // 返回的msg数据
	RespChan chan (proto.Message) // 收到对端返回

	Timeout   time.Duration // 超时时间
	IsTimeout atomic.Bool
}

func (r *RpcEntry) setTimeoutFlag() {
	r.IsTimeout.Store(true)
}

func (r *RpcEntry) getTimeoutFlag() bool {
	return r.IsTimeout.Load()
}

func genNextRpcCallId() int64 {
	return atomic.AddInt64(&nextRpcCallId, 1)
}

func createRpc(rpcMode RpcModeType, targetSvrType int, guid int64, req proto.Message,
	resp proto.Message, options ...Option) *RpcEntry {

	ops := LoadOptions(options...)
	reqMsgId := pb.GetMsgIdByName(reflect.TypeOf(req).Elem().Name())

	rpc := &RpcEntry{
		RpcMode:       rpcMode,
		CallId:        genNextRpcCallId(),
		MsgId:         reqMsgId,
		IsOneway:      false,
		TargetSvrType: targetSvrType,
		Guid:          guid,
		ReqMsg:        req,
		RespMsg:       resp,
		RespChan:      make(chan proto.Message),
	}

	if ops.RpcTimout != nil {
		rpc.Timeout = *ops.RpcTimout
	} else {
		rpc.Timeout = DefaultRpcTimeout
	}

	if ops.RpcRouteType != nil {
		rpc.RouteType = *ops.RpcRouteType
	} else {
		rpc.RouteType = RandomRoute
	}

	return rpc
}

// 初始化rpc管理器
// msgHandler为struct指针
func InitRpc(mode RpcModeType, msgHandler any, options ...Option) {
	rpcMgr := NewRpcManager(mode, msgHandler, options...)
	rpcMgrMap.Store(mode, rpcMgr)
}

// 根据类型获取rpc管理器
func GetRpcManager(mode RpcModeType) *RpcManager {
	rpcMgr, ok := rpcMgrMap.Load(mode)
	if !ok {
		return nil
	}
	return rpcMgr.(*RpcManager)
}

// 监听端口
func TcpListen(mode RpcModeType, ip string, port int) error {
	rpcMgr := GetRpcManager(mode)
	if rpcMgr == nil {
		return ErrorRpcMgrNotFound
	}
	return rpcMgr.TcpListen(ip, port)
}

// 获取注册中心服务以及监听服务变化事件
func FetchWatchService(mode RpcModeType, reg registry.Registry) error {
	rpcMgr := GetRpcManager(mode)
	if rpcMgr == nil {
		return ErrorRpcMgrNotFound
	}
	rpcMgr.FetchWatchService(reg)
	return nil
}

// rpc同步调用
func Call(targetSvrType int, guid int64, req proto.Message, resp proto.Message, options ...Option) error {
	rpcMode := DefaultRpcMode
	return doCall(rpcMode, targetSvrType, guid, req, resp, options...)
}

// rpc to mysql proxy server
func CallMysqlProxy(req proto.Message, resp proto.Message, options ...Option) error {
	return Call(consts.ServerTypeMysqlProxy, 0, req, resp, options...)
}

// rpc通知
func Notify(targetSvrType int, guid int64, req proto.Message, options ...Option) error {
	rpcMode := DefaultRpcMode
	rpcMgr := GetRpcManager(rpcMode)
	if rpcMgr == nil {
		return ErrorRpcMgrNotFound
	}

	rpc := createRpc(RpcModeInner, targetSvrType, guid, req, nil, options...)
	rpc.IsOneway = true
	rpc.Timeout = 0

	addResult := false
	for i := 0; i < 3; i++ {
		addResult = rpcMgr.AddRpc(rpc)
		if addResult {
			break
		}
		time.Sleep(1 * time.Second)
	}

	if !addResult {
		return ErrorAddRpc
	}
	return nil
}

// 外部rpc同步调用(client->gate)
func OuterCall(req proto.Message, resp proto.Message, options ...Option) (err error) {
	targetSvrType := consts.ServerTypeGate
	return doCall(RpcModeOuter, targetSvrType, 0, req, resp, options...)
}

// rpc同步调用
func doCall(rpcMode RpcModeType, targetSvrType int, guid int64,
	req proto.Message, resp proto.Message, options ...Option) (err error) {
	if req == nil {
		return errors.New("request param nil error")
	}

	if resp == nil {
		return errors.New("response param nil error")
	}

	rpcMgr := GetRpcManager(rpcMode)
	if rpcMgr == nil {
		return ErrorRpcMgrNotFound
	}

	rpc := createRpc(rpcMode, targetSvrType, guid, req, resp, options...)

	// 由于服务启动时序问题，可能暂未拉取到注册中心的服务，导致没有分配到代理管道
	// 等待一段时间进行重试操作
	addResult := false
	for i := 0; i < 3; i++ {
		addResult = rpcMgr.AddRpc(rpc)
		if addResult {
			break
		}
		time.Sleep(1 * time.Second)
	}

	if !addResult {
		err = ErrorAddRpc
		return
	}
	defer rpcMgr.RemoveRpc(rpc.CallId)

	// wait for rpc response util timeout
	bTimeout := false
	tick := time.NewTicker(rpc.Timeout)
	defer tick.Stop()
	select {
	case rpc.RespMsg = <-rpc.RespChan:
		bTimeout = false
	case <-tick.C:
		bTimeout = true
	}

	if bTimeout {
		err = ErrorRpcTimeOut
		rpc.setTimeoutFlag()
	}

	return
}
