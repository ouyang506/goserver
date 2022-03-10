package network

import (
	"common/log"
	"common/utility/ringbuffer"
	"context"
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

	foceClose int32 // atomic
	tcpConn   *net.TCPConn
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
	ctx         context.Context
	ctxCancelFn context.CancelFunc

	listener net.TCPListener

	eventHandler         NetEventHandler
	socketSendBufferSize int
	socketRcvBufferSize  int
	socketTcpNoDelay     bool
	codec                Codec

	waitConnTimer time.Ticker

	connMapMutex sync.RWMutex
	waitConnMap  map[int64]*NetConn // 待发起的连接(client)
	connectedMap map[int64]*NetConn // 已建立的连接
}

func newNetworkCore(opts ...Option) *NetPollCore {
	options := loadOptions(opts)
	netcore := &NetPollCore{}
	netcore.ctx, netcore.ctxCancelFn = context.WithCancel(context.Background())
	netcore.eventHandler = options.eventHandler
	netcore.socketSendBufferSize = options.socketSendBufferSize
	netcore.socketRcvBufferSize = options.socketRcvBufferSize
	netcore.socketTcpNoDelay = options.socketTcpNoDelay
	netcore.codec = options.codec
	netcore.connectedMap = map[int64]*NetConn{}
	netcore.waitConnMap = map[int64]*NetConn{}
	netcore.waitConnTimer = *time.NewTicker(time.Duration(100) * time.Millisecond)

	return netcore
}

func (netcore *NetPollCore) Start() {
	netcore.startWaitConnTimer()
}

func (netcore *NetPollCore) Stop() {
	netcore.ctxCancelFn()
}

func (netcore *NetPollCore) startWaitConnTimer() {
	go func() {
		defer netcore.waitConnTimer.Stop()

		ctx, cancel := context.WithCancel(netcore.ctx)
		defer cancel()

		for {
			bStop := false
			select {
			case t := <-netcore.waitConnTimer.C:
				{
					netcore.onWaitConnTimer(t)
				}
			case <-ctx.Done():
				{
					bStop = true
				}
			}
			if bStop {
				return
			}
		}
	}()
}

func (netcore *NetPollCore) onWaitConnTimer(t time.Time) {
	waitConnList := netcore.getWaitConnList()
	if len(waitConnList) <= 0 {
		return
	}

	for _, conn := range waitConnList {
		// 已经强制关闭的不再请求连接
		if atomic.LoadInt32(&conn.foceClose) > 0 {
			continue
		}

		if t.Unix()-conn.lastTryConTime < int64(RECONNECT_DELTA_TIME_SEC) {
			continue
		}

		// 上一连接的事件loop未结束
		if conn.GetConnState() != ConnStateInit {
			continue
		}

		endpoint := fmt.Sprintf("%s:%d", conn.peerHost, conn.peerPort)
		dialer := net.Dialer{Timeout: time.Duration(200) * time.Millisecond}
		tcpConn, err := dialer.Dial("tcp", endpoint)
		if err != nil {
			log.Error("dial tcp error: %v, endpoint: %v", err, endpoint)
			netcore.eventHandler.OnConnect(conn, err)

			if conn.autoReconnect {
				conn.lastTryConTime = t.Unix()
			} else {
				netcore.removeWaitingConn(conn.sessionId)
			}
			return
		}

		if !netcore.moveConnToConnected(conn.sessionId) {
			log.Info("net connected but connection is force closed, sessionId: %v, peerHost: %v, peerPort: %v",
				conn.sessionId, conn.peerHost, conn.peerPort)
			return
		}

		conn.tcpConn = tcpConn.(*net.TCPConn)
		conn.tcpConn.SetWriteBuffer(netcore.socketSendBufferSize)
		conn.tcpConn.SetReadBuffer(netcore.socketRcvBufferSize)
		conn.tcpConn.SetNoDelay(netcore.socketTcpNoDelay)

		conn.SetConnState(ConnStateConnected)
		conn.lastTryConTime = 0

		netcore.eventHandler.OnConnect(conn, nil)

		netcore.loopEvent(conn)
	}
}

func (netcore *NetPollCore) loopAccept() {
	for {
		tcpConn, err := netcore.listener.AcceptTCP()
		if err != nil {
			log.Error("accept error :%v", err)
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

		netcore.addConnectedConn(conn.sessionId, conn)
		netcore.eventHandler.OnAccept(conn)

		netcore.loopEvent(conn)
	}
}

func (netcore *NetPollCore) loopEvent(conn *NetConn) {
	go func() {
		wg := sync.WaitGroup{}
		wg.Add(2)
		go func() {
			netcore.loopRead(conn)
			wg.Done()
		}()
		go func() {
			netcore.loopWrite(conn)
			wg.Done()
		}()
		//loop结束重置为初始化状态
		conn.SetConnState(ConnStateInit)
	}()
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

			tcpConn.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
			n, err := tcpConn.Read(buff)
			if err != nil || n <= 0 {
				if !errors.Is(err, os.ErrDeadlineExceeded) {
					log.Error("tcp conn read error:%v, readLen:%v", err, n)
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
				log.Debug("loop read decode msg error, %s", err)
				bClose = true
				break
			}
			if len(msgBuf) > 0 {
				log.Debug("session:%v rcv data : %v", conn.sessionId, string(msgBuf))
			} else {
				if conn.rcvBuff.IsFull() {
					oldCap := conn.rcvBuff.Cap()
					conn.rcvBuff.Grow(oldCap + oldCap/2)
					log.Info("loop read grow rcv buffer capcity from %v to %v, sessionId:%v", oldCap, conn.rcvBuff.Cap(), conn.sessionId)
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
			netcore.close(conn)
			break
		}

	}
	return nil
}

func (netcore *NetPollCore) loopWrite(conn *NetConn) error {
	for {
		bClose := false
		for {
			timeout := false
			t := time.NewTimer(time.Millisecond * 100)
			select {
			case buff := <-conn.sendChann:
				err := netcore.codec.Encode(buff, conn.sendBuff)
				if err != nil {
					bClose = true
					log.Error("loop write encode buff error:%s, sessionId:%d", err, conn.sessionId)
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
				conn.tcpConn.SetWriteDeadline(time.Now().Add(time.Millisecond * 100))
				n, err := conn.tcpConn.Write(b)
				if err != nil || n <= 0 {
					if !errors.Is(err, os.ErrDeadlineExceeded) {
						log.Error("tcp conn write error:%v, writeLen:%v, sessionId:%v",
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
			netcore.close(conn)
			break
		}
	}

	return nil
}

func (netcore *NetPollCore) close(conn *NetConn) error {
	if conn.CompareAndSwapConnState(ConnStateConnected, ConnStateClosed) {
		conn.tcpConn.Close()
		netcore.eventHandler.OnClosed(conn)

		if !conn.isClient {
			netcore.removeConnectedConn(conn.sessionId)
		} else {
			if conn.autoReconnect {
				netcore.moveConnToWaiting(conn.sessionId)
			}
		}
	}

	return nil
}

// 获取待连接列表的拷贝
func (netcore *NetPollCore) getWaitConnList() []*NetConn {
	netcore.connMapMutex.RLock()
	defer netcore.connMapMutex.RUnlock()
	values := make([]*NetConn, 0, len(netcore.waitConnMap))
	for _, v := range netcore.waitConnMap {
		values = append(values, v)
	}
	return values
}

// 从待连接列表移动到已连接列表
func (netcore *NetPollCore) moveConnToConnected(sessionId int64) bool {
	netcore.connMapMutex.Lock()
	defer netcore.connMapMutex.Unlock()
	conn, ok := netcore.waitConnMap[sessionId]
	if !ok {
		return false
	}
	netcore.connectedMap[sessionId] = conn
	return true
}

// 从已连接列表移动到待连接列表
func (netcore *NetPollCore) moveConnToWaiting(sessionId int64) bool {
	netcore.connMapMutex.Lock()
	defer netcore.connMapMutex.Unlock()
	conn, ok := netcore.connectedMap[sessionId]
	if !ok {
		return false
	}
	netcore.waitConnMap[sessionId] = conn
	return true
}

// 添加到待连接列表
func (netcore *NetPollCore) addWaitingConn(sessionId int64, conn *NetConn) bool {
	netcore.connMapMutex.Lock()
	defer netcore.connMapMutex.Unlock()
	_, ok := netcore.waitConnMap[sessionId]
	if !ok {
		return false
	}
	netcore.waitConnMap[sessionId] = conn
	return true
}

// 从待连接列表移除
func (netcore *NetPollCore) removeWaitingConn(sessionId int64) bool {
	netcore.connMapMutex.Lock()
	defer netcore.connMapMutex.Unlock()
	_, ok := netcore.waitConnMap[sessionId]
	if !ok {
		return false
	}
	delete(netcore.waitConnMap, sessionId)
	return true
}

// 从已连接列表查找
func (netcore *NetPollCore) getConnectedConn(sessionId int64) *NetConn {
	netcore.connMapMutex.RLock()
	defer netcore.connMapMutex.RUnlock()
	conn, ok := netcore.connectedMap[sessionId]
	if !ok {
		return nil
	}
	return conn
}

// 添加到已连接列表
func (netcore *NetPollCore) addConnectedConn(sessionId int64, conn *NetConn) bool {
	netcore.connMapMutex.Lock()
	defer netcore.connMapMutex.Unlock()
	_, ok := netcore.connectedMap[sessionId]
	if !ok {
		return false
	}
	netcore.connectedMap[sessionId] = conn
	return true
}

// 从已连接列表移除
func (netcore *NetPollCore) removeConnectedConn(sessionId int64) bool {
	netcore.connMapMutex.Lock()
	defer netcore.connMapMutex.Unlock()
	_, ok := netcore.connectedMap[sessionId]
	if !ok {
		return false
	}
	delete(netcore.connectedMap, sessionId)
	return true
}

// 从待连接列表和已连接列表移除
func (netcore *NetPollCore) forceRemoveConn(sessionId int64) bool {
	netcore.connMapMutex.Lock()
	defer netcore.connMapMutex.Unlock()

	conn, ok := netcore.connectedMap[sessionId]
	if ok {
		atomic.StoreInt32(&conn.foceClose, 1)
	}
	delete(netcore.connectedMap, sessionId)

	conn, ok = netcore.waitConnMap[sessionId]
	if ok {
		atomic.StoreInt32(&conn.foceClose, 1)
	}
	delete(netcore.waitConnMap, sessionId)
	return true
}

func (netcore *NetPollCore) TcpListen(host string, port int) error {
	endpoint := fmt.Sprintf("%s:%d", host, port)
	tcpAddr, err := net.ResolveTCPAddr("tcp4", endpoint)
	if err != nil {
		log.Error("netcore resolve tcp addr error:%v, endpoint:%v", err, endpoint)
		return nil
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		if listener != nil {
			listener.Close()
		}
		log.Error("netcore listen error:%v, endpoint:%v", err, endpoint)
		return nil
	}
	netcore.listener = *listener

	go netcore.loopAccept()

	return nil
}

func (netcore *NetPollCore) TcpConnect(host string, port int, autoReconnect bool, attrib map[interface{}]interface{}) (Connection, error) {
	conn := NewNetConn(netcore.socketSendBufferSize, netcore.socketRcvBufferSize)
	conn.isClient = true
	conn.autoReconnect = autoReconnect
	conn.peerHost = host
	conn.peerPort = port

	for k, v := range attrib {
		conn.SetAttrib(k, v)
	}

	netcore.addWaitingConn(conn.sessionId, conn)
	return conn, nil
}

func (netcore *NetPollCore) TcpSend(sessionId int64, buff []byte) error {
	if len(buff) <= 0 {
		return nil
	}

	conn := netcore.getConnectedConn(sessionId)
	if conn == nil {
		log.Error("tcp send connection not found, sessionId:%v", sessionId)
		return errors.New("tcp send connection not found")
	}

	select {
	case conn.sendChann <- buff:
	default:
		log.Error("tcp send channel full")
		return errors.New("send channel full")
	}

	return nil
}

func (netcore *NetPollCore) TcpClose(sessionId int64) error {
	netcore.forceRemoveConn(sessionId)
	return nil
}
