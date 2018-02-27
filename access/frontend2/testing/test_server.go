/*
 *  Copyright (c) 2018, https://github.com/nebulaim
 *  All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// 同端口多协议（http）支持测试
// https://core.telegram.org/mtproto#tcp-transport

package main

import (
	"fmt"
	"net/http"
	"net"
	"time"
	"github.com/golang/glog"
	net2 "github.com/nebulaim/telegramd/net"
	"bufio"
	"sync"
	"os"
	"bytes"
)

var _ net.Listener = &HttpListener{}

type HttpListener struct {
	base         net.Listener
	acceptChan   chan net.Conn
	closed       bool
	closeOnce    sync.Once
	closeChan    chan struct{}
	//atomicConnID uint64
	//connsMutex   sync.Mutex
	//conns        map[uint64]*MTProtoConn
}

func Listen(listenFunc func() (net.Listener, error)) (*HttpListener, error) {
	listener, err := listenFunc()
	if err != nil {
		return nil, err
	}
	l := &HttpListener{
		base:       listener,
		closeChan:  make(chan struct{}),
		acceptChan: make(chan net.Conn, 1000),
		// conns:      make(map[uint64]*MTProtoConn),
	}
	// go l.acceptLoop()
	return l, nil
}

func (l *HttpListener) Addr() net.Addr {
	return l.base.Addr()
}

func (l *HttpListener) Close() error {
	l.closeOnce.Do(func() {
		l.closed = true
		close(l.closeChan)
	})
	return l.base.Close()
}

func (l *HttpListener) Accept() (net.Conn, error) {
	select {
	case conn := <-l.acceptChan:
		return conn, nil
	case <-l.closeChan:
	}
	return nil, os.ErrInvalid
}

func onEcho(w http.ResponseWriter, req *http.Request) {
	fmt.Println("req: ", req.RequestURI)
	_, _ = w.Write([]byte(req.RequestURI + "\n"))
}

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:80")

	if err != nil {
		glog.Fatal(err)
	}

	http.HandleFunc("/echo", onEcho)

	srv := &http.Server{Addr: ":8080", Handler: nil}

	ln := &HttpListener{
		// base:       listener,
		closeChan:  make(chan struct{}),
		acceptChan: make(chan net.Conn, 1000),
		// conns:      make(map[uint64]*MTProtoConn),
	}

	go srv.Serve(ln)

	for {
		conn, err := net2.Accept(listener)
		if err != nil {
			panic(err)
		}

		conn2 := newMTProtoConn(conn)
		go func() {
			// r := bufio.NewReaderSize(conn, 1024)
			// hd, _ := r.Peek(4)
			b, e := conn2.Peek(4)
			if e != nil {
				fmt.Println(string(b))
				conn2.Close()
				return
			}
			fmt.Println(string(b))

			if !bytes.Equal(b, []byte("POST")) {
				conn2.Close()
				return
			}

			select {
			case ln.acceptChan <- conn2:
			case <-ln.closeChan:
			}
		}()
	}
}

type MTProtoConn struct {
	base      net.Conn
	r *bufio.Reader

	// listener *Listener
	// id       uint64

	closed    bool
	closeChan chan struct{}
	closeOnce sync.Once
}

func newMTProtoConn(base net.Conn) (conn *MTProtoConn) {
	return &MTProtoConn{
		base:           base,
		r:				bufio.NewReaderSize(base, 1024),
		// id:             id,
		closeChan:      make(chan struct{}),
	}
}

func (c *MTProtoConn) WrapBaseForTest(wrap func(net.Conn) net.Conn) {
	c.base = wrap(c.base)
}

func (c *MTProtoConn) RemoteAddr() net.Addr {
	return c.base.RemoteAddr()
}

func (c *MTProtoConn) LocalAddr() net.Addr {
	return c.base.LocalAddr()
}

func (c *MTProtoConn) SetDeadline(t time.Time) error {
	return c.base.SetDeadline(t)
}

func (c *MTProtoConn) SetReadDeadline(t time.Time) error {
	return c.base.SetReadDeadline(t)
}

func (c *MTProtoConn) SetWriteDeadline(t time.Time) error {
	return c.base.SetWriteDeadline(t)
}

func (c *MTProtoConn) Close() error {
	// glog.Info("Close()")
	c.closeOnce.Do(func() {
		c.closed = true
		// if c.listener != nil {
		// 	c.listener.delConn(c.id)
		// }
		close(c.closeChan)
	})
	return c.base.Close()
}

func (c *MTProtoConn) Peek(n int) ([]byte, error) {
	return c.r.Peek(n)
}

func (c *MTProtoConn) Read(b []byte) (n int, err error) {
	n, err = c.r.Read(b)
	// n, err = c.base.Read(b)
	if err == nil {
		return
	}

	// glog.Warning("MTProtoConn - Will close conn by ", c.base.RemoteAddr(), ", reason: ", err)
	c.base.Close()
	return
}

func (c *MTProtoConn) Write(b []byte) (n int, err error) {
	return c.base.Write(b)
}
