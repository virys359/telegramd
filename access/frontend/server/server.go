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
	"github.com/nebulaim/telegramd/baselib/net2/watcher2"
	"github.com/coreos/etcd/clientv3"
	// "github.com/nebulaim/telegramd/baselib/grpc_util"
	"github.com/nebulaim/telegramd/baselib/grpc_util/load_balancer"
	"github.com/nebulaim/telegramd/baselib/base"
	// "github.com/nebulaim/telegramd/baselib/sync2"
)

type ServerConfig struct {
	Name      string
	ProtoName string
	Addr      string
}

type ClientConfig struct {
	Name      string
	ProtoName string
	AddrList  []string
	EtcdAddrs []string
	Balancer  string
}

func newTcpServer(config *ServerConfig, cb net2.TcpConnectionCallback) (*net2.TcpServer, error) {
	lsn, err := net.Listen("tcp", config.Addr)
	if err != nil {
		// glog.Errorf("listen error: %v", err)
		return nil, err
	}
	server := net2.NewTcpServer(net2.TcpServerArgs{
		Listener:           lsn,
		ServerName:         config.Name,
		ProtoName:          "mtproto",
		SendChanSize:       1024,
		ConnectionCallback: cb,
	}) // todo(yumcoder): set max connection
	return server, nil
}

type FrontendConfig struct {
	ServerId      int32 // 服务器ID
	Server80      *ServerConfig
	Server443     *ServerConfig
	Server5222    *ServerConfig
	SessionClient *ClientConfig
	AuthKeyClient *ClientConfig
}

func (c *FrontendConfig) String() string {
	return fmt.Sprintf("{server_id: %d, server80: %v. server443: %v, server5222: %v, session_client: %v, auth_key_client: %v}",
		c.ServerId,
		c.Server80,
		c.Server443,
		c.Server5222,
		c.SessionClient,
		c.AuthKeyClient)
}

type handshakeState struct {
	state    int    // 状态
	resState int    // 后端握手返回的结果
	ctx      []byte // 握手上下文数据，透传给后端
}

type connContext struct {
	// TODO(@benqi): lock
	sync.Mutex
	state          int // 是否握手阶段
	md             *mtproto.ZProtoMetadata
	handshakeState *mtproto.HandshakeState
	seqNum         uint64

	sessionAddr    string
	authKeyId      int64
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
	configPath     string
	config         *FrontendConfig
	server80       *net2.TcpServer
	server443      *net2.TcpServer
	server5222     *net2.TcpServer
	clientManager  *net2.TcpClientGroupManager
	sessionWatcher *watcher2.ClientWatcher
	ketama         *load_balancer.Ketama
	//authKeyClient        *net2.TcpClientGroupManager
	authKeyClientWatcher *watcher2.ClientWatcher
}

func NewFrontendServer(configPath string) *FrontendServer {
	return &FrontendServer{
		configPath: configPath,
		config:     &FrontendConfig{},
		ketama:     load_balancer.NewKetama(10, nil),
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

	glog.Infof("config loaded: %v", s.config)

	if s.config.Server80 == nil && s.server443 == nil {
		err := fmt.Errorf("config error, server80 and server443 is nil, in config: %v", s.config)
		return err
	}
	// glog.Info(s.config)

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

	s.server5222, err = newTcpServer(s.config.Server5222, s)
	if err != nil {
		glog.Error(err)
		return err
	}

	clients := map[string][]string{
		// "session": s.config.SessionClient.AddrList,
		// s.config.SessionClient.Name: s.config.SessionClient.AddrList,
	}
	s.clientManager = net2.NewTcpClientGroupManager(s.config.SessionClient.ProtoName, clients, s)

	// session service discovery
	etcdConfg := clientv3.Config{
		Endpoints: s.config.SessionClient.EtcdAddrs,
	}
	s.sessionWatcher, _ = watcher2.NewClientWatcher("/nebulaim", "session", etcdConfg, s.clientManager)

	///////////////////////////////////////
	//s.authKeyClient = net2.NewTcpClientGroupManager(s.config.AuthKeyClient.ProtoName, clients, s)
	etcdConfg2 := clientv3.Config{
		Endpoints: s.config.AuthKeyClient.EtcdAddrs,
	}
	s.authKeyClientWatcher, _ = watcher2.NewClientWatcher("/nebulaim", "handshake", etcdConfg2, s.clientManager)

	return nil
}

func (s *FrontendServer) RunLoop() {
	go s.server80.Serve2()
	go s.server443.Serve2()
	go s.server5222.Serve2()

	// go s.clientManager.Serve()
	go s.authKeyClientWatcher.WatchClients(nil)
	go s.sessionWatcher.WatchClients(func(etype, addr string) {
		switch etype {
		case "add":
			s.ketama.Add(addr)
		case "delete":
			s.ketama.Remove(addr)
		}
	})
}

func (s *FrontendServer) Destroy() {
	s.server80.Stop()
	s.server443.Stop()
	s.server5222.Stop()
	s.clientManager.Stop()
	// s.authKeyClient.Stop()
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// TcpConnectionCallback

func (s *FrontendServer) newMetadata(conn *net2.TcpConnection) *mtproto.ZProtoMetadata {
	md := &mtproto.ZProtoMetadata{
		ServerId:     int(s.config.ServerId),
		ClientConnId: conn.GetConnID(),
		ClientAddr:   conn.RemoteAddr().String(),
		From:         "frontend",
		ReceiveTime:  time.Now().Unix(),
	}
	md.SpanId, _ = id.NextId()
	md.TraceId, _ = id.NextId()
	return md
}

func (s *FrontendServer) OnNewConnection(conn *net2.TcpConnection) {
	// glog.Infof("OnNewConnection - peer(%v)", conn.RemoteAddr())
	// TODO(@benqi): peekCodec
	err := conn.Codec().(*mtproto.MTProtoProxyCodec).PeekCodec()
	if err != nil {
		glog.Error(err)
		conn.Close()
		return
	}

	conn.Context = &connContext{
		state: mtproto.STATE_CONNECTED2,
		md: &mtproto.ZProtoMetadata{
			ServerId:     int(s.config.ServerId),
			ClientConnId: conn.GetConnID(),
			ClientAddr:   conn.RemoteAddr().String(),
			From:         "frontend",
		},
		handshakeState: &mtproto.HandshakeState{
			State:    mtproto.STATE_CONNECTED2,
			ResState: mtproto.RES_STATE_NONE,
		},
		seqNum: 1,
	}
	glog.Infof("onNewConnection - peer(%s), ctx: {%v}", conn, conn.Context)
}

func (s *FrontendServer) OnConnectionDataArrived(conn *net2.TcpConnection, msg interface{}) error {
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
	// ctx, _ := conn.Context.(*connContext)
	s.sendClientClosed(conn)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// TcpClientCallBack
func (s *FrontendServer) OnNewClient(client *net2.TcpClient) {
	glog.Infof("onNewClient - peer(%s)", client.GetConnection())
}

func (s *FrontendServer) genSessionId(conn *net2.TcpConnection) uint64 {
	var sid = conn.GetConnID()
	if conn.Name() == "frontend443" {
		// sid = sid | 0 << 56
	} else if conn.Name() == "frontend80" {
		sid = sid | 1 << 56
	} else if conn.Name() == "frontend5222" {
		sid = sid | 2 << 56
	}

	return sid
}

func (s *FrontendServer) getConnBySessionID(id uint64) *net2.TcpConnection {
	//
	var server *net2.TcpServer
	sid := id >> 56
	if sid == 0 {
		server = s.server443
	} else if sid == 1 {
		server = s.server80
	} else if sid == 2 {
		server = s.server5222
	} else {
		return nil
	}

	id = id & 0xffffffffffffff
	return server.GetConnection(id)
}

func (s *FrontendServer) OnClientDataArrived(client *net2.TcpClient, msg interface{}) error {
	zmsg, _ := msg.(*mtproto.ZProtoMessage)

	///////////////////////////////////////////////////////////////////
	conn := s.getConnBySessionID(zmsg.SessionId)
	// s.server443.GetConnection(zmsg.SessionId)
	if conn == nil {
		glog.Warning("conn closed, connID = ", zmsg.SessionId)
		return nil
	}

	payload, _ := zmsg.Message.(*mtproto.ZProtoRawPayload)
	msgType := binary.LittleEndian.Uint32(payload.Payload)
	switch msgType {
	case mtproto.SESSION_HANDSHAKE:
		hmsg := &mtproto.ZProtoHandshakeMessage{
			State:      &mtproto.HandshakeState{},
			MTPMessage: &mtproto.MTPRawMessage{},
		}
		hmsg.Decode(payload.Payload[4:])

		glog.Infof("onClientDataArrived - handshake: peer(%s), state: {%v}",
			client.GetConnection(),
			hmsg.State)

		if hmsg.State.ResState == mtproto.RES_STATE_ERROR {
			// TODO(@benqi): Close.
			conn.Close()
			return nil
		} else {
			ctx := conn.Context.(*connContext)
			ctx.Lock()
			ctx.handshakeState = hmsg.State
			ctx.Unlock()
			return conn.Send(hmsg.MTPMessage)
		}
	case mtproto.SESSION_SESSION_DATA:
		smsg := &mtproto.ZProtoSessionData{
			MTPMessage: &mtproto.MTPRawMessage{},
		}
		smsg.Decode(payload.Payload[4:])

		glog.Infof("onClientDataArrived - send clientManager to: peer(%s), auth_key_id: %d, len: %d",
			conn,
			smsg.MTPMessage.AuthKeyId,
			len(payload.Payload))

		conn = s.getConnBySessionID(zmsg.SessionId)
		if conn != nil {
			err := conn.Send(smsg.MTPMessage)
			if err != nil {
				glog.Info("onClientDataArrived: send data to clientManager error: ", err)
			}
			return err
		}
		return nil
	default:
		err := fmt.Errorf("invalid zmsg: %v", zmsg)

		glog.Errorf("onClientDataArrived - invalid zmsg: peer(%s), zmsg: {%v}",
			client.GetConnection(),
			zmsg)

		return err
	}
}

func (s *FrontendServer) OnClientClosed(client *net2.TcpClient) {
	glog.Infof("onClientClosed - peer(%s)", client.GetConnection())

	if client.AutoReconnect() {
		client.Reconnect()
	}
}

func (s *FrontendServer) OnClientTimer(client *net2.TcpClient) {
	glog.Infof("onClientTimer")
}

////////////////////////////////////////////////////////////////////////////////////////////////////
func (s *FrontendServer) onUnencryptedRawMessage(ctx *connContext, conn *net2.TcpConnection, mmsg *mtproto.MTPRawMessage) error {
	glog.Infof("onUnencryptedRawMessage - peer(%s) recv data, len = %d, ctx: %v", conn, len(mmsg.Payload), ctx)

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
		State:      ctx.handshakeState,
		MTPMessage: mmsg,
	}

	zmsg := &mtproto.ZProtoMessage{
		SessionId: s.genSessionId(conn), // conn.GetConnID(),
		SeqNum:    ctx.seqNum, // TODO(@benqi): gen seqNum
		Metadata:  s.newMetadata(conn),
		Message:   &mtproto.ZProtoRawPayload{
			Payload: hmsg.Encode(),
		},
	}

	ctx.seqNum++
	// glog.Infof("sendToSessionClient: %v", zmsg)
	return s.clientManager.SendData("handshake", zmsg)
}

func (s *FrontendServer) onEncryptedRawMessage(ctx *connContext, conn *net2.TcpConnection, mmsg *mtproto.MTPRawMessage) error {
	glog.Infof("onEncryptedRawMessage - peer(%s) recv data, len = %d, auth_key_id = %d", conn, len(mmsg.Payload), mmsg.AuthKeyId)

	if kaddr, ok := s.ketama.Get(base.Int64ToString(mmsg.AuthKeyId)); ok {
		s.checkAndSendClientNew(ctx, conn, kaddr, mmsg.AuthKeyId)

		// sentToClient
		hmsg := &mtproto.ZProtoSessionData{
			MTPMessage: mmsg,
		}
		zmsg := &mtproto.ZProtoMessage{
			SessionId: s.genSessionId(conn), // conn.GetConnID(),
			SeqNum:    ctx.seqNum,
			Metadata:  s.newMetadata(conn),
			Message:   &mtproto.ZProtoRawPayload{
				Payload: hmsg.Encode(),
			},
		}
		ctx.seqNum++
		return s.clientManager.SendDataToAddress("session", kaddr, zmsg)
	} else {
		return fmt.Errorf("kaddr not exists")
	}
}

func (s *FrontendServer) checkAndSendClientNew(ctx *connContext, conn *net2.TcpConnection, kaddr string, authKeyId int64) error {
	var err error
	if ctx.sessionAddr == "" {
		hmsg := &mtproto.ZProtoSessionClientNew{
			// MTPMessage: mmsg,
		}
		zmsg := &mtproto.ZProtoMessage{
			SessionId: s.genSessionId(conn), // conn.GetConnID(),
			SeqNum:    ctx.seqNum,
			Metadata:  s.newMetadata(conn),
			Message:   &mtproto.ZProtoRawPayload{
				Payload: hmsg.Encode(),
			},
		}
		ctx.seqNum++
		err = s.clientManager.SendDataToAddress("session", kaddr, zmsg)
		if err == nil {
			ctx.sessionAddr = kaddr
			ctx.authKeyId = authKeyId
		} else {
			glog.Error(err)
		}
	} else {
		// TODO(@benqi): check ctx.sessionAddr == kaddr
	}

	return err
}

func (s *FrontendServer) sendClientClosed(conn *net2.TcpConnection) {
	if conn.Context == nil {
		return
	}

	ctx, _ := conn.Context.(*connContext)
	if ctx.sessionAddr == "" || ctx.authKeyId == 0 {
		return
	}

	var err error
	if kaddr, ok := s.ketama.Get(base.Int64ToString(ctx.authKeyId)); ok && kaddr == ctx.sessionAddr {
		hmsg := &mtproto.ZProtoSessionClientClosed{
			// MTPMessage: mmsg,
		}
		zmsg := &mtproto.ZProtoMessage{
			SessionId: s.genSessionId(conn), // conn.GetConnID(),
			SeqNum:    ctx.seqNum,
			Metadata:  s.newMetadata(conn),
			Message:   &mtproto.ZProtoRawPayload{
				Payload: hmsg.Encode(),
			},
		}
		ctx.seqNum++
		err = s.clientManager.SendDataToAddress("session", ctx.sessionAddr, zmsg)
		if err != nil {
			glog.Error(err)
		}
	}
}
