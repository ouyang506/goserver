package main

import (
	"fmt"
	"goserver/common/network"
)

func main() {
	fmt.Println("gateserver start!!")

	port := 9000
	rcvBuffCap, sendBuffCap := 64*1024, 64*1024
	network.TcpListen(int32(port), int32(rcvBuffCap), int32(sendBuffCap))
}
