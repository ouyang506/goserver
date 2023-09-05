package netmgr

import (
	"framework/log"
	"framework/network"
	"framework/rpc"
)

type NetAttrPlayer struct{}

// 网络事件回调
type ClientNetEventHandler struct {
	rpc.OuterNetEventHandler
}

func NewClientNetEventHandler() *ClientNetEventHandler {
	return &ClientNetEventHandler{}
}

func (e *ClientNetEventHandler) OnAccept(c network.Connection) {
	e.OuterNetEventHandler.OnAccept(c)
}

func (e *ClientNetEventHandler) OnConnect(c network.Connection, err error) {
	//服务器不会主动去连接客户端
	log.Error("unexpected network event OnConnect")
}

func (e *ClientNetEventHandler) OnClosed(c network.Connection) {
	e.OuterNetEventHandler.OnClosed(c)
}

func (e *ClientNetEventHandler) OnRcvMsg(c network.Connection, msg interface{}) {
	e.OuterNetEventHandler.OnRcvMsg(c, msg)
}
