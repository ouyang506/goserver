package network

import (
	"common/log"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

const (
	E_CONN_ATTRIB_TCP_CONN             = 1
	E_CONN_ATTRIB_TCP_SEND_CHAN        = 2
	E_CONN_ATTRIB_TCP_CLOSE_WRITE_CHAN = 3
	E_CONN_ATTRIB_TCP_CLOSE_READ_CHAN  = 4
	E_CONN_ATTRIB_TCP_LOOP_WRITING     = 5
	E_CONN_ATTRIB_TCP_LOOP_READING     = 6

	RECONNECT_DELTA_TIME_SEC = 2
)

// NetPollCore implements the NetworkCore interface
type NetPollCore struct {
	logger        log.Logger
	eventHandler  NetEventHandler
	listener      net.TCPListener
	connMap       sync.Map // sessionId->connection
	waitConnMap   sync.Map // sessionId->connection
	waitConnTimer time.Ticker
}

func newNetworkCore(numLoops int, loadBalance LoadBalance, eventHandler NetEventHandler, logger log.Logger) *NetPollCore {
	netcore := &NetPollCore{}
	netcore.logger = logger
	netcore.eventHandler = eventHandler
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
		conn := value.(*Connection)
		if t.Unix()-conn.GetLastTryConnectTime() < int64(RECONNECT_DELTA_TIME_SEC) {
			return true
		}

		if conn.GetAttrib(E_CONN_ATTRIB_TCP_LOOP_WRITING) != nil ||
			conn.GetAttrib(E_CONN_ATTRIB_TCP_LOOP_READING) != nil {
			conn.SetLastTryConnectTime(conn.GetLastTryConnectTime() + 1)
			return true
		}

		endpoint := fmt.Sprintf("%s:%d", conn.peerHost, conn.peerPort)
		dialer := net.Dialer{Timeout: time.Duration(200) * time.Millisecond}
		netConn, err := dialer.Dial("tcp", endpoint)
		if err != nil {
			netcore.logger.LogError("dial tcp error: %v, endpoint: %v", err, endpoint)
			conn.SetLastTryConnectTime(t.Unix())
			return true
		}
		conn.SetConnected()
		conn.SetLastTryConnectTime(0)
		conn.SetAttrib(E_CONN_ATTRIB_TCP_CONN, netConn)
		conn.SetAttrib(E_CONN_ATTRIB_TCP_SEND_CHAN, make(chan []byte, 65535))
		conn.SetAttrib(E_CONN_ATTRIB_TCP_CLOSE_WRITE_CHAN, make(chan int, 1))
		conn.SetAttrib(E_CONN_ATTRIB_TCP_CLOSE_READ_CHAN, make(chan int, 1))
		netcore.waitConnMap.Delete(conn.sessionId)
		netcore.connMap.Store(conn.sessionId, conn)
		netcore.eventHandler.OnConnected(conn)

		conn.SetAttrib(E_CONN_ATTRIB_TCP_LOOP_WRITING, true)
		conn.SetAttrib(E_CONN_ATTRIB_TCP_LOOP_READING, true)
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

		conn := NewConnection()
		conn.SetConnected()
		conn.SetAttrib(E_CONN_ATTRIB_TCP_CONN, tcpConn)
		conn.SetAttrib(E_CONN_ATTRIB_TCP_SEND_CHAN, make(chan []byte, 65535))
		conn.SetAttrib(E_CONN_ATTRIB_TCP_CLOSE_WRITE_CHAN, make(chan int, 1))
		conn.SetAttrib(E_CONN_ATTRIB_TCP_CLOSE_READ_CHAN, make(chan int, 1))
		netcore.connMap.Store(conn.sessionId, conn)
		netcore.eventHandler.OnAccept(conn)

		go netcore.loopRead(conn)
		go netcore.loopWrite(conn)
	}
}

func (netcore *NetPollCore) loopWrite(conn *Connection) error {

	for {
		sendChann := conn.GetAttrib(E_CONN_ATTRIB_TCP_SEND_CHAN).(chan []byte)
		for {
			timeout := false
			t := time.NewTimer(time.Duration(100) * time.Millisecond)
			select {
			case buff := <-sendChann:
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
			tcpConn := conn.GetAttrib(E_CONN_ATTRIB_TCP_CONN).(*net.TCPConn)
			tcpConn.SetWriteDeadline(time.Now().Add(time.Duration(100) + time.Second))
			n, err := tcpConn.Write(conn.sendBuff)
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
			closeChan := conn.GetAttrib(E_CONN_ATTRIB_TCP_CLOSE_WRITE_CHAN)
			if closeChan != nil {
				select {
				case <-closeChan.(chan int):
					bClose = true
				default:
				}
			}
		}

		if bClose {
			conn.SetAttrib(E_CONN_ATTRIB_TCP_LOOP_WRITING, nil)
			netcore.logger.LogDebug("clear loop writing flag, is_client:%v", conn.isClient)
			netcore.close(conn)
			break
		}
	}

	return nil
}

func (netcore *NetPollCore) loopRead(conn *Connection) error {
	c := conn.GetAttrib(E_CONN_ATTRIB_TCP_CONN)
	if c == nil {
		return errors.New("get attrib tcp conn nil")
	}
	tcpConn := c.(*net.TCPConn)
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
			closeChan := conn.GetAttrib(E_CONN_ATTRIB_TCP_CLOSE_READ_CHAN)
			if closeChan != nil {
				select {
				case <-closeChan.(chan int):
					bClose = true
				default:
				}
			}
		}

		if bClose {
			conn.SetAttrib(E_CONN_ATTRIB_TCP_LOOP_READING, nil)
			netcore.logger.LogDebug("clear loop reading flag, is_client:%v", conn.isClient)
			netcore.close(conn)
			break
		}

	}
	return nil
}

func (netcore *NetPollCore) close(conn *Connection) error {
	if !conn.IsClosed() {
		tcpConn := conn.GetAttrib(E_CONN_ATTRIB_TCP_CONN)
		tcpConn.(*net.TCPConn).Close()

		conn.SetClosed()
		netcore.connMap.Delete(conn.sessionId)
		netcore.eventHandler.OnClosed(conn)
	}

	if conn.isClient {
		netcore.waitConnMap.Store(conn.sessionId, conn)
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
	conn := NewConnection()
	conn.SetClient(true)
	conn.SetLastTryConnectTime(0)
	conn.SetPeerAddr(host, port)
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
	conn := c.(*Connection)
	sendChan := conn.GetAttrib(E_CONN_ATTRIB_TCP_SEND_CHAN)
	if sendChan == nil {
		netcore.logger.LogError("tcp send get connection send chan nil")
		return errors.New("connection send chan nil")
	}
	select {
	case sendChan.(chan []byte) <- buff:
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
	conn := c.(*Connection)
	closeWriteChan := conn.GetAttrib(E_CONN_ATTRIB_TCP_CLOSE_WRITE_CHAN)
	select {
	case closeWriteChan.(chan int) <- 1:
	default:
	}

	closeReadChan := conn.GetAttrib(E_CONN_ATTRIB_TCP_CLOSE_READ_CHAN)
	select {
	case closeReadChan.(chan int) <- 1:
	default:
	}

	return nil
}
