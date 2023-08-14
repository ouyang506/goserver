package network

import (
	"errors"
	"fmt"
	"framwork/log"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
	"utility/ringbuffer"

	"golang.org/x/sys/unix"
)

type NetConn struct {
	BaseConn
	fd int
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
	numLoops             int
	loadBalance          LoadBalance
	socketSendBufferSize int
	socketRcvBufferSize  int
	socketTcpNoDelay     bool
	codecs               []Codec

	polls      []*Poll
	acceptPoll *Poll

	listenFd      int32
	waitConnMap   sync.Map // sessionId->connection
	waitConnTimer time.Ticker

	eventHandlers []NetEventHandler
}

func newNetworkCore(opts ...Option) *NetPollCore {
	options := loadOptions(opts)
	if options.numLoops <= 0 {
		options.numLoops = runtime.NumCPU()
	}

	netcore := &NetPollCore{}
	netcore.numLoops = options.numLoops
	netcore.loadBalance = options.loadBalance
	netcore.eventHandlers = options.eventHandlers
	netcore.socketSendBufferSize = options.socketSendBufferSize
	netcore.socketRcvBufferSize = options.socketRcvBufferSize
	netcore.socketTcpNoDelay = options.socketTcpNoDelay
	netcore.codecs = options.codecs

	netcore.waitConnMap = sync.Map{}
	netcore.waitConnTimer = *time.NewTicker(time.Duration(100) * time.Millisecond)

	return netcore
}

func (netcore *NetPollCore) Start() {
	netcore.startLoop()
	netcore.startWaitConnTimer()
}

func (netcore *NetPollCore) Stop() {

}

// NetPollCore loop
func (netcore *NetPollCore) startLoop() error {
	pollFds := []int{}
	wakeFds := []int{}

	for i := 0; i < netcore.numLoops+1; i++ {
		pollFd, err := unix.EpollCreate1(unix.EPOLL_CLOEXEC)
		if err != nil {
			log.Error("EpollCreate1 error : %v", err)
			return err
		}

		poll := NewNetPoll()
		poll.pollIndex = i
		poll.netcore = netcore
		poll.eventHandlers = netcore.eventHandlers
		poll.pollFd = pollFd

		poll.wakeFd, err = unix.Eventfd(0, unix.EFD_NONBLOCK|unix.EFD_CLOEXEC)
		if err != nil {
			log.Error("create Eventfd error : %v", err)
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

	log.Info("netcore start loop, acceptPollFd: %v, numLoops: %v, pollFds: %+v, wakeFds: %+v",
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

			err := poll.tcpConnect(conn)
			if err != nil {
				if err != unix.EINPROGRESS {
					log.Error("tcp connect error, peerHost:%v, peerPort:%v, error : %s", conn.peerHost, conn.peerPort, err)
				}
			} else {
				log.Info("tcp connect success, peerHost:%v, peerPort:%v", conn.peerHost, conn.peerPort)
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
			log.Error("tcp listen at %v:%v error : %s", host, port, err)
		} else {
			log.Info("start tcp listen at %v:%v, fd:%v", host, port, listenFd)
		}
		return err
	}
	task := NewEventTask(taskFunc, param)
	poll.wake(task)
	return nil
}

// implement network core TcpConnect
func (netcore *NetPollCore) TcpConnect(host string, port int, autoReconnect bool) (Connection, error) {

	conn := NewNetConn(netcore.socketSendBufferSize, netcore.socketRcvBufferSize)
	conn.isClient = true
	conn.autoReconnect = autoReconnect
	conn.peerHost = host
	conn.peerPort = port
	netcore.loadBalance.AllocConnection(conn.sessionId)

	netcore.addWaitConn(conn)
	return conn, nil
}

// implement network core TcpSendMsg
func (netcore *NetPollCore) TcpSendMsg(sessionId int64, msg interface{}) error {
	pollIndex := netcore.loadBalance.GetConnection(sessionId)
	if pollIndex < 0 {
		log.Error("TcpSendMsg connection poll not found, sessionId:%v", sessionId)
		return errors.New("connection poll not found")
	}
	poll := netcore.polls[pollIndex]
	param := []interface{}{poll, sessionId, msg}
	taskFunc := func(param interface{}) error {
		poll := param.([]interface{})[0].(*Poll)
		sessionId := param.([]interface{})[1].(int64)
		msg := param.([]interface{})[2]

		err := poll.TcpSendMsg(sessionId, msg)
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
		log.Error("TcpClose connection poll not found, sessionId:%v", sessionId)
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

// ****************** net poll ******************
var (
	wakeInt64 = int64(1)
	wakeBytes = (*(*[8]byte)(unsafe.Pointer(&wakeInt64)))[:]
)

// net poll, one poll => one goroutine
type Poll struct {
	pollIndex      int
	netcore        *NetPollCore
	logger         log.Logger
	eventHandlers  []NetEventHandler
	pollFd         int
	wakeFd         int
	wfdBuf         []byte
	wakeEventQueue EventTaskQueue
	connMap        map[int64]*NetConn
	connFdMap      map[int]*NetConn
}

func NewNetPoll() *Poll {
	return &Poll{}
}

func (poll *Poll) tcpListen(host string, port int) (int, error) {
	log.Debug("poll tcp listen called , pollFd : %v", poll.pollFd)

	tcpAddr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		log.Error("tcpListen ResolveTCPAddr error : %v", err)
		return 0, err
	}

	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM|unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC, unix.IPPROTO_TCP)
	if err != nil {
		log.Error("tcpListen create socket error : %v", err)
		return 0, err
	}

	sa4 := &unix.SockaddrInet4{Port: tcpAddr.Port}
	if tcpAddr.IP != nil {
		if len(tcpAddr.IP) == 16 {
			copy(sa4.Addr[:], tcpAddr.IP[12:16]) // copy last 4 bytes of slice to array
		} else {
			copy(sa4.Addr[:], tcpAddr.IP) // copy all bytes of slice to array
		}
	}

	err = unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
	if err != nil {
		log.Error("tcpListen set reuse addr error : %v", err)
		return 0, err
	}

	err = unix.Bind(fd, sa4)
	if err != nil {
		log.Error("tcpListen bind error : %v", err)
		return 0, err
	}

	err = unix.Listen(fd, unix.SOMAXCONN)
	if err != nil {
		log.Error("tcpListen listen error : %v", err)
		return 0, err
	}

	atomic.StoreInt32(&poll.netcore.listenFd, int32(fd))

	// set listen fd nonblock
	err = unix.SetNonblock(fd, true)
	if err != nil {
		log.Error("tcpListen set nonblock error : %v", err)
		return 0, err
	}

	// add listen fd to epoll
	err = poll.addRead(fd)
	if err != nil {
		log.Error("tcpListen add read error : %v", err)
		return 0, err
	}

	return fd, nil
}

func (poll *Poll) tcpConnect(conn *NetConn) error {

	if conn.state == int32(ConnStateConnected) {
		poll.netcore.removeWaitConn(conn.sessionId)
		if poll.getConnection(conn.sessionId) == nil {
			poll.addConnection(conn)
		}
		return nil
	}

	if conn.state == int32(ConnStateConnecting) {
		if conn.fd > 0 {
			poll.close(conn.fd)
		}
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%d", conn.peerHost, conn.peerPort))
	if err != nil {
		log.Error("tcpConnect ResolveTCPAddr error : %v", err)
		return err
	}

	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM|unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC, unix.IPPROTO_TCP)
	if err != nil {
		log.Error("tcpConnect create socket error : %v", err)
		return err
	}

	if poll.netcore.socketSendBufferSize > 0 {
		err = unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_RCVBUF, poll.netcore.socketSendBufferSize)
		if err != nil {
			log.Error("tcpConnect set socket send buffer size option error: %v", err)
			return err
		}
	}

	if poll.netcore.socketRcvBufferSize > 0 {
		err = unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_RCVBUF, poll.netcore.socketRcvBufferSize)
		if err != nil {
			log.Error("tcpConnect set socket rcv buffer size option error: %v", err)
			return err
		}
	}

	tcpNodelay := 0
	if poll.netcore.socketTcpNoDelay {
		tcpNodelay = 1
	}
	err = unix.SetsockoptInt(fd, unix.IPPROTO_TCP, unix.TCP_NODELAY, tcpNodelay)
	if err != nil {
		log.Error("tcpConnect set socket tcp_nodelay option error: %v", err)
		return err
	}

	sa4 := &unix.SockaddrInet4{Port: tcpAddr.Port}
	if tcpAddr.IP != nil {
		if len(tcpAddr.IP) == 16 {
			copy(sa4.Addr[:], tcpAddr.IP[12:16])
		} else {
			copy(sa4.Addr[:], tcpAddr.IP)
		}
	}

	if poll.getConnection(conn.sessionId) != nil {
		poll.removeConnection(conn.sessionId)
	}

	conn.fd = fd
	poll.addConnection(conn)

	err = unix.Connect(fd, sa4)
	if err != nil {
		if err == unix.EINPROGRESS {
			log.Info("tcpConnect connect peer endpoint in progress, peerHost:%v, peerPort:%v, fd:%v",
				conn.peerHost, conn.peerPort, conn.fd)
			conn.state = int32(ConnStateConnecting)
			poll.addReadWrite(fd)
			return err
		}
		log.Error("tcpConnect connect peer endpoint error : %v, peerHost:%v, peerPort:%v", err, conn.peerHost, conn.peerPort)
		return err
	}

	//connected success directly
	conn.state = int32(ConnStateConnected)
	poll.addConnection(conn)
	poll.addReadWrite(fd)
	for _, h := range poll.eventHandlers {
		h.OnConnect(conn, nil)
	}
	poll.netcore.removeWaitConn(conn.sessionId)

	return nil
}

func (poll *Poll) TcpSendMsg(sessionId int64, msg interface{}) error {
	if msg == nil {
		return nil
	}

	c := poll.getConnection(sessionId)
	if c == nil {
		return errors.New("connection session not found")
	}
	if c.state != int32(ConnStateConnected) {
		return errors.New("connection closed")
	}

	var in interface{} = nil
	in = msg
	for _, codec := range poll.netcore.codecs {
		out, bChain, err := codec.Encode(c, in)
		in = out
		if err != nil {
			log.Error("TcpSendMsg encode buff error:%s, sessionId:%d", err, c.sessionId)
			poll.close(c.fd)
			return err
		}
		if !bChain {
			break
		}
	}

	return poll.loopWrite(c.fd) //TODO:会多一次event_mod
}

func (poll *Poll) wake(t *EventTask) error {

	poll.wakeEventQueue.Enqueue(t)

	for {
		_, err := unix.Write(poll.wakeFd, wakeBytes)
		if err != nil {
			log.Warn("poll wake error : %v, %v, %v", err, unix.EINTR, unix.EAGAIN)
			if err == unix.EINTR || err == unix.EAGAIN {
				continue
			} else {
				return err
			}
		} else {
			//log.Debug("send wake event to wakeFd:%v", poll.wakeFd)
			break
		}
	}
	return nil
}

func (poll *Poll) loopEpollWait() error {
	events := make([]unix.EpollEvent, 1024)
	for {
		n, err := unix.EpollWait(poll.pollFd, events, 100)
		if err != nil && err != unix.EINTR {
			log.Error("loop epoll wait error : %v", err)
			return err
		}

		for i := 0; i < n; i++ {
			eventFd := int(events[i].Fd)
			pollEvent := events[i].Events

			// log.Debug("loopEpollWait trigger event, poll: %v, eventfd : %d, events :%v",
			// 	poll.pollIndex, eventFd, pollEvent)

			if eventFd == poll.wakeFd {
				_, _ = unix.Read(eventFd, poll.wfdBuf)
				for task := poll.wakeEventQueue.Dequeue(); task != nil; task = poll.wakeEventQueue.Dequeue() {
					task.eventFunc(task.param)
				}
				continue
			}

			if eventFd == int(atomic.LoadInt32(&poll.netcore.listenFd)) {
				if pollEvent&unix.EPOLLIN > 0 {
					poll.loopAccept(eventFd)
				}
				continue
			}

			if pollEvent&unix.EPOLLERR > 0 || pollEvent&unix.EPOLLHUP > 0 {
				log.Info("rcv connection epollerr or epollhup event:%v, close the socket:%d", pollEvent, eventFd)
				poll.loopError(eventFd)
				continue
			}

			if int(pollEvent) & ^int(unix.EPOLLIN) & ^int(unix.EPOLLOUT) > 0 {
				log.Info("rcv unexpected event, close the socket : %d, events :%v", eventFd, pollEvent)
				poll.loopError(eventFd)
				continue
			}

			if pollEvent&unix.EPOLLIN > 0 {
				poll.loopRead(eventFd)
			}
			if pollEvent&unix.EPOLLOUT > 0 {
				poll.loopWrite(eventFd)
			}

		}
	}

	return nil
}

func (poll *Poll) loopAccept(fd int) error {
	nfd, sa, err := unix.Accept(fd)
	if err != nil {
		if err == unix.EAGAIN {
			log.Debug("loopAccept accept return EAGAIN, fd:%d", fd)
			return nil
		}
		return err
	}

	if err := unix.SetNonblock(nfd, true); err != nil {
		log.Error("loop accept set socket non block option error: %v", err)
		return err
	}

	if poll.netcore.socketSendBufferSize > 0 {
		err = unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_RCVBUF, poll.netcore.socketSendBufferSize)
		if err != nil {
			log.Error("loop accept set socket send buffer size option error: %v", err)
			return err
		}
	}

	if poll.netcore.socketRcvBufferSize > 0 {
		err = unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_RCVBUF, poll.netcore.socketRcvBufferSize)
		if err != nil {
			log.Error("loop accept set socket rcv buffer size option error: %v", err)
			return err
		}
	}

	tcpNodelay := 0
	if poll.netcore.socketTcpNoDelay {
		tcpNodelay = 1
	}
	err = unix.SetsockoptInt(fd, unix.IPPROTO_TCP, unix.TCP_NODELAY, tcpNodelay)
	if err != nil {
		log.Error("loop accept set socket tcp_nodelay option error: %v", err)
		return err
	}

	peerHost := ""
	peerPort := 0
	switch sa.(type) {
	case *unix.SockaddrInet4:
		sa4 := sa.(*unix.SockaddrInet4)
		peerHost = fmt.Sprintf("%d.%d.%d.%d", int(sa4.Addr[0]), int(sa4.Addr[1]), int(sa4.Addr[2]), int(sa4.Addr[3]))
		peerPort = sa4.Port
		log.Info("accept connection  fd : %d, remote_addr: %v, remote_port :%v",
			nfd, peerHost, peerPort)
	default:
		log.Info("accept connection  fd : %d, sa : %+v", nfd, sa)
	}

	// add connection to pool
	conn := NewNetConn(poll.netcore.socketSendBufferSize, poll.netcore.socketRcvBufferSize)
	conn.fd = nfd
	conn.state = int32(ConnStateConnected)
	conn.peerHost = peerHost
	conn.peerPort = peerPort

	allocPollIndex := poll.netcore.loadBalance.AllocConnection(conn.sessionId)
	allocPoll := poll.netcore.polls[allocPollIndex]
	//log.Debug("alloc accepted connection to poll, sessionId:%v, pollIndex:%v", conn.sessionId, allocPollIndex)

	param := []interface{}{allocPoll, conn}
	taskFunc := func(param interface{}) error {
		allocPoll := param.([]interface{})[0].(*Poll)
		conn := param.([]interface{})[1].(*NetConn)

		allocPoll.addConnection(conn)
		err = allocPoll.addRead(conn.fd)
		if err != nil {
			return err
		}
		for _, h := range allocPoll.eventHandler {
			h.OnAccept(conn)
		}
		return err
	}
	task := NewEventTask(taskFunc, param)
	allocPoll.wake(task)

	return nil
}

func (poll *Poll) loopRead(fd int) error {
	//log.Debug("connection trigger read event, pollFd:%v, fd:%v", poll.pollFd, fd)
	c := poll.getConnectionByFd(fd)
	if c == nil {
		log.Error("loop read connection not found, fd: %v", fd)
		return errors.New("loop read connection not found")
	}

	bReadFinish := false // case ring buffer full but not read all from socket
	for {
		head, tail := c.rcvBuff.PeekFreeAll()
		totalRead := 0

		for _, buff := range [2][]byte{head, tail} {
			if len(buff) <= 0 {
				break
			}
			n, err := unix.Read(fd, buff)
			if err != nil {
				if err == unix.EAGAIN {
					log.Debug("loopRead read return EAGAIN, fd:%d", fd)
					bReadFinish = true
					break
				}

				log.Error("connnection read error : %s, force close the socket :%d", err, fd)
				poll.close(fd)
				return fmt.Errorf("connnection read error : %s, fd : %d", err, fd)
			}

			if n <= 0 {
				log.Error("connnection read error length : %d, force close the socket :%d", n, fd)
				poll.close(fd)
				return fmt.Errorf("connnection read error length : %d, fd : %d", n, fd)
			}

			totalRead += n

			if n < len(buff) {
				bReadFinish = true
				break
			}
		}

		if totalRead > 0 {
			c.rcvBuff.Foward(totalRead)
			for {
				var in interface{} = nil
				codecSize := len(poll.netcore.codecs)
				for i := codecSize - 1; i >= 0; i-- {
					codec := poll.netcore.codecs[i]
					out, bChain, err := codec.Decode(c, in)
					in = out
					if err != nil {
						log.Error("loop read decode msg error:%s, fd:%d", err, c.fd)
						poll.close(fd)
						return err
					}
					if !bChain {
						break
					}
				}

				msg := in
				if msg != nil {
					//log.Debug("session:%v rcv msg : %v", c.sessionId, msg)
					for _, h := range poll.netcore.eventHandlers {
						h.OnRcvMsg(c, msg)
					}
				} else {
					break
				}
			}
		}

		if bReadFinish {
			break
		} else {
			//bReadFinish = true // just try again once
			oldCap := c.rcvBuff.Cap()
			c.rcvBuff.Grow(oldCap + oldCap/2)
			log.Info("loop read grow rcv buffer capcity from %v to %v, fd:%v", oldCap, c.rcvBuff.Cap(), fd)
			continue
		}

	}

	if !c.sendBuff.IsEmpty() {
		poll.modRead(fd)
	} else {
		poll.modReadWrite(fd)
	}
	return nil
}

func (poll *Poll) loopWrite(fd int) error {
	//log.Debug("connection trigger write event, pollFd:%v, eventFd:%v", poll.pollFd, fd)

	conn := poll.getConnectionByFd(fd)
	if conn == nil {
		return errors.New("connection not found")
	}

	// in process connecting succeed
	if conn.state == int32(ConnStateConnecting) {
		log.Info("connect to peer server success, peerHost:%v, peerPort:%v, sessionId:%v, fd:%v",
			conn.peerHost, conn.peerPort, conn.sessionId, conn.fd)

		conn.state = int32(ConnStateConnected)
		poll.netcore.removeWaitConn(conn.sessionId)
		for _, h := range poll.eventHandlers {
			h.OnConnect(conn, nil)
		}
	}

	if conn.sendBuff.IsEmpty() {
		poll.modRead(fd)
		return nil
	}

	head, tail := conn.sendBuff.PeekAll()
	totalWrite := 0
	for _, b := range [2][]byte{head, tail} {
		if len(b) <= 0 {
			break
		}
		n, err := unix.Write(conn.fd, b)
		if err != nil {
			if err == unix.EAGAIN {
				poll.logger.LogDebug("loopWrite write return EAGAIN, fd:%d", fd)
				totalWrite += n
				break
			} else {
				poll.logger.LogError("poll write error : %v", err)
				conn.sendBuff.Reset()
				poll.close(conn.fd)
				return err
			}
		} else {
			if n != len(b) {
				poll.logger.LogError("poll write return len error, write len:%d, data len:%d", n, len(b))
				conn.sendBuff.Reset()
				poll.close(conn.fd)
				return errors.New("write return length error")
			}
			totalWrite += n
		}
	}

	conn.sendBuff.Discard(totalWrite)

	if !conn.sendBuff.IsEmpty() {
		poll.modRead(fd)
	} else {
		poll.modReadWrite(fd)
	}

	return nil
}

func (poll *Poll) loopError(fd int) {
	conn := poll.getConnectionByFd(fd)
	if conn != nil && conn.state == int32(ConnStateConnecting) {
		log.Error("connect to peer server failed, peerHost:%v, peerPort:%v, sessionId:%v, fd:%v",
			conn.peerHost, conn.peerPort, conn.sessionId, conn.fd)
		for _, h := range poll.eventHandlers {
			h.OnConnect(conn, errors.New("epoll connect to peer server failed"))
		}
	}

	poll.close(fd)
}

func (poll *Poll) close(fd int) {
	conn := poll.getConnectionByFd(fd)
	if conn == nil {
		return
	}

	isConnected := (conn.state == int32(ConnStateConnected))
	conn.state = int32(ConnStateClosed)

	err := poll.modDetach(fd)
	if err != nil {
		log.Error("network close socket mod detach from epoll error : %s", err)
	}
	err = unix.Close(fd)
	if err != nil {
		log.Error("network close socket error : %s", err)
	}

	poll.removeConnectionByFd(fd)

	if isConnected {
		for _, h := range poll.eventHandlers {
			h.OnClosed(conn)
		}
	}

	if conn.IsClient() && conn.autoReconnect {
		poll.netcore.addWaitConn(conn)
	}
}

func (poll *Poll) addConnection(c *NetConn) {
	_, ok := poll.connMap[c.sessionId]
	if ok {
		log.Warn("add a existed fd to connection map, session_id:%v, fd:%v", c.sessionId, c.fd)
	}
	poll.connMap[c.sessionId] = c
	poll.connFdMap[c.fd] = c

	//log.Debug("poll index:%v, connMap:%+v, connFdMap:%+v", poll.pollIndex, poll.connMap, poll.connFdMap)
}

func (poll *Poll) removeConnection(sessionId int64) {
	c, ok := poll.connMap[sessionId]
	if !ok {
		log.Warn("remome connection not found, sessionId : %v", sessionId)
		return
	}
	delete(poll.connMap, c.sessionId)
	delete(poll.connFdMap, c.fd)

	//log.Debug("poll index:%v, connMap:%+v, connFdMap:%+v", poll.pollIndex, poll.connMap, poll.connFdMap)
}

func (poll *Poll) removeConnectionByFd(fd int) {
	c, ok := poll.connFdMap[fd]
	if !ok {
		log.Warn("remome connection not found, fd : %v", fd)
		return
	}
	delete(poll.connMap, c.sessionId)
	delete(poll.connFdMap, c.fd)

	//log.Debug("poll index:%v, connMap:%+v, connFdMap:%+v", poll.pollIndex, poll.connMap, poll.connFdMap)
}

func (poll *Poll) getConnection(sessionId int64) *NetConn {
	c, ok := poll.connMap[sessionId]
	if !ok {
		return nil
	}
	return c
}

func (poll *Poll) getConnectionByFd(fd int) *NetConn {
	c, ok := poll.connFdMap[fd]
	if !ok {
		return nil
	}
	return c
}

// addReadWrite ...
func (poll *Poll) addReadWrite(fd int) error {
	//log.Debug("addReadWrite fd : %v", fd)
	return unix.EpollCtl(poll.pollFd, unix.EPOLL_CTL_ADD, fd,
		&unix.EpollEvent{Fd: int32(fd),
			Events: unix.EPOLLET | unix.EPOLLIN | unix.EPOLLOUT,
		})
}

// addRead ...(listenFd只需要读取数据使用)
func (poll *Poll) addRead(fd int) error {
	//log.Debug("addRead fd : %v", fd)
	return unix.EpollCtl(poll.pollFd, unix.EPOLL_CTL_ADD, fd,
		&unix.EpollEvent{Fd: int32(fd),
			Events: unix.EPOLLET | unix.EPOLLIN,
		})
}

// // addWrite ...(用不到,fd任何时候都需要读取数据)
// func (poll *Poll) addWrite(fd int) error {
// 	log.Debug("addWrite fd : %v", fd)
// 	return unix.EpollCtl(poll.pollFd, unix.EPOLL_CTL_ADD, fd,
// 		&unix.EpollEvent{Fd: int32(fd),
// 			Events: unix.EPOLLET | unix.EPOLLOUT,
// 		})
// }

// modReadWrite ...
func (poll *Poll) modReadWrite(fd int) error {
	//log.Debug("modReadWrite fd : %v", fd)
	return unix.EpollCtl(poll.pollFd, unix.EPOLL_CTL_MOD, fd,
		&unix.EpollEvent{Fd: int32(fd),
			Events: unix.EPOLLIN | unix.EPOLLOUT,
		})
}

// modRead ...
func (poll *Poll) modRead(fd int) error {
	//log.Debug("modRead fd : %v", fd)
	return unix.EpollCtl(poll.pollFd, unix.EPOLL_CTL_MOD, fd,
		&unix.EpollEvent{Fd: int32(fd),
			Events: unix.EPOLLIN,
		})
}

// // modWrite ...(用不到,fd任何时候都需要读取数据)
// func (poll *Poll) modWrite(fd int) error {
// 	log.Debug("modWrite fd : %v", fd)
// 	return unix.EpollCtl(poll.pollFd, unix.EPOLL_CTL_MOD, fd,
// 		&unix.EpollEvent{Fd: int32(fd),
// 			Events: unix.EPOLLOUT,
// 		})
// }

// modDetach ...
func (poll *Poll) modDetach(fd int) error {
	//log.Debug("modDetach fd : %v", fd)
	return unix.EpollCtl(poll.pollFd, unix.EPOLL_CTL_DEL, fd,
		&unix.EpollEvent{Fd: int32(fd),
			Events: unix.EPOLLIN | unix.EPOLLOUT,
		})
}
