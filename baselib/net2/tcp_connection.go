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

package net2

import (
	"net"
	"sync"
	"sync/atomic"
	"errors"
)

type TcpConnectionCallback interface {
	OnNewConnection(conn *TcpConnection)
	OnDataArrived(c *TcpConnection, msg interface{}) error
	OnConnectionClosed(c *TcpConnection)
}

var ConnectionClosedError = errors.New("Connection Closed")
var ConnectionBlockedError = errors.New("Connection Blocked")

var globalConnectionId uint64

type TcpConnection struct {
	name       string
	conn 			  *net.TCPConn
	id                uint64
	codec             Codec
	sendChan          chan interface{}
	recvMutex         sync.Mutex
	sendMutex         sync.RWMutex

	closeFlag  int32
	closeChan  chan int
	closeMutex sync.Mutex

	closeCallback closeCallback
	Context interface{}
}

func NewServerTcpConnection(conn *net.TCPConn, sendChanSize int, codec Codec, cb closeCallback) *TcpConnection {
	conn2 := &TcpConnection{
		conn:          conn,
		codec:         codec,
		closeChan:     make(chan int),
		id:            atomic.AddUint64(&globalConnectionId, 1),
		closeCallback: cb,
	}

	if sendChanSize > 0 {
		conn2.sendChan = make(chan interface{}, sendChanSize)
		go conn2.sendLoop()
	}
	return conn2
}

func NewClientTcpConnection(conn *net.TCPConn, sendChanSize int, codec Codec, cb closeCallback) *TcpConnection {
	conn2 := &TcpConnection{
		conn:          conn,
		codec:         codec,
		closeChan:     make(chan int),
		id:            atomic.AddUint64(&globalConnectionId, 1),
		closeCallback: cb,
	}

	if sendChanSize > 0 {
		conn2.sendChan = make(chan interface{}, sendChanSize)
		go conn2.sendLoop()
	}
	return conn2
}

func (c *TcpConnection) LoadAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *TcpConnection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (conn *TcpConnection) GetConnID() uint64 {
	return conn.id
}

func (conn *TcpConnection) IsClosed() bool {
	return atomic.LoadInt32(&conn.closeFlag) == 1
}

func (conn *TcpConnection) Close() error {
	if atomic.CompareAndSwapInt32(&conn.closeFlag, 0, 1) {
		if conn.closeCallback != nil {
			conn.closeCallback.OnConnectionClosed(conn)
		}

		close(conn.closeChan)

		if conn.sendChan != nil {
			conn.sendMutex.Lock()
			close(conn.sendChan)
			if clear, ok := conn.codec.(ClearSendChan); ok {
				clear.ClearSendChan(conn.sendChan)
			}
			conn.sendMutex.Unlock()
		}

		err := conn.codec.Close()
		return err
	}
	return ConnectionClosedError
}

func (conn *TcpConnection) Codec() Codec {
	return conn.codec
}

func (conn *TcpConnection) Receive() (interface{}, error) {
	conn.recvMutex.Lock()
	defer conn.recvMutex.Unlock()

	msg, err := conn.codec.Receive()
	if err != nil {
		conn.Close()
	}
	return msg, err
}

func (conn *TcpConnection) sendLoop() {
	defer conn.Close()
	for {
		select {
		case msg, ok := <-conn.sendChan:
			if !ok || conn.codec.Send(msg) != nil {
				return
			}
		case <-conn.closeChan:
			return
		}
	}
}

func (conn *TcpConnection) Send(msg interface{}) error {
	if conn.sendChan == nil {
		if conn.IsClosed() {
			return ConnectionClosedError
		}

		conn.sendMutex.Lock()
		defer conn.sendMutex.Unlock()

		err := conn.codec.Send(msg)
		if err != nil {
			conn.Close()
		}
		return err
	}

	conn.sendMutex.RLock()
	if conn.IsClosed() {
		conn.sendMutex.RUnlock()
		return ConnectionClosedError
	}

	select {
	case conn.sendChan <- msg:
		conn.sendMutex.RUnlock()
		return nil
	default:
		conn.sendMutex.RUnlock()
		conn.Close()
		return ConnectionBlockedError
	}
}
