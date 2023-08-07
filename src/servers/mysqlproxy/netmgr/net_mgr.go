package netmgr

import (
	"framework/rpc"
	"mysqlproxy/config"
	"mysqlproxy/handler"
)

// 网络管理
type NetMgr struct {
	conf       *config.Config
	msgHandler *handler.MessageHandler
}

func NewNetMgr(conf *config.Config) *NetMgr {
	mgr := &NetMgr{
		conf: conf,
	}

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

}
