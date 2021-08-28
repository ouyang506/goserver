// Copyright
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// +build linux

package gnet

import (
	"net"
	"os"
	"sync"
	"time"

	"goserver/common/network/gnet/errors"
	"goserver/common/network/gnet/internal/logging"
	"goserver/common/network/gnet/internal/socket"

	"golang.org/x/sys/unix"
)

type connector struct {
	once          sync.Once
	fd            int
	lnaddr        net.Addr
	addr, network string
	sockopts      []socket.Option
}

func (c *connector) normalize() (err error) {
	switch c.network {
	case "tcp", "tcp4", "tcp6":
		c.fd, c.lnaddr, err = socket.TCPSocketConnect(c.network, c.addr, c.sockopts...)
		c.network = "tcp"
	// case "udp", "udp4", "udp6":
	// 	c.fd, c.lnaddr, err = socket.UDPSocket(c.network, c.addr, c.sockopts...)
	// 	c.network = "udp"
	// case "unix":
	// 	_ = os.RemoveAll(c.addr)
	// 	c.fd, c.lnaddr, err = socket.UnixSocket(c.network, c.addr, c.sockopts...)
	default:
		err = errors.ErrUnsupportedProtocol
	}
	return
}

func (c *connector) close() {
	c.once.Do(
		func() {
			if c.fd > 0 {
				logging.LogErr(os.NewSyscallError("close", unix.Close(c.fd)))
			}
			if c.network == "unix" {
				logging.LogErr(os.RemoveAll(c.addr))
			}
		})
}

func initConnector(network, addr string, options *Options) (c *connector, err error) {
	var sockopts []socket.Option

	if network == "tcp" && options.TCPNoDelay == TCPNoDelay {
		sockopt := socket.Option{SetSockopt: socket.SetNoDelay, Opt: 1}
		sockopts = append(sockopts, sockopt)
	}
	if network == "tcp" && options.TCPKeepAlive > 0 {
		sockopt := socket.Option{SetSockopt: socket.SetKeepAlive, Opt: int(options.TCPKeepAlive / time.Second)}
		sockopts = append(sockopts, sockopt)
	}
	if options.SocketRecvBuffer > 0 {
		sockopt := socket.Option{SetSockopt: socket.SetRecvBuffer, Opt: options.SocketRecvBuffer}
		sockopts = append(sockopts, sockopt)
	}
	if options.SocketSendBuffer > 0 {
		sockopt := socket.Option{SetSockopt: socket.SetSendBuffer, Opt: options.SocketSendBuffer}
		sockopts = append(sockopts, sockopt)
	}
	c = &connector{network: network, addr: addr, sockopts: sockopts}
	err = c.normalize()
	return
}
