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
	// startup rpc
	netEventhandler := NewNetEventHandler()
	msgHandler := handler.NewMessageHandler()
	rpc.InitRpc(rpc.RpcModeInner, msgHandler, rpc.WithNetEventHandler(netEventhandler))

	conf := config.GetConfig()
	etcdConf := registry.EtcdConfig{
		Endpoints: conf.RegistryConf.EtcdConf.Endpoints.Items,
		Username:  conf.RegistryConf.EtcdConf.Username,
		Password:  conf.RegistryConf.EtcdConf.Password,
	}
	rpc.RegisterServiceToEtcd(rpc.RpcModeInner, etcdConf,
		consts.ServerTypeMysqlProxy, conf.ListenConf.Ip, conf.ListenConf.Port)
}
