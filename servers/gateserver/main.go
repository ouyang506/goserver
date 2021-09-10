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
	port := 9000
	numberLoops := 4
	logger := log.NewDefaultLogger()
	logger.AddSink(log.NewStdLogSink())
	eventHandler := NewCommNetEventHandler(logger)
	lb := network.NewLoadBalanceRoundRobin(numberLoops)

	logger.LogInfo("gateserver start !")

	net = network.NewNetworkCore(numberLoops, lb, eventHandler, logger)

	go startClient(net, host, port)
	time.Sleep(time.Duration(2) * time.Second)
	net.TcpListen(host, port)

	update()
}

func startClient(net network.NetworkCore, peerHost string, peerPort int) {
	//time.Sleep(time.Duration(2) * time.Second)
	net.TcpConnect(peerHost, peerPort)
}

func update() {
	for {
		time.Sleep(time.Duration(1) * time.Second)

		connMap.Range(func(key, value interface{}) bool {
			c := value.(*network.Connection)
			if c.IsClient() {
				net.TcpSend(key.(int64), []byte("hello , this is client .."))
			} else {
				net.TcpSend(key.(int64), []byte("hello , this is server"))
			}

			return true
		})
	}
}

type CommNetEventHandler struct {
	logger log.Logger
}

func NewCommNetEventHandler(logger log.Logger) network.NetEventHandler {
	return &CommNetEventHandler{
		logger: logger,
	}
}

func (h *CommNetEventHandler) OnAccept(c *network.Connection) {
	h.logger.LogDebug("CommNetEventHandler OnAccept, connection info: %+v", c)
	connMap.Store(c.GetSessionId(), c)
}

func (h *CommNetEventHandler) OnConnected(c *network.Connection) {
	h.logger.LogDebug("CommNetEventHandler OnConnected, connection info : %+v", c)
	//peerHost, peerHost := c.GetPeerAddr()
	connMap.Store(c.GetSessionId(), c)
}

func (h *CommNetEventHandler) OnDisconnected(c *network.Connection) {
	h.logger.LogDebug("CommNetEventHandler OnDisconnected, connection info : %+v", c)
	connMap.Delete(c.GetSessionId())
}
