package network

import (
	"common/log"
	"common/utility/ringbuffer"
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
	sendChann chan []byte
	sendBuff  *ringbuffer.RingBuffer
	rcvBuff   *ringbuffer.RingBuffer

	foceClose     int32
	loopWriteFlag int32
	loopReadFlag  int32
	tcpConn       *net.TCPConn
}

func NewNetConn(sendBuffSize int, rcvBuffSize int) *NetConn {
	c := &NetConn{}
	c.sessionId = genNextSessionId()
	c.state = int32(ConnStateInit)
	c.attrMap = sync.Map{}
	c.sendChann = make(chan []byte, sendBuffSize/4)
	c.sendBuff = ringbuffer.NewRingBuffer(sendBuffSize)
	c.rcvBuff = ringbuffer.NewRingBuffer(rcvBuffSize)
	return c
}

// NetPollCore implements the NetworkCore interface
type NetPollCore struct {
	logger               log.Logger
	eventHandler         NetEventHandler
	socketSendBufferSize int
	socketRcvBufferSize  int
	socketTcpNoDelay     bool
	codec                Codec

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
	netcore.socketSendBufferSize = options.socketSendBufferSize
	netcore.socketRcvBufferSize = options.socketRcvBufferSize
	netcore.socketTcpNoDelay = options.socketTcpNoDelay
	netcore.codec = options.codec
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
		conn.tcpConn.SetWriteBuffer(netcore.socketSendBufferSize)
		conn.tcpConn.SetReadBuffer(netcore.socketRcvBufferSize)
		conn.tcpConn.SetNoDelay(netcore.socketTcpNoDelay)

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

		conn := NewNetConn(netcore.socketSendBufferSize, netcore.socketRcvBufferSize)
		conn.state = int32(ConnStateConnected)
		conn.tcpConn = tcpConn
		conn.tcpConn.SetWriteBuffer(netcore.socketSendBufferSize)
		conn.tcpConn.SetReadBuffer(netcore.socketRcvBufferSize)
		conn.tcpConn.SetNoDelay(netcore.socketTcpNoDelay)
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
		bClose := false
		for {
			timeout := false
			t := time.NewTimer(time.Duration(100) * time.Millisecond)
			select {
			case buff := <-conn.sendChann:
				err := netcore.codec.Encode(buff, conn.sendBuff)
				if err != nil {
					bClose = true
					netcore.logger.LogError("loop write encode buff error:%s, sessionId:%d", err, conn.sessionId)
				}
			case <-t.C:
				timeout = true
			}
			if timeout || bClose {
				break
			}
		}

		totalWrite := 0
		if !conn.sendBuff.IsEmpty() {
			head, tail := conn.sendBuff.PeekAll()
			for _, b := range [2][]byte{head, tail} {
				if len(b) <= 0 {
					break
				}
				conn.tcpConn.SetWriteDeadline(time.Now().Add(time.Duration(100) + time.Second))
				n, err := conn.tcpConn.Write(b)
				if err != nil || n <= 0 {
					if !errors.Is(err, os.ErrDeadlineExceeded) {
						netcore.logger.LogError("tcp conn write error:%v, writeLen:%v, sessionId:%v",
							err, n, conn.sessionId)
						bClose = true
						break
					}
				} else {
					totalWrite += n
				}
			}
		}
		if totalWrite > 0 {
			conn.sendBuff.Discard(totalWrite)
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

	for {
		head, tail := conn.rcvBuff.PeekFreeAll()
		bClose := false
		totalRead := 0
		for _, buff := range [2][]byte{head, tail} {
			if len(buff) <= 0 {
				break
			}

			tcpConn.SetReadDeadline(time.Now().Add(time.Duration(100) + time.Second))
			n, err := tcpConn.Read(buff)
			if err != nil || n <= 0 {
				if !errors.Is(err, os.ErrDeadlineExceeded) {
					netcore.logger.LogError("tcp conn read error:%v, readLen:%v", err, n)
					bClose = true
					break
				}
			} else {
				totalRead += n
			}
		}

		if totalRead > 0 {
			conn.rcvBuff.Foward(totalRead)
		}

		for {
			msgBuf, err := netcore.codec.Decode(conn.rcvBuff)
			if err != nil {
				netcore.logger.LogDebug("loop read decode msg error, %s", err)
				bClose = true
				break
			}
			if len(msgBuf) > 0 {
				netcore.logger.LogDebug("session:%v rcv data : %v", conn.sessionId, string(msgBuf))
			} else {
				if conn.rcvBuff.IsFull() {
					oldCap := conn.rcvBuff.Cap()
					conn.rcvBuff.Grow(oldCap + oldCap/2)
					netcore.logger.LogInfo("loop read grow rcv buffer capcity from %v to %v, sessionId:%v", oldCap, conn.rcvBuff.Cap(), conn.sessionId)
				}
				break
			}
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

func (netcore *NetPollCore) TcpConnect(host string, port int) (int64, error) {
	conn := NewNetConn(netcore.socketSendBufferSize, netcore.socketRcvBufferSize)
	conn.isClient = true
	conn.peerHost = host
	conn.peerPort = port
	netcore.waitConnMap.Store(conn.sessionId, conn)
	return conn.GetSessionId(), nil
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
