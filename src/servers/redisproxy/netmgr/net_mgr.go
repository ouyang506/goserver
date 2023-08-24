package netmgr

import (
	"common"
	"framework/registry"
	"framework/rpc"
	"redisproxy/config"
	"redisproxy/handler"
	"sync"
)

var (
	once           = sync.Once{}
	netMgr *NetMgr = nil
)

// singleton
func Instance() *NetMgr {
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
	msgHandler := handler.NewMessageHandler()
	rpc.InitRpc(rpc.RpcModeInner, msgHandler)

	// register self endpoint to center
	etcdConf := registry.EtcdConfig{
		Endpoints: conf.RegistryConf.EtcdConf.Endpoints.Items,
		Username:  conf.RegistryConf.EtcdConf.Username,
		Password:  conf.RegistryConf.EtcdConf.Password,
	}
	regCenter := registry.NewEtcdRegistry(etcdConf)

	skey := registry.ServiceKey{
		ServerType: common.ServerTypeRedisProxy,
		IP:         conf.ListenConf.Ip,
		Port:       conf.ListenConf.Port,
	}
	regCenter.RegService(skey)
	// fetch current all services and then watch
	rpc.FetchWatchService(rpc.RpcModeInner, regCenter)

	//start listen
	rpc.TcpListen(rpc.RpcModeInner, conf.ListenConf.Ip, conf.ListenConf.Port)
}
