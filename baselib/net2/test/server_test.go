/*
 * Copyright (c) 2018-present, Yumcoder, LLC.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree.
 */
package test

import (
	"net"
	"github.com/golang/glog"
	"github.com/nebulaim/telegramd/baselib/net2"
	"fmt"
)

type TestServer struct {
	server *net2.TcpServer
	serverName string
}

func NewTestServer(listener net.Listener, serverName, protoName string, chanSize int, maxConn int) *TestServer {
	s := &TestServer{}
	s.server = net2.NewTcpServer(
		net2.TcpServerArgs{
			Listener:                listener,
			ServerName:              serverName,
			ProtoName:               protoName,
			SendChanSize:            chanSize,
			ConnectionCallback:      s,
			MaxConcurrentConnection: maxConn,
		})
	s.serverName = serverName
	return s
}

func (s *TestServer) Serve() {
	s.server.Serve()
}

func (s *TestServer) Stop() {
	s.server.Stop()
}

func (s *TestServer) OnNewConnection(conn *net2.TcpConnection) {
	glog.Infof("server OnNewConnection %v", conn.String())
}

func (s *TestServer) OnConnectionDataArrived(conn *net2.TcpConnection, msg interface{}) error {
	glog.Infof("%s server receive peer(%v) data: %v", s.serverName, conn.RemoteAddr(), msg)
	conn.Send(fmt.Sprintf("%s(pong), %s", conn.Name(), msg))
	return nil
}

func (s *TestServer) OnConnectionClosed(conn *net2.TcpConnection) {
	glog.Infof("server OnConnectionClosed - %v", conn.RemoteAddr())
}
