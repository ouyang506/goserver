package netmgr

import (
	"framework/consts"
	"framework/registry"
	"framework/rpc"
	"mysqlproxy/config"
	"mysqlproxy/handler"
	"sync"
)

var (
	once           = sync.Once{}
	netMgr *NetMgr = nil
)

// singleton
func GetNetMgr() *NetMgr {
	once.Do(func() {
		netMgr = NewNetMgr()
	})
	return netMgr
}

// 网络管理
type NetMgr struct {
}

func NewNetMgr() *NetMgr {
	mgr := &NetMgr{}
	return mgr
}

func (mgr *NetMgr) Start() {
	conf := config.GetConfig()

	// startup rpc
	netEventhandler := NewNetEventHandler()
	msgHandler := handler.NewMessageHandler()
	rpc.InitRpc(rpc.RpcModeInner, msgHandler, rpc.WithNetEventHandler(netEventhandler))
	rpc.TcpListen(rpc.RpcModeInner, conf.ListenConf.Ip, conf.ListenConf.Port)

	etcdConf := registry.EtcdConfig{
		Endpoints: conf.RegistryConf.EtcdConf.Endpoints.Items,
		Username:  conf.RegistryConf.EtcdConf.Username,
		Password:  conf.RegistryConf.EtcdConf.Password,
	}

	regCenter := registry.NewEtcdRegistry(etcdConf)
	rpc.RegisterService(rpc.RpcModeInner, regCenter,
		consts.ServerTypeMysqlProxy, conf.ListenConf.Ip, conf.ListenConf.Port)
	rpc.FetchWatchService(rpc.RpcModeInner, regCenter)
}
