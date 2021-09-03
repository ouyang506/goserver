package network

import (
	"goserver/common/log"
)

type NetworkCore interface {
	TcpListen(host string, port int) error
	TcpConnect(host string, port int) error
	TcpSend(int, []byte) error
}

func NewNetworkCore(numLoops int, loadBalance LoadBalance, eventHandler NetEventHandler, logger log.Logger) NetworkCore {
	return newNetworkCore(numLoops, loadBalance, eventHandler, logger)
}

/// net connection event handler
type NetEventHandler interface {
	OnConnected(*Connection)
	OnDisconnected(*Connection)
}

type DefaultNetEventHandler struct {
	logger log.Logger
}

func NewDefaultNetEventHandler(logger log.Logger) NetEventHandler {
	return &DefaultNetEventHandler{
		logger: logger,
	}
}

func (h *DefaultNetEventHandler) OnConnected(c *Connection) {
	h.logger.LogDebug("DefaultNetEventHandler OnConnected, detail : %+v", c)
}

func (h *DefaultNetEventHandler) OnDisconnected(c *Connection) {
	h.logger.LogDebug("DefaultNetEventHandler OnDisconnected, detail : %+v", c)
}
