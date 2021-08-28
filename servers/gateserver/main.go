package main

import (
	"fmt"
	"goserver/common/network"
	"time"
)

func main() {
	fmt.Println("gateserver start!!")

	port := 9000
	host := "127.0.0.1"
	rcvBuffCap, sendBuffCap := 64*1024, 64*1024
	go func() {
		time.Sleep(time.Duration(5) * time.Second)
		network.TcpConnect(host, int32(port), int32(rcvBuffCap), int32(sendBuffCap))
		time.Sleep(time.Duration(50) * time.Second)
	}()

	network.TcpListen(host, int32(port), int32(rcvBuffCap), int32(sendBuffCap))

}
