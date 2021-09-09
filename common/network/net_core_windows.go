package network

import (
	"common/log"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

var (
	E_CONN_ATTRIB_TCP_CONN = 1
)

// NetPollCore implements the NetworkCore interface
type NetPollCore struct {
	logger       log.Logger
	numLoops     int
	loadBalance  LoadBalance
	eventHandler NetEventHandler
	listener     net.TCPListener
	connMap      map[int64]*Connection // sessionId->connection
	connMapMutex sync.Mutex
}

func newNetworkCore(numLoops int, loadBalance LoadBalance, eventHandler NetEventHandler, logger log.Logger) *NetPollCore {
	netcore := &NetPollCore{}
	netcore.logger = logger
	netcore.numLoops = numLoops
	netcore.loadBalance = loadBalance
	netcore.eventHandler = eventHandler
	netcore.startLoop()
	return netcore
}

func (netcore *NetPollCore) startLoop() error {

	for i := 0; i < netcore.numLoops; i++ {

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

	go func() {
		for {
			tcpConn, err := netcore.listener.AcceptTCP()
			if err != nil {
				netcore.logger.LogError("accept error :%v", err)
				continue
			}

			conn := NewConnection()
			conn.SetAttrib(E_CONN_ATTRIB_TCP_CONN, tcpConn)
			netcore.connMap[conn.sessionId] = conn
			netcore.eventHandler.OnAccept(conn)
		}

	}()
	return nil
}

func (netcore *NetPollCore) TcpConnect(host string, port int) error {

	endpoint := fmt.Sprintf("%s:%d", host, port)
	// tcpAddr, err := net.ResolveTCPAddr("tcp4", endpoint)
	// if err != nil {
	// 	netcore.logger.LogError("netcore resolve tcp addr error:%v, endpoint:%v", err, endpoint)
	// 	return nil
	// }

	dialer := net.Dialer{Timeout: time.Duration(1) * time.Second}
	netConn, err := dialer.Dial("tcp", endpoint)
	if err != nil {
		return err
	}

	tcpConn, ok := netConn.(*net.TCPConn)
	if !ok {
		return errors.New("convert to tcp conn error")
	}

	conn := NewConnection()
	conn.SetAttrib(E_CONN_ATTRIB_TCP_CONN, tcpConn)
	netcore.connMap[conn.sessionId] = conn

	netcore.eventHandler.OnConnected(conn)

	return nil
}

func (netcore *NetPollCore) TcpSend(sessionId int64, buff []byte) error {

	conn, ok := netcore.connMap[sessionId]
	if !ok {
		netcore.logger.LogError("tcp send connection not found, sessionId:%v", sessionId)
		return errors.New("tcp send connection not found")
	}

	conn.sendBuff = append(conn.sendBuff, buff...)
	tcpConn := conn.GetAttrib(E_CONN_ATTRIB_TCP_CONN)
	if tcpConn == nil {
		return errors.New("tcp send connection not found")
	}

	n, err := tcpConn.(*net.TCPConn).Write(conn.sendBuff)
	if err != nil {
		netcore.logger.LogError("tcp conn write error:%v, sessionId:%v", err, conn.sessionId)
		netcore.TcpClose(sessionId)
		return err
	}
	if n != len(conn.sendBuff) {
		netcore.logger.LogError("tcp conn write ret len error, writeLen:%v, buffLen:%v, sessionId:%v",
			n, len(conn.sendBuff), conn.sessionId)
		netcore.TcpClose(sessionId)
		return errors.New("tcp conn write return length error")
	}
	conn.sendBuff = conn.sendBuff[:0]

	return nil
}

func (netcore *NetPollCore) TcpClose(sessionId int64) error {
	conn, ok := netcore.connMap[sessionId]
	if !ok {
		netcore.logger.LogError("tcp close connection not found, sessionId:%v", sessionId)
		return errors.New("tcp close connection not found")
	}

	tcpConn := conn.GetAttrib(E_CONN_ATTRIB_TCP_CONN)
	if tcpConn == nil {
		return errors.New("tcp close connection not found")
	}

	err := tcpConn.(*net.TCPConn).Close()
	if err != nil {
		return err
	}

	return nil
}
