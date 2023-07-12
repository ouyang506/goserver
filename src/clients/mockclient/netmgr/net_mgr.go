package netmgr

import (
	"framework/consts"
	"framework/rpc"
	"mockclient/config"
	"mockclient/handler"
)

// 网络管理
type NetMgr struct {
	conf *config.Config
}

func NewNetMgr() *NetMgr {
	mgr := &NetMgr{}

	rpc.InitOuterRpc(NewNetMessageEvent(), handler.NewMessageHandler())

	return mgr
}

func (mgr *NetMgr) Init(conf *config.Config) {
	mgr.conf = conf
}

func (mgr *NetMgr) Start() {
	outerRpcMgr := rpc.GetOuterRpcManager()
	for i, gateInfo := range mgr.conf.Gates.GateInfos {
		outerRpcMgr.AddStub(int(consts.ServerTypeGate), i, gateInfo.IP, int(gateInfo.Port))
		break
	}
}
