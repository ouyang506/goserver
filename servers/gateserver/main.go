package main

import (
	"common/log"
	"common/network"
	"strconv"
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
	numberLoops := 0
	logger := log.NewCommonLogger()
	logger.AddSink(log.NewStdLogSink())
	logger.AddSink(log.NewFileLogSink("./log/"))
	logger.Start()
	eventHandler := NewCommNetEventHandler(logger)
	lb := network.NewLoadBalanceRoundRobin(numberLoops)

	logger.LogInfo("gateserver start .")

	net = network.NewNetworkCore(network.WithEventHandler(eventHandler), network.WithLogger(logger),
		network.WithNumLoop(numberLoops), network.WithLoadBalance(lb))

	go startClient(net, host, port)

	time.Sleep(time.Duration(2) * time.Second)
	net.TcpListen(host, port)

	update()
}

func startClient(net network.NetworkCore, peerHost string, peerPort int) {
	net.TcpConnect(peerHost, peerPort)
}

func update() {
	count := 1
	clienIndex := 1000
	serverIndex := 2000
	for {
		time.Sleep(time.Duration(1) * time.Second)

		connMap.Range(func(key, value interface{}) bool {
			c := value.(network.Connection)
			if c.IsClient() {
				clienIndex++
				net.TcpSend(key.(int64), []byte("hello , this is client "+strconv.Itoa(clienIndex)))
			} else {
				count++
				if count%10 == 0 {
					net.TcpClose(key.(int64))
				} else {
					serverIndex++
					net.TcpSend(key.(int64), []byte("hello , this is server "+strconv.Itoa(serverIndex)))
				}
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

func (h *CommNetEventHandler) OnAccept(c network.Connection) {
	peerHost, peerPort := c.GetPeerAddr()
	h.logger.LogDebug("NetEventHandler OnAccept, peerHost:%v, peerPort:%v", peerHost, peerPort)
	connMap.Store(c.GetSessionId(), c)
}

func (h *CommNetEventHandler) OnConnected(c network.Connection) {
	peerHost, peerPort := c.GetPeerAddr()
	h.logger.LogDebug("NetEventHandler OnConnected, peerHost:%v, peerPort:%v", peerHost, peerPort)
	connMap.Store(c.GetSessionId(), c)
}

func (h *CommNetEventHandler) OnClosed(c network.Connection) {
	peerHost, peerPort := c.GetPeerAddr()
	h.logger.LogDebug("NetEventHandler OnClosed, peerHost:%v, peerPort:%v", peerHost, peerPort)
	connMap.Delete(c.GetSessionId())
}
