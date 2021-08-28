package network

import (
	"fmt"
	"golang.org/x/sys/unix"
	"net"
)

func (mgr *NetworkMgr) openPoll() error {
	for i := 0; i < mgr.numLoops; i++ {
		epollFd, err := unix.EpollCreate1(unix.EPOLL_CLOEXEC)
		if err != nil {
			return err
		}
		mgr.pollFds = append(mgr.pollFds, epollFd)
	}

	// start goroutines for loop
	for i := 0; i < mgr.numLoops; i++ {
		go func(epollFd int) {
			mgr.loopEpollWait(epollFd)
		}(mgr.pollFds[i])
	}
	return nil
}

func (mgr *NetworkMgr) closePoll() error {
	for _, fd := range mgr.pollFds {
		err := unix.Close(fd)
		if err != nil {
			return err
		}
	}
	return nil
}

func (mgr *NetworkMgr) loopEpollWait(epollFd int) error {
	events := make([]unix.EpollEvent, 1024)
	for {
		n, err := unix.EpollWait(epollFd, events, 100)
		if err != nil && err != unix.EINTR {
			mgr.logger.LogError("loop epoll wait error : %v", err)
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

func (mgr *NetworkMgr) accept(eventFd int) error {
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
	switch sa.(type) {
	case *unix.SockaddrInet4:
		sa4 := sa.(*unix.SockaddrInet4)
		mgr.logger.LogInfo("accept connection  fd : %d, remote_addr: %d.%d.%d.%d, remote_port :%v",
			nfd, int(sa4.Addr[0]), int(sa4.Addr[1]), int(sa4.Addr[2]), int(sa4.Addr[3]), sa4.Port)
		break
	default:
		mgr.logger.LogInfo("accept connection  fd : %d, sa : %+v", nfd, sa)
	}

	allocEpollFd := mgr.pollFds[nfd%mgr.numLoops]
	err = mgr.AddRead(allocEpollFd, nfd)
	if err != nil {
		return err
	}

	return nil
}

func (mgr *NetworkMgr) read(epollFd int, eventFd int) error {
	mgr.logger.LogDebug("connection trigger read event")

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

	if n < 0 {
		mgr.logger.LogError("connnection read error length : %d, force close the socket :%d", n, eventFd)
		mgr.close(epollFd, eventFd)
		return fmt.Errorf("connnection read error length : %d, fd : %d", n, eventFd)
	}

	if n > 0 {
		mgr.logger.LogDebug("rcv buffer :%v", string(packet))
	}
	mgr.ModReadWrite(epollFd, eventFd)
	return nil
}

func (mgr *NetworkMgr) write(epollFd int, eventFd int) error {
	mgr.logger.LogDebug("connection trigger write event")
	return nil
}

func (mgr *NetworkMgr) close(epollFd int, eventFd int) {
	err := mgr.ModDetach(epollFd, eventFd)
	if err != nil {
		mgr.logger.LogError("network close socket mod detach from epoll error : %s", err)
	}
	err = unix.Close(eventFd)
	if err != nil {
		mgr.logger.LogError("network close socket error : %s", err)
	}
}

func (mgr *NetworkMgr) tcpListen(host string, port int) error {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		mgr.logger.LogError("tcpListen ResolveTCPAddr error : %v", err)
		return err
	}

	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM|unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC, unix.IPPROTO_TCP)
	if err != nil {
		mgr.logger.LogError("tcpListen create socket error : %v", err)
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

	err = unix.Bind(fd, sa4)
	if err != nil {
		mgr.logger.LogError("tcpListen bind error : %v", err)
		return err
	}

	unix.Listen(fd, unix.SOMAXCONN)
	if err != nil {
		mgr.logger.LogError("tcpListen listen error : %v", err)
		return err
	}

	mgr.listenFd = fd

	// set listen fd nonblock
	err = unix.SetNonblock(mgr.listenFd, true)
	if err != nil {
		mgr.logger.LogError("tcpListen set nonblock error : %v", err)
		return err
	}

	// add listen fd to epoll
	err = mgr.AddRead(mgr.pollFds[0], mgr.listenFd)
	if err != nil {
		mgr.logger.LogError("tcpListen add read error : %v", err)
		return err
	}

	return nil
}

// AddReadWrite ...
func (p *NetworkMgr) AddReadWrite(epollFd int, fd int) error {
	return unix.EpollCtl(epollFd, unix.EPOLL_CTL_ADD, fd,
		&unix.EpollEvent{Fd: int32(fd),
			Events: unix.EPOLLET | unix.EPOLLIN | unix.EPOLLOUT,
		})
}

// AddRead ...
func (p *NetworkMgr) AddRead(epollFd int, fd int) error {
	return unix.EpollCtl(epollFd, unix.EPOLL_CTL_ADD, fd,
		&unix.EpollEvent{Fd: int32(fd),
			Events: unix.EPOLLET | unix.EPOLLIN,
		})

}

// ModRead ...
func (mgr *NetworkMgr) ModRead(epollFd int, fd int) error {
	return unix.EpollCtl(epollFd, unix.EPOLL_CTL_MOD, fd,
		&unix.EpollEvent{Fd: int32(fd),
			Events: unix.EPOLLET | unix.EPOLLIN,
		})
}

// ModReadWrite ...
func (mgr *NetworkMgr) ModReadWrite(epollFd int, fd int) error {
	return unix.EpollCtl(epollFd, unix.EPOLL_CTL_MOD, fd,
		&unix.EpollEvent{Fd: int32(fd),
			Events: unix.EPOLLET | unix.EPOLLIN | unix.EPOLLOUT,
		})
}

// ModDetach ...
func (mgr *NetworkMgr) ModDetach(epollFd int, fd int) error {
	return unix.EpollCtl(epollFd, unix.EPOLL_CTL_DEL, fd,
		&unix.EpollEvent{Fd: int32(fd),
			Events: unix.EPOLLET | unix.EPOLLIN | unix.EPOLLOUT,
		})
}
