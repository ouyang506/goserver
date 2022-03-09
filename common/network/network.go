package network

import (
	"common/log"
)

type NetworkCore interface {
	TcpListen(host string, port int) error
	TcpConnect(host string, port int, autoReconnect bool) (Connection, error) //nonblock
	TcpSend(sessionId int64, buff []byte) error
	TcpClose(sessionId int64) error
}

func NewNetworkCore(options ...Option) NetworkCore {
	return newNetworkCore(options...)
}

/// net connection event handler
type NetEventHandler interface {
	OnAccept(c Connection)
	OnConnected(c Connection)
	OnConnectFailed(c Connection)
	OnClosed(c Connection)
}

type DefaultNetEventHandler struct {	
}

func NewDefaultNetEventHandler() NetEventHandler {
	return &DefaultNetEventHandler{}
}

func (h *DefaultNetEventHandler) OnAccept(c Connection) {
	log.Debug("DefaultNetEventHandler OnAccept, connection info: %+v", c)
}

func (h *DefaultNetEventHandler) OnConnected(c Connection) {
	log.Debug("DefaultNetEventHandler OnConnected, connection info : %+v", c)
}

func (h *DefaultNetEventHandler) OnConnectFailed(c Connection) {
	log.Debug("DefaultNetEventHandler OnConnectFailed, connection info : %+v", c)
}

func (h *DefaultNetEventHandler) OnClosed(c Connection) {
	log.Debug("DefaultNetEventHandler OnClosed, connection info : %+v", c)
}
