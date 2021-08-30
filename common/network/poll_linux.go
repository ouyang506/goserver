package network

import (
	"fmt"
	"golang.org/x/sys/unix"
	"goserver/common/log"
	"net"
	"sync/atomic"
)

const(
	WakeEventTaskTypeTcpListen = 1
	WakeEventTaskTypeTcpConnect = 2
)

type NetPollCore struct {
	logger       log.LoggerInterface
	numLoops     int
	listenFd     int32
	connMap      map[int]*Connection // all established connections
	connMapMutex sync.Mutex
	polls        []*Poll
}

type Poll struct {
	netcore     *NetPollCore
	logger      log.LoggerInterface
	pollFd      int
	wakeFd 		int
	wakeEventQueue EventTaskQueue
}


func newNetworkCore(numLoops int, logger log.LoggerInterface) *NetCore {
	netcore := &NetPollCore{}
	netcore.numLoops = numLoops
	netcore.logger = logger
	netcore.startLoop()
	return netcore
}

// implement network core TcpListen
func (netcore *NetPollCore) TcpListen(host string, port int) {
	poll := netcore.polls[0]

	task := &EventTask{}
	task.taskType = WakeEventTaskTypeTcpListen
	task.taskFunc = func(interface{}) error{
		err := poll.tcpListen(host, port)
		if err != nil {
			poll.logger.LogError("tcp listen at %v:%v error : %s", host, port, err)
		} else {
			poll.logger.LogInfo("start tcp listen at %v:%v, listen fd: %v", host, port, mgr.listenFd)
		}
		return err
	}

	poll.Wake(task)
	return nil
}

// implement network core TcpConnect
func (netcore *NetPollCore) TcpConnect(host string, port int) {
	err := netcore.tcpConnect(host, port)
	if err != nil {
		netcore.logger.LogError("tcp connect error, host:%v, port:%v, error : %s", host, port, err)
	} else {
		netcore.logger.LogError("tcp connect ok, host:%v, port:%v", host, port)
	}
	return err
}

func (poll *Poll) tcpListen(host string, port int) error {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		poll.logger.LogError("tcpListen ResolveTCPAddr error : %v", err)
		return err
	}

	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM|unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC, unix.IPPROTO_TCP)
	if err != nil {
		poll.logger.LogError("tcpListen create socket error : %v", err)
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

	err = unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
	if err != nil {
		poll.logger.LogError("tcpListen set reuse addr error : %v", err)
		return err
	}

	err = unix.Bind(fd, sa4)
	if err != nil {
		poll.logger.LogError("tcpListen bind error : %v", err)
		return err
	}

	err = unix.Listen(fd, unix.SOMAXCONN)
	if err != nil {
		poll.logger.LogError("tcpListen listen error : %v", err)
		return err
	}

	atomic.StoreInt32(&netcore.listenFd, fd)

	// set listen fd nonblock
	err = unix.SetNonblock(fd, true)
	if err != nil {
		netcore.logger.LogError("tcpListen set nonblock error : %v", err)
		return err
	}

	// add listen fd to epoll
	err = netcore.AddRead(netcore.polls[0].pollFd, fd)
	if err != nil {
		netcore.logger.LogError("tcpListen add read error : %v", err)
		return err
	}

	return nil
}

func (netcore *NetPollCore) tcpConnect(host string, port int) error {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		netcore.logger.LogError("tcpConnect ResolveTCPAddr error : %v", err)
		return err
	}

	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM|unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC, unix.IPPROTO_TCP)
	if err != nil {
		netcore.logger.LogError("tcpConnect create socket error : %v", err)
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

	poll := netcore.polls[fd%netcore.numLoops]

	err = unix.Connect(fd, sa4)
	if err != nil {
		if err == unix.EINPROGRESS {
			netcore.logger.LogDebug("tcpConnect connect peer endpoint in progress, host:%v, port %v", host, port)
			poll.AddWrite(mgr.pollFds[fd%mgr.numLoops], fd)
			netcore.AddWaitConnection(fd, conn)
			return nil
		}
		netcore.logger.LogError("tcpConnect connect peer endpoint error : %v, peerHost:%v, peerPort:%v", err, host, port)
		return err
	}

	//connected success directly
	conn.SetConnected(true)
	netcore.AddConnection(fd, conn)
	netcore.AddRead(netcore.pollFds[], fd)

	return nil
}

func (mgr *NetPollCore) AddConnection(fd int, c *Connection) {
	mgr.connMapMutex.Lock()
	defer mgr.connMapMutex.Unlock()
	_, ok := mgr.connMap[fd]
	if ok {
		mgr.logger.LogWarn("add a existed fd to connection map, fd : %v", fd)
	}
	mgr.connMap[fd] = c
}

func (mgr *NetPollCore) RemoveConnection(fd int) {
	mgr.connMapMutex.Lock()
	defer mgr.connMapMutex.Unlock()
	_, ok := mgr.connMap[fd]
	if !ok {
		mgr.logger.LogWarn("remome connection not found, fd : %v", fd)
		return
	}
	delete(mgr.connMap, fd)
}

func (mgr *NetPollCore) GetConnection(fd int) *Connection {
	mgr.connMapMutex.Lock()
	defer mgr.connMapMutex.Unlock()
	c, ok := mgr.connMap[fd]
	if !ok {
		return nil
	}
	return c
}

func (netcore *NetPollCore) startLoop() error {
	for i := 0; i < mgr.numLoops; i++ {
		pollFd, err := unix.EpollCreate1(unix.EPOLL_CLOEXEC)
		if err != nil {
			return err
		}

		poll := &Poll{}
		poll.netcore = netcore
		poll.logger = net.logger
		poll.pollFd = pollFd

		poll.wakeFd, err = unix.Eventfd(0, unix.EFD_NONBLOCK|unix.EFD_CLOEXEC)
		if err != nil{
			return err
		}
		
		netcore.polls = append(netcore.polls, poll)
	}

	// start goroutines for loop
	for i := 0; i < mgr.numLoops; i++ {
		go func(poll *Poll) {
			poll.loopEpollWait()
		}(mgr.polls[i])
	}
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
			if eventFd == mgr.listenFd {
				if events[i].Events&unix.EPOLLIN > 0 {
					mgr.accept(eventFd)
				}
			} else {
				if events[i].Events&unix.EPOLLERR > 0 {
					mgr.logger.LogInfo("rcv connection epollerr event, close the socket : %d", eventFd)
					mgr.close(epollFd, eventFd)
				} else {
					if events[i].Events&unix.EPOLLIN > 0 {
						mgr.read(epollFd, eventFd)
					}
					if events[i].Events&unix.EPOLLOUT > 0 {
						mgr.write(epollFd, eventFd)
					}
				}
			}
		}
	}

	return nil
}

func (mgr *NetPollCore) accept(eventFd int) error {
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
		mgr.logger.LogInfo("accept connection  fd : %d, remote_addr: %v, remote_port :%v",
			nfd, peerHost, peerPort)
		break
	default:
		mgr.logger.LogInfo("accept connection  fd : %d, sa : %+v", nfd, sa)
	}

	// add connection to pool
	conn := NewConnection()
	conn.SetFd(fd)
	conn.SetConnected(true)
	conn.SetPeerAddr(peerHost, peerPort)
	mgr.AddConnection(fd, conn)

	allocEpollFd := mgr.pollFds[nfd%mgr.numLoops]
	err = mgr.AddRead(allocEpollFd, nfd)
	if err != nil {
		return err
	}

	return nil
}

func (mgr *NetPollCore) read(epollFd int, eventFd int) error {
	mgr.logger.LogDebug("connection trigger read event, epollFd:%v, eventFd:%v", epollFd, eventFd)

	packet := make([]byte, 1024)
	n, err := unix.Read(eventFd, packet)

	if err != nil {
		if err == unix.EAGAIN {
			mgr.ModReadWrite(epollFd, eventFd)
			return nil
		}

		mgr.logger.LogError("connnection read error : %s, force close the socket :%d", err, eventFd)
		mgr.close(epollFd, eventFd)
		return fmt.Errorf("connnection read error : %s, fd : %d", err, eventFd)
	}

	if n <= 0 {
		mgr.logger.LogError("connnection read error length : %d, force close the socket :%d", n, eventFd)
		mgr.close(epollFd, eventFd)
		return fmt.Errorf("connnection read error length : %d, fd : %d", n, eventFd)
	}

	mgr.logger.LogDebug("rcv buffer :%v", string(packet))

	mgr.ModRead(epollFd, eventFd)
	return nil
}

func (mgr *NetPollCore) write(epollFd int, eventFd int) error {
	mgr.logger.LogDebug("connection trigger write event, epollFd:%v, eventFd:%v", epollFd, eventFd)

	return nil
}

func (mgr *NetPollCore) close(epollFd int, eventFd int) {
	err := mgr.ModDetach(epollFd, eventFd)
	if err != nil {
		mgr.logger.LogError("network close socket mod detach from epoll error : %s", err)
	}
	err = unix.Close(eventFd)
	if err != nil {
		mgr.logger.LogError("network close socket error : %s", err)
	}

	mgr.RemoveConnection(eventFd)
}

func (poll *Poll) Wake(t *EventTask) error {

	b  := (*(*[8]byte)(unsafe.Pointer(&u)))[:]

	poll.taskQueue.Enqueue(t)

	for _, err = unix.Write(poll.wakeFd, b); 
		err == unix.EINTR || err == unix.EAGAIN;
		_, err = unix.Write(poll.wakeFd, b) {
	}
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
