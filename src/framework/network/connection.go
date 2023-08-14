package network

import (
	"sync"
	"sync/atomic"
	"utility/ringbuffer"
)

type ConnState int

const (
	ConnStateInit ConnState = iota
	ConnStateConnecting
	ConnStateConnected
	ConnStateClosed
)

const (
	RECONNECT_DELTA_TIME_SEC = 2
)

var (
	nextSessionId = int64(0)
)

func genNextSessionId() int64 {
	return atomic.AddInt64(&nextSessionId, 1)
}

type Connection interface {
	GetSessionId() int64
	GetAddr() (string, int)
	GetPeerAddr() (peerIp string, peerPort int)
	IsClient() bool
	GetState() ConnState
	CasState(oldState ConnState, newState ConnState) bool
	GetSendBuff() *ringbuffer.RingBuffer
	GetRcvBuff() *ringbuffer.RingBuffer
	SetAttrib(k interface{}, v interface{})
	GetAttrib(k interface{}) (interface{}, bool)
}

type BaseConn struct {
	sessionId int64

	host     string
	port     int
	peerHost string
	peerPort int
	state    int32

	isClient       bool
	autoReconnect  bool
	lastTryConTime int64

	sendBuff *ringbuffer.RingBuffer
	rcvBuff  *ringbuffer.RingBuffer

	attrMap sync.Map
}

func (c *BaseConn) GetSessionId() int64 {
	return c.sessionId
}

func (c *BaseConn) GetAddr() (string, int) {
	return c.host, c.port
}

func (c *BaseConn) GetPeerAddr() (string, int) {
	return c.peerHost, c.peerPort
}

func (c *BaseConn) IsClient() bool {
	return c.isClient
}

func (c *BaseConn) GetState() ConnState {
	return ConnState(atomic.LoadInt32(&c.state))
}

func (c *BaseConn) SetConnState(v ConnState) {
	atomic.StoreInt32(&c.state, int32(v))
}

func (c *BaseConn) CasState(oldState ConnState, newState ConnState) bool {
	return atomic.CompareAndSwapInt32(&c.state, int32(oldState), int32(newState))
}

func (c *BaseConn) GetSendBuff() *ringbuffer.RingBuffer {
	return c.sendBuff
}

func (c *BaseConn) GetRcvBuff() *ringbuffer.RingBuffer {
	return c.rcvBuff
}

func (c *BaseConn) SetAttrib(k interface{}, v interface{}) {
	c.attrMap.Store(k, v)
}

func (c *BaseConn) GetAttrib(k interface{}) (interface{}, bool) {
	return c.attrMap.Load(k)
}
