package netmgr

import (
	"framework/rpc"
	"gate/config"
	"strings"
)

// 网络管理
type NetMgr struct {
	conf *config.Config
}

func NewNetMgr() *NetMgr {
	mgr := &NetMgr{}

	return mgr
}

func (mgr *NetMgr) Init(conf *config.Config) {
	mgr.conf = conf

	rpc.InitOuterRpc(NewNetMessageEvent())
}

func (mgr *NetMgr) Start() {
	mgr.listenForClients()
}

func (mgr *NetMgr) listenForClients() error {
	ip := strings.TrimSpace(mgr.conf.Outer.IP)
	port := mgr.conf.Outer.Port

	rpc.OuterTcpListen(ip, port)

	return nil
}
