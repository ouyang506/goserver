package netmgr

import (
	"common/consts"
	"common/rpc"
	"mockclient/config"
)

// 网络管理
type NetMgr struct {
	conf   *config.Config
	rpcMgr *rpc.RpcManager
}

func NewNetMgr() *NetMgr {
	mgr := &NetMgr{}
	mgr.rpcMgr = rpc.NewRpcManager()
	rpc.InitRpc(mgr.rpcMgr)
	return mgr
}

func (mgr *NetMgr) Init(conf *config.Config) {
	mgr.conf = conf
}

func (mgr *NetMgr) Start() {
	for i, gateInfo := range mgr.conf.Gates.GateInfos {
		stubMgr := mgr.rpcMgr.GetStubMgr()
		stubMgr.AddStub(int(consts.ServerTypeGate), i, gateInfo.IP, int(gateInfo.Port))
	}
}
