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
	"github.com/nebulaim/telegramd/baselib/net2"
	"github.com/BurntSushi/toml"
	"github.com/golang/glog"
	"github.com/nebulaim/telegramd/baselib/redis_client"
	"github.com/nebulaim/telegramd/baselib/mysql_client"
	"github.com/nebulaim/telegramd/baselib/grpc_util"
	"github.com/nebulaim/telegramd/baselib/grpc_util/service_discovery"
	"github.com/nebulaim/telegramd/biz/dal/dao"
	"google.golang.org/grpc"
	"github.com/nebulaim/telegramd/mtproto"
	"time"
	"github.com/gogo/protobuf/proto"
	"fmt"
	"encoding/binary"
)

func init() {
	proto.RegisterType((*mtproto.ConnectToSessionServerReq)(nil), "mtproto.ConnectToSessionServerReq")
	proto.RegisterType((*mtproto.SessionServerConnectedRsp)(nil), "mtproto.SessionServerConnectedRsp")
	proto.RegisterType((*mtproto.PushUpdatesData)(nil), "mtproto.PushUpdatesData")
	proto.RegisterType((*mtproto.VoidRsp)(nil), "mtproto.VoidRsp")
}

type rpcServerConfig struct {
	Addr string
}

type syncConfig struct {
	Server 		*rpcServerConfig
	Discovery service_discovery.ServiceDiscoveryServerConfig
	Redis 		[]redis_client.RedisConfig
	Mysql     []mysql_client.MySQLConfig
}

type syncServer struct {
	configPath string
	config     *syncConfig
	client     *net2.TcpClientGroupManager
	server     *grpc_util.RPCServer
	impl       *SyncServiceImpl
}

func NewSyncServer(configPath string) *syncServer {
	return &syncServer{
		configPath: configPath,
		config:     &syncConfig{},
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// AppInstance interface
func (s *syncServer) Initialize() error {
	var err error

	if _, err = toml.DecodeFile(s.configPath, s.config); err != nil {
		glog.Errorf("decode config file %s error: %v", s.configPath, err)
		return err
	}

	glog.Infof("config loaded: %v", s.config)

	// 初始化mysql_client、redis_client
	mysql_client.InstallMysqlClientManager(s.config.Mysql)
	redis_client.InstallRedisClientManager(s.config.Redis)

	// 初始化redis_dao、mysql_dao
	dao.InstallMysqlDAOManager(mysql_client.GetMysqlClientManager())
	dao.InstallRedisDAOManager(redis_client.GetRedisClientManager())

	clients := map[string][]string{
		"session":[]string{"127.0.0.1:10000"},
	}

	s.client = net2.NewTcpClientGroupManager("zproto", clients, s)
	s.server = grpc_util.NewRpcServer(s.config.Server.Addr, &s.config.Discovery)

	return nil
}

func (s *syncServer) RunLoop() {
	go s.client.Serve()
	go s.server.Serve(func(s2 *grpc.Server) {
		// cache := cache2.NewAuthKeyCacheManager()
		// mtproto.RegisterRPCAuthKeyServer(s, rpc.NewAuthKeyService(cache))
		// mtproto.RegisterRPCSyncServer(s2, NewSyncService(s))
		s.impl = NewSyncService(s)
		mtproto.RegisterRPCSyncServer(s2, s.impl)
	})
}

func (s *syncServer) Destroy() {
	if s.impl != nil {
		s.impl.Destroy()
	}

	s.server.Stop()
	s.client.Stop()
}

////////////////////////////////////////////////////////////////////////////////////////////////////
func (s *syncServer) newMetadata() *mtproto.ZProtoMetadata {
	md := &mtproto.ZProtoMetadata{
		//ServerId: 1,
		//ClientConnId: conn.GetConnID(),
		//ClientAddr: conn.RemoteAddr().String(),
		From: "sync",
		ReceiveTime: time.Now().Unix(),
	}
	md.SpanId, _ = id.NextId()
	md.TraceId, _ = id.NextId()
	return md
}

func protoToRawPayload(m proto.Message) (*mtproto.ZProtoRawPayload, error) {
	x := mtproto.NewEncodeBuf(128)
	x.UInt(mtproto.SYNC_DATA)
	n := proto.MessageName(m)
	// glog.Infof("messageName: %s, d: {%v}", n, m)
	x.Int(int32(len(n)))
	x.Bytes([]byte(n))
	b, err := proto.Marshal(m)
	x.Bytes(b)
	return &mtproto.ZProtoRawPayload{Payload: x.GetBuf()}, err
}

// TcpClientCallBack
func (s *syncServer) OnNewClient(client *net2.TcpClient) {
	glog.Infof("OnNewConnection")

	// register
	zmsg := &mtproto.ZProtoMessage{
		SessionId: client.GetConnection().GetConnID(),
		SeqNum:    1, // TODO(@benqi): gen seqNum
		Metadata:  s.newMetadata(),
	}

	req := &mtproto.ConnectToSessionServerReq{}
	zmsg.Message, _ = protoToRawPayload(req)
	client.Send(zmsg)
}

func (s *syncServer) OnClientDataArrived(client *net2.TcpClient, msg interface{}) error {
	// glog.Infof("recv peer(%v) data: {%v}", client.GetRemoteAddress(), msg)
	// var err error
	zmsg, ok := msg.(*mtproto.ZProtoMessage)
	if !ok {
		return fmt.Errorf("invalid ZProtoMessage type: %v", msg)
	}

	payload, _ := zmsg.Message.(*mtproto.ZProtoRawPayload)
	msgType := binary.LittleEndian.Uint32(payload.Payload[:4])
	buf := payload.Payload[4:]
	switch msgType {
	case mtproto.SYNC_DATA:
		dbuf := mtproto.NewDecodeBuf(buf)
		len2 := int(dbuf.Int())
		messageName := string(dbuf.Bytes(len2))
		message, err := grpc_util.NewMessageByName(messageName)
		if err != nil {
			glog.Error(err)
			return err
		}

		err = proto.Unmarshal(buf[4+len2:], message)
		if err != nil {
			glog.Error(err)
			return err
		}

		switch message.(type) {
		case *mtproto.SessionServerConnectedRsp:
			glog.Infof("onSyncData - request(SessionServerConnectedRsp): {%v}", message)
			// TODO(@benqi): bind server_id, server_name
			res, _ := message.(*mtproto.SessionServerConnectedRsp)
			res.GetServerId()
		case *mtproto.VoidRsp:
			glog.Infof("onSyncData - request(PushUpdatesData): {%v}", message)
		default:
			glog.Errorf("invalid register proto type: {%v}", message)
		}
	default:
		return fmt.Errorf("invalid payload type: %v", msg)
	}
	return nil
}

func (s *syncServer) OnClientClosed(client *net2.TcpClient) {
	glog.Infof("OnConnectionClosed")

	if client.AutoReconnect() {
		client.Reconnect()
	}
}

func (s *syncServer) OnClientTimer(client *net2.TcpClient) {
	glog.Infof("OnTimer")
}

func (s *syncServer) sendToSessionServer(serverId int, m proto.Message) {
	zmsg := &mtproto.ZProtoMessage{
		// SessionId: client.GetConnection().GetConnID(),
		SeqNum:    1, // TODO(@benqi): gen seqNum
		Metadata:  s.newMetadata(),
	}
	zmsg.Message, _ = protoToRawPayload(m)
	s.client.SendData("session", zmsg)
}
