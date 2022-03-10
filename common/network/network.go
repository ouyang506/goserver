package network

type NetworkCore interface {
	Start()
	Stop()

	TcpListen(host string, port int) error
	TcpConnect(host string, port int, autoReconnect bool, attrib map[interface{}]interface{}) (Connection, error) //nonblock
	TcpSend(sessionId int64, buff []byte) error
	TcpClose(sessionId int64) error
}

func NewNetworkCore(options ...Option) NetworkCore {
	return newNetworkCore(options...)
}

/// net connection event handler
type NetEventHandler interface {
	OnAccept(c Connection)
	OnConnect(c Connection, err error)
	OnClosed(c Connection)
}
