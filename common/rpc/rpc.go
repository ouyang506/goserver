package rpc

var (
	rpcMgr *RpcManager = nil
)

// 使用rpc前需要先初始化rpc管理器
func InitRpc(mgr *RpcManager) {
	rpcMgr = mgr
}

type Rpc struct {
	SessionID     int64 // rpc请求唯一ID
	ErrCode       int32 // rpc返回的错误码
	TargetSvrType int   // 目标服务类型
	RouteKey      int64 // 路由key
	Request       []byte
	Response      []byte
}

func CreateRpc() *Rpc {
	rpc := &Rpc{}
	return rpc
}

func (rpc *Rpc) Call() {
	rpcMgr.AddRpc(rpc)
}
