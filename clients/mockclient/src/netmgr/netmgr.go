package netmgr

import (
	"common/enums"
	"common/rpc"
	"mockclient/config"
)

// 网络管理
type NetMgr struct {
	rpcMgr *rpc.RpcManager
}

func NewNetMgr() *NetMgr {
	mgr := &NetMgr{}
	mgr.rpcMgr = rpc.NewRpcManager()
	rpc.InitRpc(mgr.rpcMgr)
	return mgr
}

func (mgr *NetMgr) Init(conf *config.Config) {
	for i, gateInfo := range conf.Gates.GateInfos {
		stubMgr := mgr.rpcMgr.GetStubMgr()
		stubMgr.AddStub(&rpc.RpcStub{
			ServerType:       int(enums.ServerTypeGate),
			ServerTypeInstID: i,
			RemoteIP:         gateInfo.IP,
			RemotePort:       int(gateInfo.Port),
		})
	}
}
