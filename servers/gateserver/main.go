package main

import (
	"goserver/common/log"
	"goserver/common/network"
	"time"
)

func main() {
	host := "127.0.0.1"
	port := 9000
	numberLoops := 4
	logger := log.NewDefaultLogger()
	logger.AddSink(log.NewStdLogSink())
	eventHandler := network.NewDefaultNetEventHandler(logger)
	lb := network.NewLoadBalanceRoundRobin(numberLoops)

	logger.LogInfo("gateserver start !")

	net := network.NewNetworkCore(numberLoops, lb, eventHandler, logger)
	net.TcpListen(host, port)

	go startClient(net, host, port)

	update()
}

func startClient(net network.NetworkCore, peerHost string, peerPort int) {
	time.Sleep(time.Duration(2) * time.Second)
	net.TcpConnect(peerHost, peerPort)
}

func update() {
	for {
		time.Sleep(time.Duration(1) * time.Second)
	}
}
