package network

import (
	"goserver/common/log"
)

type NetworkMgr struct {
	logger   log.LoggerInterface
	pollFds  []int
	numLoops int
	listenFd int
	connMap  map[int]*Connection // all connections
}

func NewNetworkMgr(numLoops int, logger log.LoggerInterface) *NetworkMgr {
	mgr := &NetworkMgr{}
	mgr.numLoops = numLoops
	mgr.logger = logger
	mgr.loop()
	return mgr
}

// func (mgr *NetworkMgr) SetLogger(logger *log.LoggerInterface) {
// 	mgr.logger = logger
// }

// func (mgr *NetworkMgr) GetLogger() *log.LoggerInterface {
// 	return mgr.logger
// }

// func (mgr *NetworkMgr) SetNumLoops(numLoops int) {
// 	mgr.numLoops = numLoops
// }

func (mgr *NetworkMgr) loop() {
	mgr.openPoll()
}

func (mgr *NetworkMgr) TcpListen(host string, port int) {
	err := mgr.tcpListen(host, port)
	if err != nil {
		mgr.logger.LogError("tcp listen at %v:%v error : %s", host, port, err)
	} else {
		mgr.logger.LogInfo("start tcp listen at %v:%v, listen fd: %v", host, port, mgr.listenFd)
	}
}

func (mgr *NetworkMgr) TcpConnect(host string, port int) {
	err := mgr.tcpConnect(host, port)
	if err != nil {
		mgr.logger.LogError("tcp connect error, host:%v, port:%v, error : %s", host, port, err)
	} else {
		mgr.logger.LogError("tcp connect ok, host:%v, port:%v", host, port)
	}
}

func (mgr *NetworkMgr) AddConnection(fd int, c*Connection){
	_, ok := mgr.connMap[fd]
	if ok{
		mgr.logger.LogWarn("add a existed fd to connection map, fd : %v", fd)
	}
	mgr.connMap[fd] = c
}

func (mgr *NetworkMgr) RemoveConnection(fd int){
	_, ok := mgr.connMap[fd]
	if !ok{
		mgr.logger.LogWarn("remome connection not found, fd : %v", fd)
		return
	}
	delete(mgr.connMap, fd)
}


func (mgr *NetworkMgr) GetConnection(fd int) *Connection{
	c, ok := mgr.connMap[fd]
	if !ok{
		return nil
	}
	return c
}