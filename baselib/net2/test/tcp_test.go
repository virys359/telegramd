/*
 * Copyright (c) 2018-present, Yumcoder, LLC.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree.
 */
package test

import (
	"testing"
	"net"
	"github.com/golang/glog"
	"fmt"
	"github.com/nebulaim/telegramd/baselib/net2"
	"flag"
)

func init() {
	 flag.Set("alsologtostderr", "true")
	 flag.Set("log_dir", "false")
}

type TestTcpSimulation struct {
	protoName     string
	// ----- server -----
	serverCnt      int
	serverChanSize int
	serverMaxConn  int
	// ----- client -----
	clientNo          int
	clientMsgCnt      int
	clientChanSizeCnt int

	receivedChan chan interface{}

	servers []*TestServer
	clients []*TestClient
}

func (ts *TestTcpSimulation) simulate() (result []string, err error) {
	listener, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		glog.Errorf("listen error: %v", err)
		return
	}

	for i := 0; i < ts.serverCnt; i ++ {
		server := NewTestServer(listener, fmt.Sprintf(`TestServer%d`, i), ts.protoName, ts.serverChanSize, ts.serverMaxConn)
		ts.servers = append(ts.servers, server)
		go server.Serve()
	}

	var clients []*TestClient
	for i := 0; i < ts.clientNo; i ++ {
		client := NewTestClient(fmt.Sprintf(`client%d`, i), ts.protoName, listener.Addr().String(), ts.receivedChan, ts.clientChanSizeCnt)
		clients = append(clients, client)
		go client.Serve()
		<-client.ready
	}
	ts.clients = append(ts.clients, clients...)

	errChan := make(chan error)

	for _, c := range clients {
		for i := 0; i < ts.clientMsgCnt; i++ {
			go func(tc *TestClient, n int) {
				if err = c.Send(fmt.Sprintf("%s(ping%d)\n", tc.Name(), n)); err != nil {
					errChan <- err
				}
			}(c, i)
		}
	}

	for cnt := 0; cnt < ts.clientNo * ts.clientMsgCnt; cnt ++{
		select {
		case msg, _ := <-ts.receivedChan:
			result = append(result, msg.(string))
		case m, _ := <-errChan:
			result = append(result, m.Error())
		}
	}

	return
}
func (ts *TestTcpSimulation) stop() {
	for _, s := range ts.servers {
		s.Stop()
	}
	for _, c := range ts.clients {
		c.Stop()
	}
}

// net2.tcpClient 																      net2.tcpServer
//     |     																	      |				|
//   --------------	   															-----------------	|
//  | net2.tcpCnn  |											  			   |   net2.tcpCnn   |	|
//   --------------   															-----------------	|
//   		|   															          |				|
//      -----------------------											   --------------------		|
//     |   chanSize (buffer)   |										  |  chanSize (buffer) |	|
//      -----------------------									           --------------------		|
//                |________________________net.TCPConn____________________________|					|
//																									|
//     .																		      				|
//     .																		     				|
//     .																		      				|
//																				                    |
// net2.tcpClient 																                    |
//     |     																	                    |
//   --------------	   															-----------------   |
//  | net2.tcpCnn  |											  			   |   net2.tcpCnn   |--
//   --------------   															-----------------
//   		|   															          |
//      -----------------------											   --------------------
//     |   chanSize (buffer)   |										  |  chanSize (buffer) |
//      -----------------------									           --------------------
//                |________________________net.TCPConn____________________________|
//
func TestClientServer(t *testing.T) {
	//remoteAddress := "127.0.0.1:12345"
	protoName := "TestCodec"

	net2.RegisterProtocol(protoName, &TestProto{})

	testTable := []struct {
		serverCnt         int
		serverChanSize    int
		serverMaxConn     int
		clientNo          int
		clientMsgCnt      int
		clientChanSizeCnt int
	}{
		{
			// 1 server, 1 client
			clientNo:          1,
			clientMsgCnt:      10,
			clientChanSizeCnt: 10,
			serverCnt:         1,
			serverChanSize:    10,
			serverMaxConn:     5,
		},
		{
			// 1 server, 2 client
			clientNo:          1,
			clientMsgCnt:      1000, // todo: why waite? what's the relation between parameters(msgcnt and chanSize and influence on request per second)?
			clientChanSizeCnt: 10,
			serverCnt:         1,
			serverChanSize:    10,
			serverMaxConn:     5,
		},
	}

	for _, data := range testTable {

		receivedChan := make(chan interface{}, 0)
		simulation := &TestTcpSimulation{
			protoName:         protoName,
			serverCnt:         data.serverCnt,
			serverChanSize:    data.serverChanSize,
			serverMaxConn:     data.serverMaxConn,
			clientNo:          data.clientNo,
			clientMsgCnt:      data.clientMsgCnt,
			clientChanSizeCnt: data.clientChanSizeCnt,
			receivedChan:      receivedChan,
		}

		result, err := simulation.simulate()
		if err != nil {
			t.Error(err)
		}

		close(receivedChan)
		simulation.stop()

		expected := make(map[string]bool)
		for c := 0; c < simulation.clientNo; c++ {
			for i := 0; i < simulation.clientMsgCnt; i++ {
				expected[fmt.Sprintf("TestServer0(pong), client%d(ping%d)\n", c, i)] = false
			}
		}

		for _, v := range result {
			expected[v] = true
		}

		for k, v := range expected {
			if v == false {
				t.Errorf("expected received msg: %s", k)
			}
		}
	}
}

// todo: config chanSeize
// todo: group client
