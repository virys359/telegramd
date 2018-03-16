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
	"github.com/BurntSushi/toml"
	"sync"
	"encoding/binary"
	"time"
)

type ServerConfig struct {
	Name      string
	ProtoName string
	Addr      string
}

func newTcpServer(config *ServerConfig, cb net2.TcpConnectionCallback) (*net2.TcpServer, error) {
	lsn, err := net.Listen("tcp", config.Addr)
	if err != nil {
		// glog.Errorf("listen error: %v", err)
		return nil, err
	}
	server := net2.NewTcpServer(lsn, config.Name, config.ProtoName, 1024, cb)
	return server, nil
}

type FrontendConfig struct {
	ServerId  int32 // 服务器ID
	Server80  *ServerConfig
	Server443 *ServerConfig
}

//const (
//	STATE_UNKNOWN = iota
//	STATE_CONNECTED
//
//	STATE_pq
//	STATE_pq_res
//	STATE_pq_ack
//
//	STATE_DH_params
//	STATE_DH_params_res
//	STATE_DH_params_ack
//
//	STATE_dh_gen
//	STATE_dh_gen_res
//	STATE_dh_gen_ack
//
//	STATE_HANDSHAKE
//	STATE_AUTH_KEY
//	STATE_ERROR
//)
//
//const (
//	RES_STATE_UNKNOWN = iota
//	RES_STATE_NONE
//	RES_STATE_OK
//	RES_STATE_ERROR
//)

func isHandshake(state int) bool {
	return state >= mtproto.STATE_CONNECTED2 && state <= mtproto.STATE_dh_gen_ack
}

type handshakeState struct {
	state     int		// 状态
	resState  int		// 后端握手返回的结果
	ctx		  []byte	// 握手上下文数据，透传给后端
}

type connContext struct {
	// TODO(@benqi): lock
	sync.Mutex
	state          int // 是否握手阶段
	md             *mtproto.ZProtoMetadata
	handshakeState *mtproto.HandshakeState
}

func (ctx *connContext) getState() int {
	ctx.Lock()
	defer ctx.Unlock()
	return ctx.state
}

func (ctx *connContext) setState(state int) {
	ctx.Lock()
	defer ctx.Unlock()
	if ctx.state != state {
		ctx.state = state
	}
}

func (ctx *connContext) encryptedMessageAble() bool {
	ctx.Lock()
	defer ctx.Unlock()
	//return ctx.state == mtproto.STATE_CONNECTED2 ||
	//	ctx.state == mtproto.STATE_AUTH_KEY ||
	//	(ctx.state == mtproto.STATE_HANDSHAKE &&
	//		(ctx.handshakeState.State == mtproto.STATE_pq_ack ||
	//		(ctx.handshakeState.State == mtproto.STATE_dh_gen_ack &&
	//			ctx.handshakeState.ResState == mtproto.RES_STATE_OK)))
	return ctx.state == mtproto.STATE_CONNECTED2 ||
		ctx.state == mtproto.STATE_AUTH_KEY ||
		(ctx.state == mtproto.STATE_HANDSHAKE &&
			(ctx.handshakeState.State == mtproto.STATE_pq_res ||
				(ctx.handshakeState.State == mtproto.STATE_dh_gen_res &&
					ctx.handshakeState.ResState == mtproto.RES_STATE_OK)))

}

type FrontendServer struct {
	configPath string
	config     *FrontendConfig

	// TODO(@benqi): manager server80 and server443
	server80   *net2.TcpServer
	server443  *net2.TcpServer
	client     *net2.TcpClientGroupManager
}

func NewFrontendServer(configPath string) *FrontendServer {
	return &FrontendServer{
		configPath: configPath,
		config:     &FrontendConfig{},
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// AppInstance interface
func (s *FrontendServer) Initialize() error {
	var err error

	if _, err = toml.DecodeFile(s.configPath, s.config); err != nil {
		glog.Errorf("decode config file %s error: %v", s.configPath, err)
		return err
	}

	if s.config.Server80 == nil && s.server443 == nil {
		err := fmt.Errorf("config error, server80 and server443 is nil, in config: %v", s.config)
		return err
	}

	glog.Info(s.config)

	s.server80, err = newTcpServer(s.config.Server80, s)
	if err != nil {
		glog.Error(err)
		return err
	}

	s.server443, err = newTcpServer(s.config.Server443, s)
	if err != nil {
		glog.Error(err)
		return err
	}

	clients := map[string][]string{
		"session":[]string{"127.0.0.1:10000"},
	}
	s.client = net2.NewTcpClientGroupManager("zproto", clients, s)
	return nil
}

func (s *FrontendServer) RunLoop() {
	go s.server80.Serve()
	go s.server443.Serve()
	go s.client.Serve()
}

func (s *FrontendServer) Destroy() {
	s.server80.Stop()
	s.server443.Stop()
	s.client.Stop()
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// TcpConnectionCallback

func (s *FrontendServer) newMetadata(conn *net2.TcpConnection) *mtproto.ZProtoMetadata {
	md := &mtproto.ZProtoMetadata{
		ServerId: int(s.config.ServerId),
		ClientConnId: conn.GetConnID(),
		ClientAddr: conn.RemoteAddr().String(),
		From: "frontend",
		ReceiveTime: time.Now().Unix(),
	}
	md.SpanId, _ = id.NextId()
	md.TraceId, _ = id.NextId()
	return md
}

func (s *FrontendServer) OnNewConnection(conn *net2.TcpConnection) {
	// glog.Infof("OnNewConnection - peer(%v)", conn.RemoteAddr())
	conn.Context = &connContext{
		state: mtproto.STATE_CONNECTED2,
		md: &mtproto.ZProtoMetadata{
			ServerId: int(s.config.ServerId),
			ClientConnId: conn.GetConnID(),
			ClientAddr: conn.RemoteAddr().String(),
			From: "frontend",
		},
		handshakeState: &mtproto.HandshakeState{
			State:    mtproto.STATE_CONNECTED2,
			ResState: mtproto.RES_STATE_NONE,
		},
	}

	glog.Infof("onNewConnection - peer(%s), ctx: {%v}", conn, conn.Context)
}

func (s *FrontendServer) OnConnectionDataArrived(conn *net2.TcpConnection, msg interface{}) error {
	glog.Infof("onConnectionDataArrived - peer(%s) recv data", conn)

	ctx, _ := conn.Context.(*connContext)
	message, ok := msg.(*mtproto.MTPRawMessage)

	var err error
	if !ok {
		err = fmt.Errorf("invalid mtproto raw message: %v", msg)
		glog.Error(err)
		conn.Close()
		return err
	}

	if message.AuthKeyId == 0 {
		if ctx.getState() == mtproto.STATE_AUTH_KEY {
			err = fmt.Errorf("invalid state STATE_AUTH_KEY")
			glog.Error(err)
			conn.Close()
		} else {
			err = s.onUnencryptedRawMessage(ctx, conn, message)
		}
	} else {
		if !ctx.encryptedMessageAble() {
			err = fmt.Errorf("invalid state: {state: %d, handshakeState: {%v}}, peer(%s)", ctx.state, ctx.handshakeState, conn)
			glog.Error(err)
			conn.Close()
		} else {
			err = s.onEncryptedRawMessage(ctx, conn, message)
		}
	}

	return err
}

func (s *FrontendServer) OnConnectionClosed(conn *net2.TcpConnection) {
	glog.Infof("onConnectionClosed - peer(%s)", conn)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// TcpClientCallBack
func (s *FrontendServer) OnNewClient(client *net2.TcpClient) {
	glog.Infof("onNewClient - peer(%s)", client.GetConnection())
}

func (s *FrontendServer) OnClientDataArrived(client *net2.TcpClient, msg interface{}) error {
	glog.Infof("onClientDataArrived - peer(%s) recv data", client.GetConnection())

	zmsg, _ := msg.(*mtproto.ZProtoMessage)
	conn := s.server443.GetConnection(zmsg.SessionId)
	if conn == nil {
		glog.Warning("conn closed, connID = ", zmsg.SessionId)
		return nil
	}
	payload, _ := zmsg.Message.(*mtproto.ZProtoRawPayload)
	msgType := binary.LittleEndian.Uint32(payload.Payload)
	switch msgType {
	case mtproto.SESSION_HANDSHAKE:
		hmsg := &mtproto.ZProtoHandshakeMessage{
			State: &mtproto.HandshakeState{},
			MTPMessage: &mtproto.MTPRawMessage{},
		}
		hmsg.Decode(payload.Payload[4:])

		glog.Infof("handshake - state: {%v}", hmsg.State)
		ctx := conn.Context.(*connContext)
		ctx.Lock()
		ctx.handshakeState = hmsg.State
		ctx.Unlock()
		return conn.Send(hmsg.MTPMessage)
	case mtproto.SESSION_SESSION_DATA:
		smsg := &mtproto.ZProtoSessionData{
			MTPMessage: &mtproto.MTPRawMessage{},
		}
		smsg.Decode(payload.Payload[4:])
		return conn.Send(smsg.MTPMessage)
	default:
		err := fmt.Errorf("invalid zmsg: %v", zmsg)
		glog.Error(err)
		return err
	}
}

func (s *FrontendServer) OnClientClosed(client *net2.TcpClient) {
	glog.Infof("onClientClosed - peer(%s) recv data", client.GetConnection())

	if client.AutoReconnect() {
		client.Reconnect()
	}
}

func (s *FrontendServer) OnClientTimer(client *net2.TcpClient) {
	glog.Infof("OnTimer")
}


////////////////////////////////////////////////////////////////////////////////////////////////////
func (s *FrontendServer) onUnencryptedRawMessage(ctx *connContext, conn *net2.TcpConnection, mmsg *mtproto.MTPRawMessage) error {
	glog.Infof("onUnencryptedRawMessage - peer(%s) recv data", conn)
	ctx.Lock()
	if ctx.state == mtproto.STATE_CONNECTED2 {
		ctx.state = mtproto.STATE_HANDSHAKE
	}
	if ctx.handshakeState.State == mtproto.STATE_CONNECTED2 {
		ctx.handshakeState.State = mtproto.STATE_pq
	}
	ctx.Unlock()

	// sentToClient
	hmsg := &mtproto.ZProtoHandshakeMessage{
		State: ctx.handshakeState,
		MTPMessage: mmsg,
	}
	zmsg := &mtproto.ZProtoMessage{
		SessionId: conn.GetConnID(),
		SeqNum: 1,	// TODO(@benqi): gen seqNum
		Metadata: s.newMetadata(conn),
		Message:&mtproto.ZProtoRawPayload{
			Payload: hmsg.Encode(),
		},
	}
	// glog.Infof("sendToSessionClient: %v", zmsg)
	return s.client.SendData("session", zmsg)
}

func (s *FrontendServer) onEncryptedRawMessage(ctx *connContext, conn *net2.TcpConnection, mmsg *mtproto.MTPRawMessage) error {
	glog.Infof("onEncryptedRawMessage - peer(%s) recv data", conn)
	// sentToClient
	hmsg := &mtproto.ZProtoSessionData{
		MTPMessage: mmsg,
	}
	zmsg := &mtproto.ZProtoMessage{
		SessionId: conn.GetConnID(),
		SeqNum: 1,	// TODO(@benqi): gen seqNum
		Metadata: s.newMetadata(conn),
		Message:&mtproto.ZProtoRawPayload{
			Payload: hmsg.Encode(),
		},
	}
	return  s.client.SendData("session", zmsg)
}
