package netmgr

import (
	"framework/network"
	"framework/rpc"
)

// 网络事件回调
type NetEventHandler struct {
	rpc.OuterNetEventHandler
}

func NewNetEventHandler() *NetEventHandler {
	return &NetEventHandler{}
}

func (e *NetEventHandler) OnAccept(c network.Connection) {
	e.OuterNetEventHandler.OnAccept(c)
}

func (e *NetEventHandler) OnConnect(c network.Connection, err error) {
	e.OuterNetEventHandler.OnConnect(c, err)
}

func (e *NetEventHandler) OnClosed(c network.Connection) {
	e.OuterNetEventHandler.OnClosed(c)
}

func (e *NetEventHandler) OnRcvMsg(c network.Connection, msg interface{}) {
	e.OuterNetEventHandler.OnRcvMsg(c, msg)
}
