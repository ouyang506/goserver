package rpc

import (
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"framework/log"
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
	nextRpcCallId     atomic.Int64

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

	ReqMsg     proto.Message   // 请求的msg数据
	RespMsg    proto.Message   // 返回的msg数据
	NotifyChan chan (struct{}) // 收到对端返回通知

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
	return nextRpcCallId.Add(1)
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
		NotifyChan:    make(chan struct{}, 1),
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

// 关闭连接
func TcpClose(mode RpcModeType, sessionId int64) error {
	rpcMgr := GetRpcManager(mode)
	if rpcMgr == nil {
		return ErrorRpcMgrNotFound
	}
	return rpcMgr.TcpClose(sessionId)
}

// 发送消息
func TcpSend(mode RpcModeType, sessionId int64, msg any) error {
	rpcMgr := GetRpcManager(mode)
	if rpcMgr == nil {
		return ErrorRpcMgrNotFound
	}
	return rpcMgr.TcpSendMsg(sessionId, msg)
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

// rpc通知
func Notify(targetSvrType int, guid int64, req proto.Message, options ...Option) error {
	rpcMode := DefaultRpcMode
	rpcMgr := GetRpcManager(rpcMode)
	if rpcMgr == nil {
		return ErrorRpcMgrNotFound
	}

	rpcEntry := createRpc(RpcModeInner, targetSvrType, guid, req, nil, options...)
	rpcEntry.IsOneway = true

	addRpcTimeout := rpcEntry.Timeout
	tryAddRpcCount := 5
	addResult := false
	for i := 0; i < tryAddRpcCount; i++ {
		addResult = rpcMgr.AddRpc(rpcEntry)
		if addResult {
			break
		}
		delta := time.Duration(int64(addRpcTimeout) / int64(tryAddRpcCount-1))
		time.Sleep(delta)
	}

	if !addResult {
		return ErrorAddRpc
	}
	return nil
}

// 外部rpc同步调用
func OuterCall(targetSvrType int, guid int64, req proto.Message, resp proto.Message, options ...Option) error {
	return doCall(RpcModeOuter, targetSvrType, guid, req, resp, options...)
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

	rpcEntry := createRpc(rpcMode, targetSvrType, guid, req, resp, options...)

	// 1.由于服务启动时序问题，可能暂未拉取到注册中心的服务，导致没有分配到代理管道
	// 2.发送缓冲区已满
	timeout := rpcEntry.Timeout
	addRpcTimeout := timeout / 2
	tryAddRpcCount := 5
	addResult := false
	for i := 0; i < tryAddRpcCount; i++ {
		addResult = rpcMgr.AddRpc(rpcEntry)
		if addResult {
			break
		}
		delta := time.Duration(int64(addRpcTimeout) / int64(tryAddRpcCount-1))
		time.Sleep(delta)
		timeout -= delta
	}

	if !addResult {
		err = ErrorAddRpc
		return
	}
	defer rpcMgr.RemoveRpc(rpcEntry.CallId)

	// wait for rpc response util timeout
	bTimeout := false
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-rpcEntry.NotifyChan:
		break
	case <-timer.C:
		bTimeout = true
	}

	if bTimeout {
		err = ErrorRpcTimeOut
		log.Error("rpc timeout, callId: %v", rpcEntry.CallId)
		rpcEntry.setTimeoutFlag()
	}

	return
}
