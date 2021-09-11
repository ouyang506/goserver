package network

import "sync/atomic"

type ConnState int

const (
	ConnStateInit ConnState = iota
	ConnStateConnecting
	ConnStateConnected
	ConnStateClosed
)

type Connection struct {
	sessionId int64
	fd        int
	host      string
	port      int
	peerHost  string
	peerPort  int
	state     ConnState

	isClient       bool
	lastTryConTime int64

	sendBuff []byte
	rcvBuff  []byte

	attrMap map[int]interface{}
}

func NewConnection() *Connection {
	c := &Connection{}
	c.sessionId = genNextSessionId()
	c.state = ConnStateInit
	c.sendBuff = []byte{}
	c.rcvBuff = []byte{}
	c.attrMap = make(map[int]interface{})
	return c
}

var (
	nextSessionId = int64(10000)
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

func (c *Connection) SetAddr(host string, port int) {
	c.host = host
	c.port = port
}

func (c *Connection) GetAddr() (string, int) {
	return c.host, c.port
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
	c.lastTryConTime = t
}

func (c *Connection) GetLastTryConnectTime() int64 {
	return c.lastTryConTime
}

func (c *Connection) SetConnecting() {
	c.state = ConnStateConnecting
}

func (c *Connection) IsConnecting() bool {
	return c.state == ConnStateConnecting
}

func (c *Connection) SetConnected() {
	c.state = ConnStateConnected
}

func (c *Connection) IsConnected() bool {
	return c.state == ConnStateConnected
}

func (c *Connection) SetClosed() {
	c.state = ConnStateClosed
}

func (c *Connection) IsClosed() bool {
	return c.state == ConnStateClosed
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
