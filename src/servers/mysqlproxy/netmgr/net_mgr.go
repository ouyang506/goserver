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

	// register self endpoint to center
	etcdConf := registry.EtcdConfig{
		Endpoints: conf.RegistryConf.EtcdConf.Endpoints.Items,
		Username:  conf.RegistryConf.EtcdConf.Username,
		Password:  conf.RegistryConf.EtcdConf.Password,
	}
	regCenter := registry.NewEtcdRegistry(etcdConf)

	skey := registry.ServiceKey{
		ServerType: consts.ServerTypeMysqlProxy,
		IP:         conf.ListenConf.Ip,
		Port:       conf.ListenConf.Port,
	}
	regCenter.RegService(skey)

	// fetch current all services and then watch
	rpc.FetchWatchService(rpc.RpcModeInner, regCenter)
}
