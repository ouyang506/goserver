package network

import "sync/atomic"

type Connection struct {
	sessionId             int64
	fd                    int
	peerHost              string
	peerPort              int
	isClient              bool
	lastTryConnectionTime int64 // used as client
	connected             bool
	sendBuff              []byte
	rcvBuff               []byte
	attrMap               map[int]interface{}
}

func NewConnection() *Connection {
	c := &Connection{}
	c.sessionId = genNextSessionId()
	c.sendBuff = []byte{}
	c.rcvBuff = []byte{}
	c.attrMap = make(map[int]interface{})
	return c
}

var (
	nextSessionId = int64(0)
)

func genNextSessionId() int64 {
	return atomic.AddInt64(&nextSessionId, 1)
}

func (c *Connection) SetSessionId(sessionId int64) {
	c.sessionId = sessionId
}

func (c *Connection) GetSessionId() int64 {
	return c.sessionId
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

func (c *Connection) SetClient(v bool) {
	c.isClient = true
}

func (c *Connection) IsClient() bool {
	return c.isClient
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

func (c *Connection) SetAttrib(k int, v interface{}) {
	c.attrMap[k] = v
}

func (c *Connection) GetAttrib(k int) interface{} {
	v, ok := c.attrMap[k]
	if !ok {
		return nil
	}
	return v
}
