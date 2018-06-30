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
	"encoding/binary"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/coreos/etcd/clientv3"
	"github.com/golang/glog"
	"github.com/nebulaim/telegramd/baselib/grpc_util"
	"github.com/nebulaim/telegramd/baselib/grpc_util/service_discovery"
	"github.com/nebulaim/telegramd/baselib/grpc_util/service_discovery/etcd3"
	"github.com/nebulaim/telegramd/baselib/net2"
	"github.com/nebulaim/telegramd/baselib/redis_client"
	"github.com/nebulaim/telegramd/biz/dal/dao"
	"github.com/nebulaim/telegramd/mtproto"
	"net"
	"time"
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
	server := net2.NewTcpServer(net2.TcpServerArgs{
		Listener:           lsn,
		ServerName:         config.Name,
		ProtoName:          config.ProtoName,
		SendChanSize:       1024,
		ConnectionCallback: cb,
	}) // todo (yumcoder): set max connection
	return server, nil
}

type SessionConfig struct {
	ServerId         int32 // 服务器ID
	Redis            []redis_client.RedisConfig
	SaltCache        redis_client.RedisConfig
	AuthKeyRpcClient service_discovery.ServiceDiscoveryClientConfig
	BizRpcClient     service_discovery.ServiceDiscoveryClientConfig
	NbfsRpcClient    service_discovery.ServiceDiscoveryClientConfig
	SyncRpcClient    service_discovery.ServiceDiscoveryClientConfig
	Server           ServerConfig
	Discovery        service_discovery.ServiceDiscoveryServerConfig
}

type SessionServer struct {
	configPath     string
	config         *SessionConfig
	server         *net2.TcpServer
	client         *net2.TcpClientGroupManager
	bizRpcClient   *grpc_util.RPCClient
	nbfsRpcClient  *grpc_util.RPCClient
	syncRpcClient  mtproto.RPCSyncClient
	sessionManager *sessionManager
	syncHandler    *syncHandler
	registry       *etcd3.EtcdReigistry
}

func NewSessionServer(configPath string) *SessionServer {
	return &SessionServer{
		configPath: configPath,
		config:     &SessionConfig{},
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
	redis_client.InstallRedisClientManager(s.config.Redis)

	// 初始化redis_dao、mysql_dao
	dao.InstallRedisDAOManager(redis_client.GetRedisClientManager())

	// TODO(@benqi): config cap
	InitCacheAuthManager(1024*1024, &s.config.AuthKeyRpcClient)

	s.sessionManager = newSessionManager()
	s.syncHandler = newSyncHandler(s.sessionManager)
	s.server, err = newTcpServer(s.config.Server, s)
	if err != nil {
		glog.Error(err)
		return err
	}

	etcdConfg := clientv3.Config{
		Endpoints: s.config.Discovery.EtcdAddrs,
	}
	s.registry, err = etcd3.NewRegistry(
		etcd3.Option{
			EtcdConfig:  etcdConfg,
			RegistryDir: "/nebulaim",
			ServiceName: s.config.Discovery.ServiceName,
			NodeID:      s.config.Discovery.NodeID,
			NData: 		 etcd3.NodeData{
				Addr:     s.config.Discovery.RPCAddr,
				Metadata: map[string]string{},
				// Metadata: map[string]string{"weight": "1"},
			},
			Ttl: time.Duration(s.config.Discovery.TTL), // * time.Second,
		})
	if err != nil {
		glog.Fatal(err)
		// return nil
	}

	return nil
}

func (s *SessionServer) RunLoop() {
	// TODO(@benqi): check error
	// timingWheel.Start()

	s.bizRpcClient, _ = grpc_util.NewRPCClient(&s.config.BizRpcClient)
	s.nbfsRpcClient, _ = grpc_util.NewRPCClient(&s.config.NbfsRpcClient)
	c, _ := grpc_util.NewRPCClient(&s.config.SyncRpcClient)
	s.syncRpcClient = mtproto.NewRPCSyncClient(c.GetClientConn())
	// client: mtproto.NewZRPCAuthKeyClient(conn),

	go s.registry.Register()
	go s.server.Serve()
	// go s.client.Serve()
}

func (s *SessionServer) Destroy() {
	glog.Infof("sessionServer - destroy...")
	// timingWheel.Stop()
	s.registry.Deregister()
	s.server.Stop()
	time.Sleep(1 * time.Second)
	// s.client.Stop()
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// TcpConnectionCallback
func (s *SessionServer) OnNewConnection(conn *net2.TcpConnection) {
	glog.Infof("OnNewConnection %v", conn.RemoteAddr())
}

func (s *SessionServer) OnConnectionDataArrived(conn *net2.TcpConnection, msg interface{}) error {
	// var err error
	zmsg, ok := msg.(*mtproto.ZProtoMessage)
	if !ok {
		return fmt.Errorf("invalid ZProtoMessage type: %v", msg)
	}
	glog.Infof("recv peer(%v) data: {%v}", conn.RemoteAddr(), zmsg)

	// var err error
	// res *mtproto.ZProtoRawPayload
	switch zmsg.Message.(type) {
	case *mtproto.ZProtoRawPayload:
		payload, _ := zmsg.Message.(*mtproto.ZProtoRawPayload)
		msgType := binary.LittleEndian.Uint32(payload.Payload[:4])
		switch msgType {
		case mtproto.SESSION_SESSION_CLIENT_NEW:
			return nil
		case mtproto.SESSION_SESSION_DATA:
			return s.sessionManager.onSessionData2(conn.GetConnID(), zmsg.SessionId, zmsg.Metadata, payload.Payload[4:])
			// return s.sessionManager.onSessionData(conn, zmsg.SessionId, zmsg.Metadata, payload.Payload[4:])
		case mtproto.SESSION_SESSION_CLIENT_CLOSED:
			return nil
		case mtproto.SYNC_DATA:
			sres, err := s.syncHandler.onSyncData(conn, payload.Payload[4:])
			if err != nil {
				glog.Error(err)
				return nil
			}
			res := &mtproto.ZProtoMessage{
				SessionId: zmsg.SessionId,
				SeqNum:    zmsg.SeqNum,
				Metadata:  zmsg.Metadata,
				Message:   sres,
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
	glog.Infof("sendToClientData - {%d, %d}", connID, sessionID)
	conn := s.server.GetConnection(connID)
	if conn != nil {
		return sendDataByConnection(conn, sessionID, md, buf)
	} else {
		err := fmt.Errorf("send data error, conn offline, connID: %d", connID)
		glog.Error(err)
		return err
	}
}
