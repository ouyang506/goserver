package network

import (
	"common/log"
	"common/utility/ringbuffer"
	"errors"
	"runtime"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

type NetConn struct {
	BaseConn
	fd       int
	sendBuff *ringbuffer.RingBuffer
	rcvBuff  *ringbuffer.RingBuffer
}

func NewNetConn(sendBuffSize int, rcvBuffSize int) *NetConn {
	c := &NetConn{}
	c.sessionId = genNextSessionId()
	c.state = int32(ConnStateInit)
	c.attrMap = sync.Map{}
	c.sendBuff = ringbuffer.NewRingBuffer(sendBuffSize)
	c.rcvBuff = ringbuffer.NewRingBuffer(rcvBuffSize)

	return c
}

// NetPollCore implements the NetworkCore interface
type NetPollCore struct {
	logger log.Logger

	numLoops             int
	loadBalance          LoadBalance
	socketSendBufferSize int
	socketRcvBufferSize  int
	socketTcpNoDelay     bool
	codec                Codec

	polls      []*Poll
	acceptPoll *Poll

	listenFd      int32
	waitConnMap   sync.Map // sessionId->connection
	waitConnTimer time.Ticker

	eventHandler NetEventHandler
}

func newNetworkCore(opts ...Option) *NetPollCore {
	options := loadOptions(opts)
	if options.numLoops <= 0 {
		options.numLoops = runtime.NumCPU()
	}

	netcore := &NetPollCore{}
	netcore.numLoops = options.numLoops
	netcore.logger = options.logger
	netcore.loadBalance = options.loadBalance
	netcore.eventHandler = options.eventHandler
	netcore.socketSendBufferSize = options.socketSendBufferSize
	netcore.socketRcvBufferSize = options.socketRcvBufferSize
	netcore.socketTcpNoDelay = options.socketTcpNoDelay
	netcore.codec = options.codec
	netcore.startLoop()

	netcore.waitConnMap = sync.Map{}
	netcore.waitConnTimer = *time.NewTicker(time.Duration(100) * time.Millisecond)
	netcore.startWaitConnTimer()

	return netcore
}

// NetPollCore loop
func (netcore *NetPollCore) startLoop() error {
	pollFds := []int{}
	wakeFds := []int{}

	for i := 0; i < netcore.numLoops+1; i++ {
		pollFd, err := unix.EpollCreate1(unix.EPOLL_CLOEXEC)
		if err != nil {
			netcore.logger.LogError("EpollCreate1 error : %v", err)
			return err
		}

		poll := NewNetPoll()
		poll.pollIndex = i
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
		poll.connMap = map[int64]*NetConn{}
		poll.connFdMap = map[int]*NetConn{}

		poll.addRead(poll.wakeFd)

		if i == 0 {
			netcore.acceptPoll = poll
		} else {
			netcore.polls = append(netcore.polls, poll)
			pollFds = append(pollFds, poll.pollFd)
			wakeFds = append(wakeFds, poll.wakeFd)
		}

	}

	// start goroutines for loop
	go netcore.acceptPoll.loopEpollWait()
	for i := 0; i < netcore.numLoops; i++ {
		go func(poll *Poll) {
			poll.loopEpollWait()
		}(netcore.polls[i])
	}

	netcore.logger.LogInfo("netcore start loop, acceptPollFd: %v, numLoops: %v, pollFds: %+v, wakeFds: %+v",
		netcore.acceptPoll.pollFd, netcore.numLoops, pollFds, wakeFds)
	return nil
}

func (netcore *NetPollCore) startWaitConnTimer() {
	go func() {
		for t := range netcore.waitConnTimer.C {
			netcore.onWaitConnTimer(t)
		}
	}()
}

func (netcore *NetPollCore) onWaitConnTimer(t time.Time) {
	netcore.waitConnMap.Range(func(key, value interface{}) bool {
		conn := value.(*NetConn)
		if t.Unix()-conn.lastTryConTime < int64(RECONNECT_DELTA_TIME_SEC) {
			return true
		}
		conn.lastTryConTime = t.Unix()

		allocIndex := netcore.loadBalance.GetConnection(conn.sessionId)
		if allocIndex < 0 {
			return true
		}
		poll := netcore.polls[allocIndex]

		param := []interface{}{poll, conn}
		taskFunc := func(param interface{}) error {
			poll := param.([]interface{})[0].(*Poll)
			conn := param.([]interface{})[1].(*NetConn)

			_, err := poll.tcpConnect(conn)
			if err != nil {
				if err != unix.EINPROGRESS {
					poll.logger.LogError("tcp connect error, peerHost:%v, peerPort:%v, error : %s", conn.peerHost, conn.peerPort, err)
				}
			} else {
				poll.logger.LogInfo("tcp connect success, peerHost:%v, peerPort:%v", conn.peerHost, conn.peerPort)
			}
			return err
		}
		task := NewEventTask(taskFunc, param)
		poll.wake(task)

		return true
	})
}

func (netcore *NetPollCore) addWaitConn(conn *NetConn) {
	netcore.waitConnMap.Store(conn.sessionId, conn)
}

func (netcore *NetPollCore) removeWaitConn(sessionId int64) {
	netcore.waitConnMap.Delete(sessionId)
}

// implement network core TcpListen
func (netcore *NetPollCore) TcpListen(host string, port int) error {
	poll := netcore.acceptPoll

	param := []interface{}{poll, host, port}
	taskFunc := func(param interface{}) error {
		poll := param.([]interface{})[0].(*Poll)
		host := param.([]interface{})[1].(string)
		port := param.([]interface{})[2].(int)

		listenFd, err := poll.tcpListen(host, port)
		if err != nil {
			poll.logger.LogError("tcp listen at %v:%v error : %s", host, port, err)
		} else {
			poll.logger.LogInfo("start tcp listen at %v:%v, fd:%v", host, port, listenFd)
		}
		return err
	}
	task := NewEventTask(taskFunc, param)
	poll.wake(task)
	return nil
}

// implement network core TcpConnect
func (netcore *NetPollCore) TcpConnect(host string, port int) (int64, error) {

	conn := NewNetConn(netcore.socketSendBufferSize, netcore.socketRcvBufferSize)
	conn.isClient = true
	conn.peerHost = host
	conn.peerPort = port
	netcore.loadBalance.AllocConnection(conn.sessionId)

	netcore.addWaitConn(conn)
	return conn.GetSessionId(), nil
}

// implement network core TcpSend
func (netcore *NetPollCore) TcpSend(sessionId int64, buff []byte) error {
	pollIndex := netcore.loadBalance.GetConnection(sessionId)
	if pollIndex < 0 {
		netcore.logger.LogError("TcpSend connection poll not found, sessionId:%v", sessionId)
		return errors.New("connection poll not found")
	}
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

// implement network core TcpClose
func (netcore *NetPollCore) TcpClose(sessionId int64) error {
	pollIndex := netcore.loadBalance.GetConnection(sessionId)
	if pollIndex < 0 {
		netcore.logger.LogError("TcpClose connection poll not found, sessionId:%v", sessionId)
		return errors.New("connection poll not found")
	}
	poll := netcore.polls[pollIndex]
	param := []interface{}{poll, sessionId}
	taskFunc := func(param interface{}) error {
		poll := param.([]interface{})[0].(*Poll)
		sessionId := param.([]interface{})[1].(int64)

		conn := poll.getConnection(sessionId)
		if conn == nil {
			return errors.New("connnection not found")
		}
		poll.close(conn.fd)
		return nil
	}
	task := NewEventTask(taskFunc, param)
	poll.wake(task)
	return nil
}
