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
	"bufio"
	"io"
	"github.com/nebulaim/telegramd/baselib/app"
)

const (
	kDedaultReadBufferSize = 1024
)

func init() {
	net2.RegisterPtotocol("echo", NewEcho(kDedaultReadBufferSize))
}

func NewEcho(bufLen int) net2.Protocol {
	if bufLen <= 0 {
		bufLen = kDedaultReadBufferSize
	}

	return &Echo{
		readBuf: bufLen,
	}
}

type Echo struct {
	readBuf  int
}

func (b *Echo) NewCodec(rw io.ReadWriter) (net2.Codec, error) {
	codec := new(EchoCodec)
	codec.conn = rw.(*net.TCPConn)
	codec.r = bufio.NewReaderSize(rw, b.readBuf)
	return codec, nil
}

type EchoCodec struct {
	conn *net.TCPConn
	r    *bufio.Reader
}

func (c *EchoCodec) Send(msg interface{}) error {
	buf := []byte(msg.(string))

	if _, err := c.conn.Write(buf); err != nil {
		return err
	}

	return nil
}

func (c *EchoCodec) Receive() (interface{}, error) {
	line, err := c.r.ReadString('\n');
	return line, err
}

func (c *EchoCodec) Close() error {
	return c.conn.Close()
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

func (s *EchoServer) OnDataArrived(conn *net2.TcpConnection, msg interface{}) error {
	glog.Infof("OnDataArrived - %v", msg)
	conn.Send(msg)
	return nil
}

func (s *EchoServer) OnConnectionClosed(conn *net2.TcpConnection) {
	glog.Infof("OnConnectionClosed - %v", conn.RemoteAddr())
}

type EchoClient struct {
	client* net2.TcpClient
}

func NewEchoClient(protoName string) *EchoClient {
	//listener, err := net.Listen("tcp", "0.0.0.0:12345")
	//if err != nil {
	//	glog.Fatalf("listen error: %v", err)
	//	// return
	//}
	c := &EchoClient{}
	c.client = net2.NewTcpClient("", 1, protoName, "127.0.0.1:22345", c)
	return c
}

func (c* EchoClient) Serve() {
	c.client.Serve()
}

func (c* EchoClient) OnNewConnection(client *net2.TcpClient) {
	glog.Infof("OnNewConnection")
	client.Send("ping\n")
}

func (c* EchoClient) OnDataArrived(client *net2.TcpClient, msg interface{}) error {
	glog.Infof("OnDataArrived - recv data: %v", msg)
	return client.Send("ping\n")
}

func (c* EchoClient) OnConnectionClosed(client *net2.TcpClient) {
	glog.Infof("OnConnectionClosed")
}

func (c* EchoClient) OnTimer(client *net2.TcpClient) {
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
	this.client = NewEchoClient("echo")
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
