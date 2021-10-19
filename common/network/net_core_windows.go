package network

import (
	"common/log"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type NetConn struct {
	BaseConn
	sendChann     chan []byte
	sendBuff      []byte
	rcvBuff       []byte
	foceClose     int32
	loopWriteFlag int32
	loopReadFlag  int32
	tcpConn       *net.TCPConn
}

func NewNetConn() *NetConn {
	c := &NetConn{}
	c.sessionId = genNextSessionId()
	c.state = int32(ConnStateInit)
	c.attrMap = sync.Map{}
	c.sendChann = make(chan []byte, 65535)
	return c
}

// NetPollCore implements the NetworkCore interface
type NetPollCore struct {
	logger        log.Logger
	eventHandler  NetEventHandler
	listener      net.TCPListener
	connMap       sync.Map // sessionId->connection
	waitConnMap   sync.Map // sessionId->connection
	waitConnTimer time.Ticker
}

func newNetworkCore(opts ...Option) *NetPollCore {
	options := loadOptions(opts)
	netcore := &NetPollCore{}
	netcore.logger = options.logger
	netcore.eventHandler = options.eventHandler
	netcore.connMap = sync.Map{}
	netcore.waitConnMap = sync.Map{}
	netcore.waitConnTimer = *time.NewTicker(time.Duration(100) * time.Millisecond)
	netcore.startWaitConnTimer()
	return netcore
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

		if atomic.LoadInt32(&conn.loopReadFlag) != 0 ||
			atomic.LoadInt32(&conn.loopWriteFlag) != 0 {
			return true
		}

		endpoint := fmt.Sprintf("%s:%d", conn.peerHost, conn.peerPort)
		dialer := net.Dialer{Timeout: time.Duration(200) * time.Millisecond}
		tcpConn, err := dialer.Dial("tcp", endpoint)
		if err != nil {
			netcore.logger.LogError("dial tcp error: %v, endpoint: %v", err, endpoint)
			conn.lastTryConTime = t.Unix()
			return true
		}
		conn.tcpConn = tcpConn.(*net.TCPConn)
		conn.SetConnState(ConnStateConnected)
		conn.lastTryConTime = 0

		netcore.waitConnMap.Delete(conn.sessionId)
		netcore.connMap.Store(conn.sessionId, conn)
		netcore.eventHandler.OnConnected(conn)

		atomic.StoreInt32(&conn.loopReadFlag, 1)
		atomic.StoreInt32(&conn.loopWriteFlag, 1)
		atomic.StoreInt32(&conn.foceClose, 0)
		go netcore.loopRead(conn)
		go netcore.loopWrite(conn)
		return true
	})
}

func (netcore *NetPollCore) loopAccept() {
	for {
		tcpConn, err := netcore.listener.AcceptTCP()
		if err != nil {
			netcore.logger.LogError("accept error :%v", err)
			continue
		}

		conn := NewNetConn()
		conn.state = int32(ConnStateConnected)
		conn.tcpConn = tcpConn
		remoteAddr := tcpConn.RemoteAddr().String()
		addrSplits := strings.Split(remoteAddr, ":")
		if len(addrSplits) >= 2 {
			conn.peerHost = addrSplits[0]
			conn.peerPort, _ = strconv.Atoi(addrSplits[1])
		}

		netcore.connMap.Store(conn.sessionId, conn)
		netcore.eventHandler.OnAccept(conn)

		go netcore.loopRead(conn)
		go netcore.loopWrite(conn)
	}
}

func (netcore *NetPollCore) loopWrite(conn *NetConn) error {
	for {
		for {
			timeout := false
			t := time.NewTimer(time.Duration(100) * time.Millisecond)
			select {
			case buff := <-conn.sendChann:
				conn.sendBuff = append(conn.sendBuff, buff...)
			case <-t.C:
				timeout = true
			}
			if timeout {
				break
			}
		}

		bClose := false
		if len(conn.sendBuff) > 0 {
			conn.tcpConn.SetWriteDeadline(time.Now().Add(time.Duration(100) + time.Second))
			n, err := conn.tcpConn.Write(conn.sendBuff)
			if err != nil || n <= 0 {
				if !errors.Is(err, os.ErrDeadlineExceeded) {
					netcore.logger.LogError("tcp conn write error:%v, writeLen:%v, sessionId:%v",
						err, n, conn.sessionId)
					bClose = true
				}
			} else {
				conn.sendBuff = conn.sendBuff[n:]
			}
		}

		if !bClose {
			if atomic.LoadInt32(&conn.foceClose) > 0 {
				bClose = true
			}
		}

		if bClose {
			atomic.StoreInt32(&conn.loopWriteFlag, 0)
			netcore.logger.LogDebug("clear loop writing flag, is_client:%v", conn.isClient)
			netcore.close(conn)
			break
		}
	}

	return nil
}

func (netcore *NetPollCore) loopRead(conn *NetConn) error {
	tcpConn := conn.tcpConn
	buff := make([]byte, 65535)
	for {
		bClose := false
		tcpConn.SetReadDeadline(time.Now().Add(time.Duration(100) + time.Second))
		n, err := tcpConn.Read(buff)
		if err != nil || n <= 0 {
			if !errors.Is(err, os.ErrDeadlineExceeded) {
				netcore.logger.LogError("tcp conn read error:%v, readLen:%v", err, n)
				bClose = true
			}
		} else {
			netcore.logger.LogDebug("session:%v rcv data : %v", conn.sessionId, string(buff[:n]))
			conn.rcvBuff = append(conn.rcvBuff, buff...)
		}

		if !bClose {
			if atomic.LoadInt32(&conn.foceClose) > 0 {
				bClose = true
			}
		}

		if bClose {
			atomic.StoreInt32(&conn.loopReadFlag, 0)
			netcore.logger.LogDebug("clear loop reading flag, is_client:%v", conn.isClient)
			netcore.close(conn)
			break
		}

	}
	return nil
}

func (netcore *NetPollCore) close(conn *NetConn) error {
	if conn.GetConnState() != ConnStateClosed {
		conn.tcpConn.Close()
		conn.SetConnState(ConnStateClosed)
		netcore.connMap.Delete(conn.sessionId)
		netcore.eventHandler.OnClosed(conn)

		if conn.isClient {
			netcore.waitConnMap.Store(conn.sessionId, conn)
		}
	}

	return nil
}

func (netcore *NetPollCore) TcpListen(host string, port int) error {
	endpoint := fmt.Sprintf("%s:%d", host, port)
	tcpAddr, err := net.ResolveTCPAddr("tcp4", endpoint)
	if err != nil {
		netcore.logger.LogError("netcore resolve tcp addr error:%v, endpoint:%v", err, endpoint)
		return nil
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		if listener != nil {
			listener.Close()
		}
		netcore.logger.LogError("netcore listen error:%v, endpoint:%v", err, endpoint)
		return nil
	}
	netcore.listener = *listener

	go netcore.loopAccept()

	return nil
}

func (netcore *NetPollCore) TcpConnect(host string, port int) error {
	conn := NewNetConn()
	conn.isClient = true
	conn.peerHost = host
	conn.peerPort = port
	netcore.waitConnMap.Store(conn.sessionId, conn)
	return nil
}

func (netcore *NetPollCore) TcpSend(sessionId int64, buff []byte) error {
	if len(buff) <= 0 {
		return nil
	}

	c, ok := netcore.connMap.Load(sessionId)
	if !ok {
		netcore.logger.LogError("tcp send connection not found, sessionId:%v", sessionId)
		return errors.New("tcp send connection not found")
	}
	conn := c.(*NetConn)
	select {
	case conn.sendChann <- buff:
	default:
		netcore.logger.LogError("tcp send channel full")
		return errors.New("send channel full")
	}

	return nil
}

func (netcore *NetPollCore) TcpClose(sessionId int64) error {
	c, ok := netcore.connMap.Load(sessionId)
	if !ok {
		netcore.logger.LogError("tcp close connection not found, sessionId:%v", sessionId)
		return errors.New("tcp close connection not found")
	}
	conn := c.(*NetConn)
	swapped := atomic.CompareAndSwapInt32(&conn.foceClose, 0, 1)
	if !swapped {
		netcore.logger.LogInfo("tcp close connection, the connection is closing yet")
	}
	return nil
}
