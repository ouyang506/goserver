package network

import "sync/atomic"

type Connection struct {
	sessionId             int64
	fd                    int
	peerHost              string
	peerPort              int
	lastTryConnectionTime int64 // used as client
	connected             bool
}

func NewConnection() *Connection {
	c := &Connection{}
	c.sessionId = genNextSessionId()
	return c
}

var (
	nextSessionId = int64(0)
)

func genNextSessionId() int64 {
	return atomic.AddInt64(&nextSessionId, 1)
}

func (c *Connection) SetFd(fd int) {
	c.fd = fd
}

func (c *Connection) GetFd() int {
	return c.fd
}

func (c *Connection) SetPeerAddr(host string, port int) {
	c.peerHost = host
	c.peerPort = port
}

func (c *Connection) GetPeerAddr() (string, int) {
	return c.peerHost, c.peerPort
}

func (c *Connection) SetLastTryConnectTime(t int64) {
	c.lastTryConnectionTime = t
}

func (c *Connection) GetLastTryConnectTime() int64 {
	return c.lastTryConnectionTime
}

func (c *Connection) SetConnected(flag bool) {
	c.connected = flag
}

func (c *Connection) IsConnected() bool {
	return c.connected
}
