package main

import (
	"common/log"
	"common/rpc"
	"time"
)

func main() {
	//init logger
	logger := log.NewCommonLogger()
	logger.AddSink(log.NewStdLogSink())
	logger.AddSink(log.NewFileLogSink("../log/", log.RotateByHour))
	logger.Start()
	log.SetLogger(logger)

	host := "127.0.0.1"
	port := 5000
	rpc.InitRpc()
	rpc.TcpListen(host, port)

	for {
		update()
		time.Sleep(time.Second)
	}
}

func update() {

}
