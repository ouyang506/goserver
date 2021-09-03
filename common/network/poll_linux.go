package network

import (
	"fmt"
	"golang.org/x/sys/unix"
	"goserver/common/log"
	"net"
	"sync"
	"sync/atomic"
	"unsafe"
)

// NetPollCore implements the NetworkCore interface
type NetPollCore struct {
	logger       log.Logger
	numLoops     int
	loadBalance LoadBalance
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

	poll.Wake(task)
	return nil
}

// implement network core TcpConnect
func (netcore *NetPollCore) TcpConnect(host string, port int) error {
	poll := netcore.loadBalance.AllocConnection

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
	poll.Wake(task)

	return nil
}

// implement network core TcpSend
func (netcore *NetPollCore) TcpSend(fd int, []byte) error{
	
	return nil
}


// NetPollCore loop
func (netcore *NetPollCore) startLoop() error {
	for i := 0; i < netcore.numLoops; i++ {
		pollFd, err := unix.EpollCreate1(unix.EPOLL_CLOEXEC)
		if err != nil {
			netcore.logger.LogError("EpollCreate1 error : %v", err)
			return err
		}

		poll := &Poll{}
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
		poll.AddRead(poll.wakeFd)

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

func (netcore *NetPollCore) AddConnection(fd int, c *Connection) {
	netcore.connMapMutex.Lock()
	defer netcore.connMapMutex.Unlock()
	_, ok := netcore.connMap[fd]
	if ok {
		netcore.logger.LogWarn("add a existed fd to connection map, fd : %v", fd)
	}
	netcore.connMap[fd] = c
}

func (netcore *NetPollCore) RemoveConnection(fd int) {
	netcore.connMapMutex.Lock()
	defer netcore.connMapMutex.Unlock()
	_, ok := netcore.connMap[fd]
	if !ok {
		netcore.logger.LogWarn("remome connection not found, fd : %v", fd)
		return
	}
	delete(netcore.connMap, fd)
}

func (netcore *NetPollCore) GetConnection(fd int) *Connection {
	netcore.connMapMutex.Lock()
	defer netcore.connMapMutex.Unlock()
	c, ok := netcore.connMap[fd]
	if !ok {
		return nil
	}
	return c
}

// one poll => one goroutine
type Poll struct {
	netcore        *NetPollCore
	logger         log.Logger
	eventHandler   NetEventHandler
	pollFd         int
	wakeFd         int
	wfdBuf         []byte
	wakeEventQueue EventTaskQueue
}

func (poll *Poll) tcpListen(host string, port int) (int, error) {
	poll.logger.LogDebug("poll tcp listen called , pollFd : %v", poll.pollFd)

	tcpAddr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		poll.logger.LogError("tcpListen ResolveTCPAddr error : %v", err)
		return 0, err
	}

	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM|unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC, unix.IPPROTO_TCP)
	if err != nil {
		poll.logger.LogError("tcpListen create socket error : %v", err)
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
		poll.logger.LogError("tcpListen set reuse addr error : %v", err)
		return 0, err
	}

	err = unix.Bind(fd, sa4)
	if err != nil {
		poll.logger.LogError("tcpListen bind error : %v", err)
		return 0, err
	}

	err = unix.Listen(fd, unix.SOMAXCONN)
	if err != nil {
		poll.logger.LogError("tcpListen listen error : %v", err)
		return 0, err
	}

	atomic.StoreInt32(&poll.netcore.listenFd, int32(fd))

	// set listen fd nonblock
	err = unix.SetNonblock(fd, true)
	if err != nil {
		poll.logger.LogError("tcpListen set nonblock error : %v", err)
		return 0, err
	}

	// add listen fd to epoll
	err = poll.AddRead(fd)
	if err != nil {
		poll.logger.LogError("tcpListen add read error : %v", err)
		return 0, err
	}

	return fd, nil
}

func (poll *Poll) tcpConnect(host string, port int) error {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		poll.logger.LogError("tcpConnect ResolveTCPAddr error : %v", err)
		return err
	}

	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM|unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC, unix.IPPROTO_TCP)
	if err != nil {
		poll.logger.LogError("tcpConnect create socket error : %v", err)
		return err
	}

	sa4 := &unix.SockaddrInet4{Port: tcpAddr.Port}
	if tcpAddr.IP != nil {
		if len(tcpAddr.IP) == 16 {
			copy(sa4.Addr[:], tcpAddr.IP[12:16]) // copy last 4 bytes of slice to array
		} else {
			copy(sa4.Addr[:], tcpAddr.IP) // copy all bytes of slice to array
		}
	}

	conn := NewConnection()
	conn.SetFd(fd)
	conn.SetPeerAddr(host, port)

	err = unix.Connect(fd, sa4)
	if err != nil {
		if err == unix.EINPROGRESS {
			poll.logger.LogDebug("tcpConnect connect peer endpoint in progress, host:%v, port %v", host, port)
			poll.AddWrite(fd)
			//poll.AddWaitConnection(fd, conn)
			return nil
		}
		poll.logger.LogError("tcpConnect connect peer endpoint error : %v, peerHost:%v, peerPort:%v", err, host, port)
		return err
	}

	//connected success directly
	conn.SetConnected(true)
	poll.AddRead(fd)

	poll.eventHandler.OnConnected(conn)

	return nil
}

func (poll *Poll) loopEpollWait() error {
	events := make([]unix.EpollEvent, 1024)
	for {
		n, err := unix.EpollWait(poll.pollFd, events, 100)
		if err != nil && err != unix.EINTR {
			poll.logger.LogError("loop epoll wait error : %v", err)
			return err
		}

		for i := 0; i < n; i++ {
			eventFd := int(events[i].Fd)

			poll.logger.LogDebug("loopEpollWait trigger eventfd : %d, wakeFd :%d, events :%v",
				eventFd, poll.wakeFd, events[i].Events)

			if eventFd == poll.wakeFd {
				_, _ = unix.Read(eventFd, poll.wfdBuf)
				for task := poll.wakeEventQueue.Dequeue(); task != nil; task = poll.wakeEventQueue.Dequeue() {
					task.eventFunc(task.param)
				}
			} else if eventFd == int(atomic.LoadInt32(&poll.netcore.listenFd)) {
				if events[i].Events&unix.EPOLLIN > 0 {
					poll.accept(eventFd)
				}
			} else {
				if events[i].Events&unix.EPOLLERR > 0 {
					poll.logger.LogInfo("rcv connection epollerr event, close the socket : %d", eventFd)
					poll.close(eventFd)
				} else {
					if events[i].Events&unix.EPOLLIN > 0 {
						poll.read(eventFd)
					}
					if events[i].Events&unix.EPOLLOUT > 0 {
						poll.write(eventFd)
					}

					if int(events[i].Events) & ^int(unix.EPOLLIN) & ^int(unix.EPOLLOUT) > 0 {
						poll.logger.LogInfo("rcv unexpected event, close the socket : %d, events :%v", eventFd, events[i].Events)
						poll.close(eventFd)
					}
				}
			}
		}
	}

	return nil
}

func (poll *Poll) accept(eventFd int) error {
	nfd, sa, err := unix.Accept(eventFd)
	if err != nil {
		if err == unix.EAGAIN {
			return nil
		}
		return err
	}
	if err := unix.SetNonblock(nfd, true); err != nil {
		return err
	}
	peerHost := ""
	peerPort := 0
	switch sa.(type) {
	case *unix.SockaddrInet4:
		sa4 := sa.(*unix.SockaddrInet4)
		peerHost = fmt.Sprintf("%d.%d.%d.%d", int(sa4.Addr[0]), int(sa4.Addr[1]), int(sa4.Addr[2]), int(sa4.Addr[3]))
		peerPort = sa4.Port
		poll.logger.LogInfo("accept connection  fd : %d, remote_addr: %v, remote_port :%v",
			nfd, peerHost, peerPort)
		break
	default:
		poll.logger.LogInfo("accept connection  fd : %d, sa : %+v", nfd, sa)
	}

	// add connection to pool
	conn := NewConnection()
	conn.SetFd(nfd)
	conn.SetConnected(true)
	conn.SetPeerAddr(peerHost, peerPort)
	//netcore.AddConnection(eventFd, conn)

	err = poll.AddRead(nfd)
	if err != nil {
		return err
	}

	poll.eventHandler.OnConnected(conn)

	return nil
}

func (poll *Poll) read(eventFd int) error {
	poll.logger.LogDebug("connection trigger read event, pollFd:%v, eventFd:%v", poll.pollFd, eventFd)

	packet := make([]byte, 1024)
	n, err := unix.Read(eventFd, packet)

	if err != nil {
		if err == unix.EAGAIN {
			poll.ModReadWrite(eventFd)
			return nil
		}

		poll.logger.LogError("connnection read error : %s, force close the socket :%d", err, eventFd)
		poll.close(eventFd)
		return fmt.Errorf("connnection read error : %s, fd : %d", err, eventFd)
	}

	if n <= 0 {
		poll.logger.LogError("connnection read error length : %d, force close the socket :%d", n, eventFd)
		poll.close(eventFd)
		return fmt.Errorf("connnection read error length : %d, fd : %d", n, eventFd)
	}

	poll.logger.LogDebug("rcv buffer :%v", string(packet))

	poll.ModRead(eventFd)
	return nil
}

func (poll *Poll) write(eventFd int) error {
	poll.logger.LogDebug("connection trigger write event, pollFd:%v, eventFd:%v", poll.pollFd, eventFd)

	return nil
}

func (poll *Poll) close(eventFd int) {
	err := poll.ModDetach(eventFd)
	if err != nil {
		poll.logger.LogError("network close socket mod detach from epoll error : %s", err)
	}
	err = unix.Close(eventFd)
	if err != nil {
		poll.logger.LogError("network close socket error : %s", err)
	}

	//netcore.RemoveConnection(eventFd)
}

func (poll *Poll) Wake(t *EventTask) error {
	u := int64(1)
	b := (*(*[8]byte)(unsafe.Pointer(&u)))[:]

	poll.wakeEventQueue.Enqueue(t)

	for {
		_, err := unix.Write(poll.wakeFd, b)
		if err != nil {
			poll.logger.LogDebug("poll wake error : %v, %v, %v", err, unix.EINTR, unix.EAGAIN)
			if err == unix.EINTR || err == unix.EAGAIN {
				continue
			} else {
				return err
			}
		} else {
			poll.logger.LogDebug("send wake event to wakeFd:%v", poll.wakeFd)
			break
		}
	}
	return nil
}

// AddReadWrite ...
func (poll *Poll) AddReadWrite(fd int) error {
	poll.logger.LogDebug("AddReadWrite fd : %v", fd)
	return unix.EpollCtl(poll.pollFd, unix.EPOLL_CTL_ADD, fd,
		&unix.EpollEvent{Fd: int32(fd),
			Events: unix.EPOLLET | unix.EPOLLIN | unix.EPOLLOUT,
		})
}

// AddRead ...
func (poll *Poll) AddRead(fd int) error {
	poll.logger.LogDebug("AddRead fd : %v", fd)
	return unix.EpollCtl(poll.pollFd, unix.EPOLL_CTL_ADD, fd,
		&unix.EpollEvent{Fd: int32(fd),
			Events: unix.EPOLLET | unix.EPOLLIN,
		})
}

// AddWrite ...
func (poll *Poll) AddWrite(fd int) error {
	poll.logger.LogDebug("AddWrite fd : %v", fd)
	return unix.EpollCtl(poll.pollFd, unix.EPOLL_CTL_ADD, fd,
		&unix.EpollEvent{Fd: int32(fd),
			Events: unix.EPOLLET | unix.EPOLLOUT,
		})
}

// ModRead ...
func (poll *Poll) ModRead(fd int) error {
	poll.logger.LogDebug("ModRead fd : %v", fd)
	return unix.EpollCtl(poll.pollFd, unix.EPOLL_CTL_MOD, fd,
		&unix.EpollEvent{Fd: int32(fd),
			Events: unix.EPOLLIN,
		})
}

// ModReadWrite ...
func (poll *Poll) ModReadWrite(fd int) error {
	poll.logger.LogDebug("ModReadWrite fd : %v", fd)
	return unix.EpollCtl(poll.pollFd, unix.EPOLL_CTL_MOD, fd,
		&unix.EpollEvent{Fd: int32(fd),
			Events: unix.EPOLLIN | unix.EPOLLOUT,
		})
}

// ModDetach ...
func (poll *Poll) ModDetach(fd int) error {
	poll.logger.LogDebug("ModDetach fd : %v", fd)
	return unix.EpollCtl(poll.pollFd, unix.EPOLL_CTL_DEL, fd,
		&unix.EpollEvent{Fd: int32(fd),
			Events: unix.EPOLLIN | unix.EPOLLOUT,
		})
}
