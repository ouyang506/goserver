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
	logger := log.NewLogger()
	logger.AddSink(log.NewStdLogSink())

	logger.LogInfo("gateserver start !")

	net := network.NewNetworkCore(numberLoops, logger)
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
