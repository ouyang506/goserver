package network

import (
	"common/log"
)

type NetworkCore interface {
	TcpListen(host string, port int) error
	TcpConnect(host string, port int) (Connection, error)
	TcpSend(int64, []byte) error
	TcpClose(int64) error
}

func NewNetworkCore(options ...Option) NetworkCore {
	return newNetworkCore(options...)
}

/// net connection event handler
type NetEventHandler interface {
	OnAccept(Connection)
	OnConnected(Connection)
	OnClosed(Connection)
}

type DefaultNetEventHandler struct {
	logger log.Logger
}

func NewDefaultNetEventHandler(logger log.Logger) NetEventHandler {
	return &DefaultNetEventHandler{
		logger: logger,
	}
}

func (h *DefaultNetEventHandler) OnAccept(c Connection) {
	h.logger.LogDebug("DefaultNetEventHandler OnAccept, connection info: %+v", c)
}

func (h *DefaultNetEventHandler) OnConnected(c Connection) {
	h.logger.LogDebug("DefaultNetEventHandler OnConnected, connection info : %+v", c)
}

func (h *DefaultNetEventHandler) OnClosed(c Connection) {
	h.logger.LogDebug("DefaultNetEventHandler OnClosed, connection info : %+v", c)
}
