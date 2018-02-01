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

package server

import (
	"github.com/golang/glog"
	"github.com/nebulaim/telegramd/baselib/net2"
	"net"
	"github.com/nebulaim/telegramd/mtproto"
	"fmt"
)

type AuthKeyServer struct {
	server* net2.TcpServer
}

func NewFrontendServer(listener net.Listener, protoName string) *AuthKeyServer {
	listener, err := net.Listen("tcp", "0.0.0.0:12345")
	if err != nil {
		glog.Fatalf("listen error: %v", err)
		// return
	}
	s := &AuthKeyServer{}
	s.server = net2.NewTcpServer(listener, "frontend", protoName, 1024, s)
	return s
}

func (s* AuthKeyServer) Serve() {
	s.server.Serve()
}

func (s* AuthKeyServer) Stop() {
	s.server.Stop()
}

func (s *AuthKeyServer) OnNewConnection(conn *net2.TcpConnection) {
	// conn.Context = &ConnContext{}
	glog.Infof("OnNewConnection %v", conn.RemoteAddr())
}

func (s *AuthKeyServer) OnDataArrived(conn *net2.TcpConnection, msg interface{}) error {
	glog.Infof("echo_server recv peer(%v) data: %v", conn.RemoteAddr(), msg)
	// connContext, _ := conn.Context.(*ConnContext)
	// _ = connContext

	//switch msg.(type) {
	//case *mtproto.UnencryptedRawMessage:
	//case *mtproto.EncryptedRawMessage:
	//default:
	//	// 不可能发生，直接coredump吧
	//	err := fmt.Errorf("")
	//	glog.Error(err)
	//	return err
	//}
	//return conn.Send(msg)

	return nil
}

func (s *AuthKeyServer) OnConnectionClosed(conn *net2.TcpConnection) {
	glog.Infof("OnConnectionClosed - %v", conn.RemoteAddr())
}
