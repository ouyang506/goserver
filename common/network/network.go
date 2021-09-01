package network

import (
	"goserver/common/log"
)

type NetworkCore interface {
	TcpListen(host string, port int) error
	TcpConnect(host string, port int) error
}

func NewNetworkCore(numLoops int, logger log.LoggerInterface) NetworkCore {
	return newNetworkCore(numLoops, logger)
}
