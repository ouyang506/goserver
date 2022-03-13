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

	rpc.InitRpc()

	return mgr
}

func (mgr *NetMgr) Init(conf *config.Config) {
	mgr.conf = conf
}

func (mgr *NetMgr) Start() {
	rpcMgr := rpc.GetRpcManager()
	for i, gateInfo := range mgr.conf.Gates.GateInfos {
		rpcMgr.AddStub(int(consts.ServerTypeGate), i, gateInfo.IP, int(gateInfo.Port))
	}
}
