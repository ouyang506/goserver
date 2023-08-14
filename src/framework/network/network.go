package network

// thread safe
type NetworkCore interface {
	Start()
	Stop()
	TcpListen(host string, port int) error
	TcpConnect(host string, port int, autoReconnect bool) (Connection, error) //nonblock
	TcpSendMsg(sessionId int64, msg interface{}) error
	TcpClose(sessionId int64) error
}

func NewNetworkCore(options ...Option) NetworkCore {
	return newNetworkCore(options...)
}

// / net connection event handler
type NetEventHandler interface {
	OnAccept(c Connection)
	OnConnect(c Connection, err error)
	OnClosed(c Connection)
	OnRcvMsg(c Connection, msg interface{})
}
