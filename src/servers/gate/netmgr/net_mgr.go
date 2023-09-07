package netmgr

import (
	"common"
	"framework/actor"
	"framework/registry"
	"framework/rpc"
	"gate/configmgr"
	"gate/logic/handler"
)

// 网络管理
type NetMgr struct {
	rootContext actor.Context
}

func NewNetMgr(rootContext actor.Context) *NetMgr {
	mgr := &NetMgr{
		rootContext: rootContext,
	}

	return mgr
}

func (mgr *NetMgr) Start() {
	conf := configmgr.Instance().GetConfig()

	// init rpc message handler
	msgHandler := handler.NewMessageHandler(mgr.rootContext)
	clientEventHandler := NewClientNetEventHandler(msgHandler)
	// startup outer rpc for clients
	rpc.InitRpc(rpc.RpcModeOuter, msgHandler, rpc.WithNetEventHandler(clientEventHandler))
	// startup inner rpc for server
	rpc.InitRpc(rpc.RpcModeInner, msgHandler)

	//start listen
	rpc.TcpListen(rpc.RpcModeInner, conf.ListenConf.Ip, conf.ListenConf.Port)
	rpc.TcpListen(rpc.RpcModeOuter, conf.Outer.ListenIp, conf.Outer.Port)

	// register self endpoint to center
	etcdConf := registry.EtcdConfig{
		Endpoints: conf.RegistryConf.EtcdConf.Endpoints.Items,
		Username:  conf.RegistryConf.EtcdConf.Username,
		Password:  conf.RegistryConf.EtcdConf.Password,
	}
	regCenter := registry.NewEtcdRegistry(etcdConf)

	skey := registry.ServiceKey{
		ServerType: common.ServerTypeGate,
		IP:         conf.ListenConf.Ip,
		Port:       conf.ListenConf.Port,
	}
	regCenter.RegService(skey)
	// fetch current all services and then watch
	rpc.FetchWatchService(rpc.RpcModeInner, regCenter)
}
