package netmgr

import (
	"common/network"
	"common/rpc"
	"gateserver/config"
	"strings"
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

}

func (mgr *NetMgr) listenForClients() error {
	ip := strings.TrimSpace(mgr.conf.Outer.IP)
	port := mgr.conf.Outer.Port

	// 初始化网络
	codecs := []network.Codec{}
	codecs = append(codecs, NewOuterMessageCodec())
	codecs = append(codecs, network.NewVariableFrameLenCodec())

	netcore := network.NewNetworkCore(
		network.WithEventHandler(&NetEvent{}),
		network.WithLoadBalance(network.NewLoadBalanceRoundRobin(0)),
		network.WithSocketSendBufferSize(32*1024),
		network.WithSocketRcvBufferSize(32*1024),
		network.WithSocketTcpNoDelay(true),
		network.WithFrameCodecs(codecs))

	netcore.Start()

	err := netcore.TcpListen(ip, port)
	if err != nil {
		return err
	}

	return nil
}
