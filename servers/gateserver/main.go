package main

import (
	"common/log"
	"common/network"
	"sync"
	"time"
)

var (
	connMap sync.Map = sync.Map{}
	net     network.NetworkCore
)

func main() {

	host := "127.0.0.1"
	port := 5000
	numberLoops := 0
	sendBuffSize := 32 * 1024
	rcvBuffSize := 32 * 1024

	//init logger
	logger := log.NewCommonLogger()
	logger.AddSink(log.NewStdLogSink())
	logger.AddSink(log.NewFileLogSink("../log/", log.RotateByHour))
	logger.Start()
	log.SetLogger(logger)

	eventHandler := NewCommNetEventHandler(logger)
	lb := network.NewLoadBalanceRoundRobin(numberLoops)
	codec := network.NewVariableFrameLenCodec()

	logger.LogInfo("gateserver start .")

	net = network.NewNetworkCore(network.WithEventHandler(eventHandler),
		network.WithNumLoop(numberLoops), network.WithLoadBalance(lb),
		network.WithSocketSendBufferSize(sendBuffSize), network.WithSocketRcvBufferSize(rcvBuffSize),
		network.WithSocketTcpNoDelay(true), network.WithFrameCodec(codec))

	net.TcpListen(host, port)
	net.Start()

	for {
		update()
		time.Sleep(time.Second)
	}
}

func update() {

}

type CommNetEventHandler struct {
	logger log.Logger
}

func NewCommNetEventHandler(logger log.Logger) network.NetEventHandler {
	return &CommNetEventHandler{
		logger: logger,
	}
}

func (h *CommNetEventHandler) OnAccept(c network.Connection) {
	peerHost, peerPort := c.GetPeerAddr()
	h.logger.LogDebug("NetEventHandler OnAccept, peerHost:%v, peerPort:%v", peerHost, peerPort)
	connMap.Store(c.GetSessionId(), c)
}

func (h *CommNetEventHandler) OnConnect(c network.Connection, err error) {
	peerHost, peerPort := c.GetPeerAddr()
	if err == nil {
		h.logger.LogDebug("NetEventHandler OnConnected, peerHost:%v, peerPort:%v", peerHost, peerPort)
		connMap.Store(c.GetSessionId(), c)
	} else {
		h.logger.LogDebug("NetEventHandler OnConnectFailed, peerHost:%v, peerPort:%v", peerHost, peerPort)
	}
}

func (h *CommNetEventHandler) OnClosed(c network.Connection) {
	peerHost, peerPort := c.GetPeerAddr()
	h.logger.LogDebug("NetEventHandler OnClosed, peerHost:%v, peerPort:%v", peerHost, peerPort)
	connMap.Delete(c.GetSessionId())
}
