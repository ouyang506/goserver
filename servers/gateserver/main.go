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

	net := network.NewNetworkMgr(numberLoops, logger)
	net.TcpListen(host, port)

	go startClient(net, host, port)

	update()
}

func startClient(net *network.NetworkMgr, peerHost string, peerPort int) {
	net.TcpConnect(peerHost, peerPort)
}

func update() {
	for {
		time.Sleep(time.Duration(1) * time.Second)
	}
}
