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
	"sync"
	"github.com/golang/glog"
	"errors"
	"math/rand"
)

type TcpClientGroupManager struct {
	protoName string
	clientMapLock sync.RWMutex
	clientMap map[string]map[string]*TcpClient
	callback TcpClientCallBack
}

func NewTcpClientGroupManager(protoName string, clients map[string][]string, cb TcpClientCallBack) *TcpClientGroupManager  {
	group := &TcpClientGroupManager{
		protoName: protoName,
		clientMap: make(map[string]map[string]*TcpClient),
		callback: cb,
	}

	for k, v := range clients {
		m := make(map[string]*TcpClient)

		for _, address := range v {
			client := NewTcpClient(k, 10*1024, group.protoName, address, group.callback)
			if client != nil {
				m[address] = client
			}
		}

		group.clientMapLock.Lock()
		group.clientMap[k] = m
		group.clientMapLock.Unlock()
	}

	glog.Info("NewTcpClientGroup group : ", group.clientMap)
	return group
}

func (this *TcpClientGroupManager) Serve() bool {
	this.clientMapLock.Lock()
	defer this.clientMapLock.Unlock()

	for _, v := range this.clientMap {
		for _, c := range v {
			c.Serve()
		}
	}

	return true
}

func (this *TcpClientGroupManager) Stop() bool {
	this.clientMapLock.Lock()
	defer this.clientMapLock.Unlock()

	for _, v := range this.clientMap {
		for _, c := range v {
			c.Stop()
		}

	}
	return true
}

func (this *TcpClientGroupManager) GetConfig() interface{} {
	return nil
}

func (this *TcpClientGroupManager) AddClient(name string, address string) {
	glog.Info("TcpClientGroup AddClient name ", name, " address ", address)
	this.clientMapLock.Lock()
	defer this.clientMapLock.Unlock()

	m, ok := this.clientMap[name]

	if !ok {
		this.clientMap[name] = make(map[string]*TcpClient)
	}

	m, _ = this.clientMap[name]

	_, ok = m[address]

	if ok {
		return
	}

	client := NewTcpClient(name, 10 * 1024, this.protoName, address, this.callback)

	m[address] = client

	client.Serve()
}

func (this *TcpClientGroupManager) RemoveClient(name string, address string) {
	glog.Info("TcpClientGroup RemoveClient name ", name, " address ", address)

	this.clientMapLock.Lock()
	defer this.clientMapLock.Unlock()

	m, ok := this.clientMap[name]

	if !ok {
		return
	}

	m, _ = this.clientMap[name]

	c, ok := m[address]

	if !ok {
		return
	}

	c.Stop()

	delete(this.clientMap[name], address)
}

func (this *TcpClientGroupManager) SendData(name string, msg interface{}) error {
	tcpConn := this.getRotationSession(name)
	if tcpConn == nil {
		return errors.New("Can not get connection!!")
	}
	return tcpConn.Send(msg)
}

func (this *TcpClientGroupManager) getRotationSession(name string) *TcpConnection {
	all_conns := this.getTcpClientsByName(name)
	if all_conns == nil || len(all_conns) == 0 {
		return nil
	}

	index := rand.Int() % len(all_conns)
	return all_conns[index]
}

func (this *TcpClientGroupManager) BroadcastData (name string, msg interface{}) error {
	all_conns := this.getTcpClientsByName(name)

	if all_conns == nil || len(all_conns) == 0 {
		return nil
	}

	for _, conn := range all_conns {
		conn.Send(msg)
	}

	return nil
}

func (this *TcpClientGroupManager) getTcpClientsByName(name string) []*TcpConnection {
	var all_conns []*TcpConnection

	this.clientMapLock.RLock()

	serviceMap, ok := this.clientMap[name]

	if !ok {
		this.clientMapLock.RUnlock()
		return nil
	}

	for _, c := range serviceMap {
		if c != nil && c.GetConnection() != nil && !c.GetConnection().IsClosed() {
			all_conns = append(all_conns, c.GetConnection())
		}
	}

	this.clientMapLock.RUnlock()

	return all_conns
}
