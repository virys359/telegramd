/*
 * Copyright (c) 2018-present, Yumcoder, LLC.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree.
 */
package test

import (
	"github.com/golang/glog"
	"github.com/nebulaim/telegramd/baselib/net2"
)

type TestClient struct {
	client   *net2.TcpClient
	ready    chan bool
	receiver chan interface{}
}

func NewTestClient(RemoteName, protoName, remoteAddress string, receiver chan interface{}, chanSize int) *TestClient {
	c := &TestClient{}
	c.client = net2.NewTcpClient(RemoteName, chanSize, protoName, remoteAddress, c)
	c.ready = make(chan bool)
	c.receiver = receiver
	return c
}

func (c *TestClient) Name() string {
	return c.client.GetRemoteName()
}

func (c *TestClient) Serve() {
	c.client.Serve()
}

func (c *TestClient) Stop() {
	c.client.Stop()
}

func (c *TestClient) Send(msg interface{}) error {
	return c.client.Send(msg)
}

func (c *TestClient) OnNewClient(client *net2.TcpClient) {
	glog.Infof("client OnNewConnection %v", client.GetRemoteName())
	c.ready <- true
}

func (c *TestClient) OnClientDataArrived(client *net2.TcpClient, msg interface{}) error {
	glog.Infof("client OnClientDataArrived - client: %s, receive data: %v", client.GetRemoteName(), msg)
	c.receiver <- msg
	return nil
}

func (c *TestClient) OnClientClosed(client *net2.TcpClient) {
	glog.Infof("client OnConnectionClosed %s", client.GetRemoteName())
}

func (c *TestClient) OnClientTimer(client *net2.TcpClient) {
	glog.Infof("OnTimer")
}
