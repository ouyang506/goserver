package network

import (
	"fmt"
	"goserver/common/network/gnet"
	"time"
)

func TcpListen(host string, port int32, rcvBuffCap int32, sendBuffCap int32) {
	addr := fmt.Sprintf("%s:%d", host, port)

	eventHandler := &TcpEventHandler{}

	options := []gnet.Option{}
	options = append(options, gnet.WithTCPNoDelay(gnet.TCPNoDelay))
	options = append(options, gnet.WithSocketRecvBuffer(int(rcvBuffCap)))
	options = append(options, gnet.WithSocketSendBuffer(int(sendBuffCap)))
	options = append(options, gnet.WithTicker(false))
	options = append(options, gnet.WithMulticore(true))
	options = append(options, gnet.WithCodec(&TcpCodec{}))

	gnet.Serve(eventHandler, addr, options...)
}

func TcpConnect(host string, port int32, rcvBuffCap int32, sendBuffCap int32) {
	addr := fmt.Sprintf("%s:%d", host, port)

	eventHandler := &TcpEventHandler{}
	options := []gnet.Option{}
	options = append(options, gnet.WithTCPNoDelay(gnet.TCPNoDelay))
	options = append(options, gnet.WithSocketRecvBuffer(int(rcvBuffCap)))
	options = append(options, gnet.WithSocketSendBuffer(int(sendBuffCap)))
	options = append(options, gnet.WithTicker(false))
	options = append(options, gnet.WithCodec(&TcpCodec{}))

	gnet.Connect(eventHandler, addr, options...)
}

func AddFd(fd int32) {

}

// implements gnet.EventHandler
type TcpEventHandler struct {
}

func (h *TcpEventHandler) OnInitComplete(server gnet.Server) (action gnet.Action) {
	return gnet.None
}

func (h *TcpEventHandler) OnShutdown(server gnet.Server) {

}

func (h *TcpEventHandler) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	return nil, gnet.None
}

func (h *TcpEventHandler) OnClosed(c gnet.Conn, err error) (action gnet.Action) {
	return gnet.None
}

func (h *TcpEventHandler) PreWrite() {

}

func (h *TcpEventHandler) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	fmt.Printf("rcv msg : %v \n", frame)

	return nil, gnet.None
}

func (h *TcpEventHandler) Tick() (delay time.Duration, action gnet.Action) {
	return time.Duration(0), gnet.None
}
