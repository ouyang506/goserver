package main

import (
	"common/log"
	"common/network"
	"common/proto"
	"common/registry"
	"encoding/json"
	"fmt"
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
	sendBuffSize := 32 * 1024
	rcvBuffSize := 32 * 1024
	logger := log.NewCommonLogger()
	logger.AddSink(log.NewStdLogSink())
	logger.AddSink(log.NewFileLogSink("./log/"))
	logger.Start()

	doRegister(logger, "gateway", 1, host, port)

	eventHandler := NewCommNetEventHandler(logger)
	lb := network.NewLoadBalanceRoundRobin(numberLoops)
	codec := network.NewVariableFrameLenCodec()

	logger.LogInfo("gateserver start .")
	a := &proto.LoginGateReqT{}
	logger.LogDebug("a = %+v", a)

	net = network.NewNetworkCore(network.WithLogger(logger), network.WithEventHandler(eventHandler),
		network.WithNumLoop(numberLoops), network.WithLoadBalance(lb),
		network.WithSocketSendBufferSize(sendBuffSize), network.WithSocketRcvBufferSize(rcvBuffSize),
		network.WithSocketTcpNoDelay(true), network.WithFrameCodec(codec))

	go startClient(net, host, port)

	time.Sleep(time.Duration(2) * time.Second)
	net.TcpListen(host, port)

	update()
}

func startClient(net network.NetworkCore, peerHost string, peerPort int) {
	net.TcpConnect(peerHost, peerPort, true)
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
				//net.TcpSend(key.(int64), []byte("hello , this is client "+strconv.Itoa(clienIndex)))
			} else {
				count++
				if count%10 == 0 {
					//net.TcpClose(key.(int64))
				} else {
					serverIndex++
					//net.TcpSend(key.(int64), []byte("hello , this is server "+strconv.Itoa(serverIndex)))
				}
			}

			return true
		})
	}
}

type RegServiceEntry struct {
	ServiceType string `json:"service"`
	InstId      int    `json:"inst"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
}

func doRegister(logger log.Logger, serviceType string, instId int, host string, port int) {

	etcdClient := registry.NewEtcdRegistry(logger, []string{"127.0.0.1:2379"}, "", "")
	regMgr := registry.NewRegistryMgr(logger, etcdClient,
		func(registry.OperType, string, string) {
		},
	)

	key := fmt.Sprintf("/%s/%d", serviceType, instId)

	entry := RegServiceEntry{
		ServiceType: serviceType,
		InstId:      instId,
		Host:        host,
		Port:        port,
	}
	value, _ := json.Marshal(entry)

	regMgr.DoRegister(key, string(value), 10)
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

func (h *CommNetEventHandler) OnConnectFailed(c network.Connection) {
	peerHost, peerPort := c.GetPeerAddr()
	h.logger.LogDebug("NetEventHandler OnConnectFailed, peerHost:%v, peerPort:%v", peerHost, peerPort)
}

func (h *CommNetEventHandler) OnClosed(c network.Connection) {
	peerHost, peerPort := c.GetPeerAddr()
	h.logger.LogDebug("NetEventHandler OnClosed, peerHost:%v, peerPort:%v", peerHost, peerPort)
	connMap.Delete(c.GetSessionId())
}
