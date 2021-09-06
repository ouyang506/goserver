package network

import (
	"goserver/common/log"
)

type NetworkCore interface {
	TcpListen(host string, port int) error
	TcpConnect(host string, port int) error
	TcpSend(int64, []byte) error
}

func NewNetworkCore(numLoops int, loadBalance LoadBalance, eventHandler NetEventHandler, logger log.Logger) NetworkCore {
	return newNetworkCore(numLoops, loadBalance, eventHandler, logger)
}

/// net connection event handler
type NetEventHandler interface {
	OnAccept(*Connection)
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

func (h *DefaultNetEventHandler) OnAccept(c *Connection) {
	h.logger.LogDebug("DefaultNetEventHandler OnAccept, connection info: %+v", c)
}

func (h *DefaultNetEventHandler) OnConnected(c *Connection) {
	h.logger.LogDebug("DefaultNetEventHandler OnConnected, connection info : %+v", c)
}

func (h *DefaultNetEventHandler) OnDisconnected(c *Connection) {
	h.logger.LogDebug("DefaultNetEventHandler OnDisconnected, connection info : %+v", c)
}
