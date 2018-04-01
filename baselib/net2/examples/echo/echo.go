/*
 *  Copyright (c) 2017, https://github.com/nebulaim
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

package main

import (
	"github.com/nebulaim/telegramd/baselib/net2"
	"github.com/golang/glog"
	"net"
	"github.com/nebulaim/telegramd/baselib/app"
	"github.com/nebulaim/telegramd/baselib/net2/codec"
)

func init() {
	net2.RegisterPtotocol("echo", codec.NewLengthBasedFrame(1024))
}

type EchoServer struct {
	server* net2.TcpServer
}

func NewEchoServer(listener net.Listener, protoName string) *EchoServer {
	//listener, err := net.Listen("tcp", "0.0.0.0:12345")
	//if err != nil {
	//	glog.Fatalf("listen error: %v", err)
	//	// return
	//}
	s := &EchoServer{}
	s.server = net2.NewTcpServer(listener, "echo", protoName, 1, s)
	return s
}

func (s* EchoServer) Serve() {
	s.server.Serve()
}

func (s *EchoServer) OnNewConnection(conn *net2.TcpConnection) {
	glog.Infof("OnNewConnection %v", conn.RemoteAddr())
}

func (s *EchoServer) OnConnectionDataArrived(conn *net2.TcpConnection, msg interface{}) error {
	glog.Infof("echo_server recv peer(%v) data: %v", conn.RemoteAddr(), msg)
	conn.Send(msg)
	return nil
}

func (s *EchoServer) OnConnectionClosed(conn *net2.TcpConnection) {
	glog.Infof("OnConnectionClosed - %v", conn.RemoteAddr())
}

type EchoClient struct {
	client *net2.TcpClientGroupManager
}

func NewEchoClient(protoName string, clients map[string][]string) *EchoClient {
	//listener, err := net.Listen("tcp", "0.0.0.0:12345")
	//if err != nil {
	//	glog.Fatalf("listen error: %v", err)
	//	// return
	//}
	c := &EchoClient{}
	c.client = net2.NewTcpClientGroupManager(protoName, clients, c)
	return c
}

func (c* EchoClient) Serve() {
	c.client.Serve()
}

func (c* EchoClient) OnNewClient(client *net2.TcpClient) {
	glog.Infof("OnNewConnection")
	client.Send("ping\n")
}

func (c* EchoClient) OnClientDataArrived(client *net2.TcpClient, msg interface{}) error {
	glog.Infof("OnDataArrived - recv data: %v", msg)
	return client.Send("ping\n")
}

func (c* EchoClient) OnClientClosed(client *net2.TcpClient) {
	glog.Infof("OnConnectionClosed")
}

func (c* EchoClient) OnClientTimer(client *net2.TcpClient) {
	glog.Infof("OnTimer")
}

type EchoInsance struct {
	server *EchoServer
	client *EchoClient
}

func (this *EchoInsance) Initialize() error {
	listener, err := net.Listen("tcp", "0.0.0.0:22345")
	if err != nil {
		glog.Errorf("listen error: %v", err)
		return err
	}

	this.server = NewEchoServer(listener, "echo")

	clients := map[string][]string{
		"echo":[]string{"127.0.0.1:22345", "192.168.1.101:22345"},
	}
	this.client = NewEchoClient("echo", clients)
	return nil
}

func (this *EchoInsance) RunLoop() {
	go this.server.Serve()
	this.client.Serve()
}

func (this *EchoInsance) Destroy() {
	this.client.client.Stop()
	this.server.server.Stop()
}

func main() {
	instance := &EchoInsance{}
	// app.AppInstance(instance)
	app.DoMainAppInsance(instance)
}
