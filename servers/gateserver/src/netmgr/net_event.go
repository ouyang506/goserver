package netmgr

import (
	"common/log"
	"common/network"
)

var (
	ConnAttrPlayer = new(struct{})
)

type NetEvent struct {
}

func (e *NetEvent) OnAccept(c network.Connection) {
	peerHost, peerPort := c.GetPeerAddr()
	log.Info("NetEvent OnAccept, peerHost:%v, peerPort:%v", peerHost, peerPort)
}

func (e *NetEvent) OnConnect(c network.Connection, err error) {
	peerHost, peerPort := c.GetPeerAddr()
	if err != nil {
		log.Info("NetEvent OnConnectFailed, sessionId: %v, peerHost:%v, peerPort:%v", c.GetSessionId(), peerHost, peerPort)
	} else {
		log.Info("NetEvent OnConnected, sessionId: %v, peerHost:%v, peerPort:%v,", c.GetSessionId(), peerHost, peerPort)
	}
	c.SetAttrib(&ConnAttrPlayer, 0)
}

func (e *NetEvent) OnClosed(c network.Connection) {
	peerHost, peerPort := c.GetPeerAddr()
	log.Info("NetEvent OnClosed, sessionId : %v, peerHost:%v, peerPort:%v", c.GetSessionId(), peerHost, peerPort)
	c.SetAttrib(&ConnAttrPlayer, 0)
}

func (e *NetEvent) OnRcvMsg(c network.Connection, msg interface{}) {
	log.Debug("NetEvent OnRcvMsg, sessionId : %v, msg: %+v", c.GetSessionId(), msg)
}
