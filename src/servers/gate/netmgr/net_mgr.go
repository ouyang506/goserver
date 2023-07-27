package netmgr

import (
	"framework/rpc"
	"gate/config"
	"gate/handler"
	"strings"
)

// 网络管理
type NetMgr struct {
	conf       *config.Config
	msgHandler *handler.MessageHandler
}

func NewNetMgr() *NetMgr {
	mgr := &NetMgr{}

	return mgr
}

func (mgr *NetMgr) Init(conf *config.Config) {
	mgr.conf = conf

	// init rpc message handler
	mgr.msgHandler = handler.NewMessageHandler()

	// startup rpc
	netEventhandler := NewNetMessageEvent()
	rpc.InitRpc(rpc.RpcModeOuter, netEventhandler, mgr.msgHandler)
	netEventhandler.SetRpcMgr(rpc.GetRpcManager(rpc.RpcModeOuter))
}

func (mgr *NetMgr) Start() {
	mgr.listenForClients()
}

func (mgr *NetMgr) listenForClients() error {
	ip := strings.TrimSpace(mgr.conf.Outer.IP)
	port := mgr.conf.Outer.Port

	rpc.TcpListen(rpc.RpcModeOuter, ip, port)

	return nil
}
