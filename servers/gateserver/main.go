package main

import (
	"fmt"
	"goserver/common/log"
	"goserver/common/network"
	"time"
)

func main() {
	fmt.Println("gateserver start!!")

	host := "127.0.0.1"
	port := 9000
	numberLoops := 4
	logger := log.NewLogger()
	logger.AddSink(log.NewStdLogSink())

	net := network.NewNetworkMgr(numberLoops, logger)
	net.TcpListen(host, port)

	update()

}

func update() {
	for {
		time.Sleep(time.Duration(1) * time.Second)
	}
}
