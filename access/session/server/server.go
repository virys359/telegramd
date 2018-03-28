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
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/nebulaim/telegramd/mtproto"
	"encoding/binary"
	"github.com/nebulaim/telegramd/baselib/mysql_client"
	"github.com/nebulaim/telegramd/biz/dal/dao"
	"github.com/nebulaim/telegramd/baselib/redis_client"
	"github.com/nebulaim/telegramd/grpc_util/service_discovery"
	"github.com/nebulaim/telegramd/grpc_util"
)

type ServerConfig struct {
	Name      string
	ProtoName string
	Addr      string
}

func newTcpServer(config ServerConfig, cb net2.TcpConnectionCallback) (*net2.TcpServer, error) {
	lsn, err := net.Listen("tcp", config.Addr)
	if err != nil {
		// glog.Errorf("listen error: %v", err)
		return nil, err
	}
	server := net2.NewTcpServer(lsn, config.Name, config.ProtoName, 1024, cb)
	return server, nil
}

type SessionConfig struct {
	ServerId     int32 // 服务器ID
	Mysql        []mysql_client.MySQLConfig
	Redis        []redis_client.RedisConfig
	BizRpcClient service_discovery.ServiceDiscoveryClientConfig
	Server       ServerConfig
}

type SessionServer struct {
	configPath     string
	config         *SessionConfig
	server         *net2.TcpServer
	client         *net2.TcpClientGroupManager
	rpcClient      *grpc_util.RPCClient
	handshake      *handshake
	sessionManager *sessionManager
	syncHandler    *syncHandler
}

func NewSessionServer(configPath string) *SessionServer {
	return &SessionServer{
		configPath:     configPath,
		config:         &SessionConfig{},
		// handshake:      &handshake{},
		// sessionManager: newSessionManager(),
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// AppInstance interface
func (s *SessionServer) Initialize() error {
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

	cache := NewAuthKeyCacheManager()
	s.handshake = newHandshake(cache)
	s.sessionManager = newSessionManager(cache)
	s.syncHandler = newSyncHandler(s.sessionManager)
	s.server, err = newTcpServer(s.config.Server, s)
	if err != nil {
		glog.Error(err)
		return err
	}
	return nil
}

func (s *SessionServer) RunLoop() {
	// TODO(@benqi): check error
	s.rpcClient, _ = grpc_util.NewRPCClient(&s.config.BizRpcClient)
	s.server.Serve()
	// go s.client.Serve()
}

func (s *SessionServer) Destroy() {
	s.server.Stop()
	// s.client.Stop()
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// TcpConnectionCallback
func (s *SessionServer) OnNewConnection(conn *net2.TcpConnection) {
	glog.Infof("OnNewConnection %v", conn.RemoteAddr())
}

func (s *SessionServer) OnConnectionDataArrived(conn *net2.TcpConnection, msg interface{}) error {
	glog.Infof("recv peer(%v) data: {%v}", conn.RemoteAddr(), msg)
	// var err error
	zmsg, ok := msg.(*mtproto.ZProtoMessage)
	if !ok {
		return fmt.Errorf("invalid ZProtoMessage type: %v", msg)
	}

	// var err error
	// res *mtproto.ZProtoRawPayload
	switch zmsg.Message.(type) {
	case *mtproto.ZProtoRawPayload:
		payload, _ := zmsg.Message.(*mtproto.ZProtoRawPayload)
		msgType := binary.LittleEndian.Uint32(payload.Payload[:4])
		switch msgType {
		case mtproto.SESSION_HANDSHAKE:
			hrsp, err := s.handshake.onHandshake(conn, payload)
			if err != nil {
				glog.Error(err)
				return nil
			}
			if hrsp != nil {
				res := &mtproto.ZProtoMessage{
					SessionId: zmsg.SessionId,
					SeqNum: 1,
					Metadata: zmsg.Metadata,
					Message:  hrsp,
				}
				return conn.Send(res)
			} else {
				return nil
			}
		case mtproto.SESSION_SESSION_DATA:
			return s.sessionManager.onSessionData(conn, zmsg.SessionId, zmsg.Metadata, payload.Payload[4:])
		case mtproto.SYNC_DATA:
			sres, err := s.syncHandler.onSyncData(conn, payload.Payload[4:])
			if err != nil {
				glog.Error(err)
				return nil
			}
			res := &mtproto.ZProtoMessage{
				SessionId: zmsg.SessionId,
				SeqNum: 1,
				Metadata: zmsg.Metadata,
				Message:  sres,
			}
			return conn.Send(res)
		default:
			return fmt.Errorf("invalid payload type: %v", msg)
		}
	default:
		return fmt.Errorf("invalid zmsg type: {%v}", zmsg)
	}
}

func (s *SessionServer) OnConnectionClosed(conn *net2.TcpConnection) {
	glog.Infof("OnConnectionClosed - %v", conn.RemoteAddr())
}

func (s *SessionServer) SendToClientData(connID, sessionID uint64, md *mtproto.ZProtoMetadata, buf []byte) error {
	conn := s.server.GetConnection(connID)
	if conn != nil {
		return sendDataByConnection(conn, sessionID, md, buf)
	} else {
		err := fmt.Errorf("send data error, conn offline, connID: %d", connID)
		glog.Error(err)
		return err
	}
}
