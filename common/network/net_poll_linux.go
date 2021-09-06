package network

import (
	"errors"
	"fmt"
	"golang.org/x/sys/unix"
	"goserver/common/log"
	"net"
	"sync/atomic"
	"unsafe"
)

// net poll, one poll => one goroutine
type Poll struct {
	netcore          *NetPollCore
	logger           log.Logger
	eventHandler     NetEventHandler
	pollFd           int
	wakeFd           int
	wfdBuf           []byte
	wakeEventQueue   EventTaskQueue
	connMap          map[int64]*Connection
	connFdMap        map[int]*Connection
	waitingConnFdMap map[int]*Connection
}

func NewNetPoll() *Poll {
	return &Poll{}
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
	err = poll.addRead(fd)
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
			poll.addWrite(fd)
			poll.addWaitingConnection(conn)
			return nil
		}
		poll.logger.LogError("tcpConnect connect peer endpoint error : %v, peerHost:%v, peerPort:%v", err, host, port)
		return err
	}

	//connected success directly
	conn.SetConnected(true)
	poll.addConnection(conn)
	poll.addRead(fd)

	poll.eventHandler.OnConnected(conn)

	return nil
}

func (poll *Poll) TcpSend(sessionId int64, buff []byte) error {
	c := poll.getConnection(sessionId)
	if c == nil {
		return errors.New("connection session not found")
	}

	c.sendBuff = append(c.sendBuff, buff...)
	n, err := unix.Write(c.fd, c.sendBuff)
	if err != nil {
		if err == unix.EAGAIN {
			if n < len(c.sendBuff) {
				c.sendBuff = c.sendBuff[n:]
				poll.modWrite(c.fd)
			} else {
				c.sendBuff = nil
			}
		} else {
			poll.logger.LogError("poll tcp send error : %v", err)
			poll.close(c.fd)
			return err
		}
	}
	c.sendBuff = c.sendBuff[:0]
	return nil
}

func (poll *Poll) wake(t *EventTask) error {
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
			pollEvent := events[i].Events

			poll.logger.LogDebug("loopEpollWait trigger eventfd : %d, wakeFd :%d, events :%v",
				eventFd, poll.wakeFd, pollEvent)

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
				poll.logger.LogInfo("rcv connection epollerr or epollhup event:%v, close the socket:%d", pollEvent, eventFd)
				poll.loopError(eventFd)
				continue
			}

			if pollEvent&unix.EPOLLIN > 0 {
				poll.loopRead(eventFd)
			}
			if pollEvent&unix.EPOLLOUT > 0 {
				poll.loopWrite(eventFd)
			}

			if int(pollEvent) & ^int(unix.EPOLLIN) & ^int(unix.EPOLLOUT) > 0 {
				poll.logger.LogInfo("rcv unexpected event, close the socket : %d, events :%v", eventFd, pollEvent)
				poll.loopError(eventFd)
			}
		}
	}

	return nil
}

func (poll *Poll) loopAccept(fd int) error {
	nfd, sa, err := unix.Accept(fd)
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
	poll.addConnection(conn)

	err = poll.addRead(nfd)
	if err != nil {
		return err
	}

	poll.eventHandler.OnAccept(conn)

	return nil
}

func (poll *Poll) loopRead(fd int) error {
	poll.logger.LogDebug("connection trigger read event, pollFd:%v, fd:%v", poll.pollFd, fd)

	packet := make([]byte, 1024)
	n, err := unix.Read(fd, packet)

	if err != nil {
		if err == unix.EAGAIN {
			poll.modReadWrite(fd)
			return nil
		}

		poll.logger.LogError("connnection read error : %s, force close the socket :%d", err, fd)
		poll.close(fd)
		return fmt.Errorf("connnection read error : %s, fd : %d", err, fd)
	}

	if n <= 0 {
		poll.logger.LogError("connnection read error length : %d, force close the socket :%d", n, fd)
		poll.close(fd)
		return fmt.Errorf("connnection read error length : %d, fd : %d", n, fd)
	}

	poll.logger.LogDebug("rcv buffer :%v", string(packet))

	poll.modRead(fd)
	return nil
}

func (poll *Poll) loopWrite(fd int) error {
	poll.logger.LogDebug("connection trigger write event, pollFd:%v, eventFd:%v", poll.pollFd, fd)

	// in process connecting succeed
	waitingConn := poll.getWaitingConnection(fd)
	if waitingConn != nil {
		poll.logger.LogInfo("connect to peer server success, peerHost:%v, peerPort:%v, sessionId:%v, fd:%v",
			waitingConn.peerHost, waitingConn.peerPort, waitingConn.sessionId, waitingConn.fd)
		waitingConn.SetConnected(true)
		poll.removeWaitingConnection(fd)
		poll.addConnection(waitingConn)
		poll.eventHandler.OnAccept(waitingConn)
		poll.addRead(fd)
		return nil
	}

	c := poll.getConnectionByFd(fd)
	if c == nil {
		return errors.New("connection not found")
	}

	n, err := unix.Write(c.fd, c.sendBuff)
	if err != nil {
		if err == unix.EAGAIN {
			if n < len(c.sendBuff) {
				c.sendBuff = c.sendBuff[n:]
				poll.modWrite(c.fd)
			} else {
				c.sendBuff = nil
			}
		} else {
			poll.logger.LogError("poll write error : %v", err)
			poll.close(c.fd)
			return err
		}
	}
	c.sendBuff = c.sendBuff[:0]

	return nil
}

func (poll *Poll) loopError(fd int) {

	waitingConn := poll.getWaitingConnection(fd)
	if waitingConn != nil {
		poll.logger.LogError("connect to peer server failed, peerHost:%v, peerPort:%v, sessionId:%v, fd:%v",
			waitingConn.peerHost, waitingConn.peerPort, waitingConn.sessionId, waitingConn.fd)
	}

	poll.close(fd)

	c := poll.getConnectionByFd(fd)
	if c != nil {
		poll.eventHandler.OnDisconnected(c)
	}
}

func (poll *Poll) close(fd int) {
	err := poll.modDetach(fd)
	if err != nil {
		poll.logger.LogError("network close socket mod detach from epoll error : %s", err)
	}
	err = unix.Close(fd)
	if err != nil {
		poll.logger.LogError("network close socket error : %s", err)
	}

	poll.removeConnectionByFd(fd)
	if poll.getWaitingConnection(fd) != nil {
		poll.removeWaitingConnection(fd)
	}
}

func (poll *Poll) addConnection(c *Connection) {
	_, ok := poll.connMap[c.sessionId]
	if ok {
		poll.logger.LogWarn("add a existed fd to connection map, session_id:%v, fd:%v", c.sessionId, c.fd)
	}
	poll.connMap[c.sessionId] = c
	poll.connFdMap[c.fd] = c
}

func (poll *Poll) removeConnection(sessionId int64) {
	c, ok := poll.connMap[sessionId]
	if !ok {
		poll.logger.LogWarn("remome connection not found, sessionId : %v", sessionId)
		return
	}
	delete(poll.connMap, c.sessionId)
	delete(poll.connFdMap, c.fd)
}

func (poll *Poll) removeConnectionByFd(fd int) {
	c, ok := poll.connFdMap[fd]
	if !ok {
		poll.logger.LogWarn("remome connection not found, fd : %v", fd)
		return
	}
	delete(poll.connMap, c.sessionId)
	delete(poll.connFdMap, c.fd)
}

func (poll *Poll) getConnection(sessionId int64) *Connection {
	c, ok := poll.connMap[sessionId]
	if !ok {
		return nil
	}
	return c
}

func (poll *Poll) getConnectionByFd(fd int) *Connection {
	c, ok := poll.connFdMap[fd]
	if !ok {
		return nil
	}
	return c
}

func (poll *Poll) addWaitingConnection(c *Connection) {
	_, ok := poll.waitingConnFdMap[c.fd]
	if ok {
		poll.logger.LogWarn("add a existed fd to waiting connection map, fd:%v", c.fd)
	}
	poll.waitingConnFdMap[c.fd] = c
}

func (poll *Poll) removeWaitingConnection(fd int) {
	_, ok := poll.waitingConnFdMap[fd]
	if !ok {
		poll.logger.LogWarn("remome waiting connection not found, fd : %v", fd)
		return
	}
	delete(poll.waitingConnFdMap, fd)
}

func (poll *Poll) getWaitingConnection(fd int) *Connection {
	c, ok := poll.waitingConnFdMap[fd]
	if !ok {
		return nil
	}
	return c
}

// addReadWrite ...
func (poll *Poll) addReadWrite(fd int) error {
	poll.logger.LogDebug("addReadWrite fd : %v", fd)
	return unix.EpollCtl(poll.pollFd, unix.EPOLL_CTL_ADD, fd,
		&unix.EpollEvent{Fd: int32(fd),
			Events: unix.EPOLLET | unix.EPOLLIN | unix.EPOLLOUT,
		})
}

// addRead ...
func (poll *Poll) addRead(fd int) error {
	poll.logger.LogDebug("addRead fd : %v", fd)
	return unix.EpollCtl(poll.pollFd, unix.EPOLL_CTL_ADD, fd,
		&unix.EpollEvent{Fd: int32(fd),
			Events: unix.EPOLLET | unix.EPOLLIN,
		})
}

// addWrite ...
func (poll *Poll) addWrite(fd int) error {
	poll.logger.LogDebug("addWrite fd : %v", fd)
	return unix.EpollCtl(poll.pollFd, unix.EPOLL_CTL_ADD, fd,
		&unix.EpollEvent{Fd: int32(fd),
			Events: unix.EPOLLET | unix.EPOLLOUT,
		})
}

// modReadWrite ...
func (poll *Poll) modReadWrite(fd int) error {
	poll.logger.LogDebug("modReadWrite fd : %v", fd)
	return unix.EpollCtl(poll.pollFd, unix.EPOLL_CTL_MOD, fd,
		&unix.EpollEvent{Fd: int32(fd),
			Events: unix.EPOLLIN | unix.EPOLLOUT,
		})
}

// modRead ...
func (poll *Poll) modRead(fd int) error {
	poll.logger.LogDebug("modRead fd : %v", fd)
	return unix.EpollCtl(poll.pollFd, unix.EPOLL_CTL_MOD, fd,
		&unix.EpollEvent{Fd: int32(fd),
			Events: unix.EPOLLIN,
		})
}

// modWrite ...
func (poll *Poll) modWrite(fd int) error {
	poll.logger.LogDebug("modWrite fd : %v", fd)
	return unix.EpollCtl(poll.pollFd, unix.EPOLL_CTL_MOD, fd,
		&unix.EpollEvent{Fd: int32(fd),
			Events: unix.EPOLLOUT,
		})
}

// modDetach ...
func (poll *Poll) modDetach(fd int) error {
	poll.logger.LogDebug("modDetach fd : %v", fd)
	return unix.EpollCtl(poll.pollFd, unix.EPOLL_CTL_DEL, fd,
		&unix.EpollEvent{Fd: int32(fd),
			Events: unix.EPOLLIN | unix.EPOLLOUT,
		})
}
