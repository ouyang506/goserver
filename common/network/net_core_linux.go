package network

import (
	"golang.org/x/sys/unix"
	"goserver/common/log"

	"sync"
)

// NetPollCore implements the NetworkCore interface
type NetPollCore struct {
	logger       log.Logger
	numLoops     int
	loadBalance  LoadBalance
	eventHandler NetEventHandler
	listenFd     int32
	connMap      map[int]*Connection // all established connections
	connMapMutex sync.Mutex
	polls        []*Poll
}

func newNetworkCore(numLoops int, loadBalance LoadBalance, eventHandler NetEventHandler, logger log.Logger) *NetPollCore {
	netcore := &NetPollCore{}
	netcore.numLoops = numLoops
	netcore.logger = logger
	netcore.loadBalance = loadBalance
	netcore.eventHandler = eventHandler
	netcore.startLoop()
	return netcore
}

// NetPollCore loop
func (netcore *NetPollCore) startLoop() error {
	for i := 0; i < netcore.numLoops; i++ {
		pollFd, err := unix.EpollCreate1(unix.EPOLL_CLOEXEC)
		if err != nil {
			netcore.logger.LogError("EpollCreate1 error : %v", err)
			return err
		}

		poll := NewNetPoll()
		poll.netcore = netcore
		poll.logger = netcore.logger
		poll.eventHandler = netcore.eventHandler
		poll.pollFd = pollFd

		poll.wakeFd, err = unix.Eventfd(0, unix.EFD_NONBLOCK|unix.EFD_CLOEXEC)
		if err != nil {
			netcore.logger.LogError("create Eventfd error : %v", err)
			return err
		}
		poll.wfdBuf = make([]byte, 8)
		poll.wakeEventQueue = EventTaskQueue{}
		poll.connMap = map[int64]*Connection{}
		poll.connFdMap = map[int]*Connection{}
		poll.waitingConnFdMap = map[int]*Connection{}

		poll.addRead(poll.wakeFd)

		netcore.logger.LogInfo("create poll[%v] success, pollFd: %d, wakeFd: %v", i+1, poll.pollFd, poll.wakeFd)

		netcore.polls = append(netcore.polls, poll)
	}

	// start goroutines for loop
	for i := 0; i < netcore.numLoops; i++ {
		go func(poll *Poll) {
			poll.loopEpollWait()
		}(netcore.polls[i])
	}
	return nil
}

// implement network core TcpListen
func (netcore *NetPollCore) TcpListen(host string, port int) error {
	poll := netcore.polls[0]

	param := []interface{}{poll, host, port}
	taskFunc := func(param interface{}) error {
		poll := param.([]interface{})[0].(*Poll)
		host := param.([]interface{})[1].(string)
		port := param.([]interface{})[2].(int)

		listenFd, err := poll.tcpListen(host, port)
		if err != nil {
			poll.logger.LogError("tcp listen at %v:%v error : %s", host, port, err)
		} else {
			poll.logger.LogInfo("start tcp listen at %v:%v, listen fd: %v", host, port, listenFd)
		}
		return err
	}
	task := NewEventTask(taskFunc, param)
	poll.wake(task)
	return nil
}

// implement network core TcpConnect
func (netcore *NetPollCore) TcpConnect(host string, port int) error {
	poll := netcore.polls[1]

	param := []interface{}{poll, host, port}
	taskFunc := func(param interface{}) error {
		poll := param.([]interface{})[0].(*Poll)
		host := param.([]interface{})[1].(string)
		port := param.([]interface{})[2].(int)

		err := poll.tcpConnect(host, port)
		if err != nil {
			poll.logger.LogError("tcp connect error, host:%v, port:%v, error : %s", host, port, err)
		} else {
			poll.logger.LogInfo("tcp connect ok, host:%v, port:%v", host, port)
		}
		return err
	}
	task := NewEventTask(taskFunc, param)
	poll.wake(task)

	return nil
}

// implement network core TcpSend
func (netcore *NetPollCore) TcpSend(sessionId int64, buff []byte) error {
	pollIndex := netcore.loadBalance.GetConnection(sessionId)
	poll := netcore.polls[pollIndex]
	param := []interface{}{poll, sessionId, buff}
	taskFunc := func(param interface{}) error {
		poll := param.([]interface{})[0].(*Poll)
		sessionId := param.([]interface{})[1].(int64)
		buff := param.([]interface{})[2].([]byte)

		err := poll.TcpSend(sessionId, buff)
		return err
	}
	task := NewEventTask(taskFunc, param)
	poll.wake(task)
	return nil
}
