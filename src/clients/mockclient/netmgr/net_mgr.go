package netmgr

import (
	"common"
	"framework/rpc"
	"mockclient/handler"
)

// 网络管理
type NetMgr struct {
}

func NewNetMgr() *NetMgr {
	mgr := &NetMgr{}
	return mgr
}

func (mgr *NetMgr) Start() {
	rpc.InitRpc(rpc.RpcModeOuter, handler.NewMessageHandler(), rpc.WithNetEventHandler(NewNetEventHandler()))
}

func (mgr *NetMgr) RemoveGateStubs() {
	rpcmgr := rpc.GetRpcManager(rpc.RpcModeOuter)
	if rpcmgr == nil {
		return
	}
	rpcmgr.DelStubsByType(common.ServerTypeGate)
}

func (mgr *NetMgr) AddGateStub(remoteIP string, remotePort int) {
	rpcmgr := rpc.GetRpcManager(rpc.RpcModeOuter)
	if rpcmgr == nil {
		return
	}
	rpcmgr.AddStub(common.ServerTypeGate, remoteIP, remotePort)
}

func (mgr *NetMgr) FindGateStub(remoteIP string, remotePort int) bool {
	rpcmgr := rpc.GetRpcManager(rpc.RpcModeOuter)
	if rpcmgr == nil {
		return false
	}
	return rpcmgr.FindStub(common.ServerTypeGate, remoteIP, remotePort) != nil
}
