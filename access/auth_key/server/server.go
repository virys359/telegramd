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
	"github.com/nebulaim/telegramd/access/auth_key/dal/dao"
	"github.com/nebulaim/telegramd/baselib/grpc_util/service_discovery"
	"github.com/nebulaim/telegramd/baselib/grpc_util/service_discovery/etcd3"
	"github.com/coreos/etcd/clientv3"
	"time"
	"github.com/nebulaim/telegramd/baselib/grpc_util"
	"google.golang.org/grpc"
)

type ServerConfig struct {
	Name      string
	ProtoName string
	Addr      string
}

type rpcServerConfig struct {
	Addr string
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
	}) // todo (omid): set max connection
	return server, nil
}

type AuthKeyConfig struct {
	ServerId      int32 // 服务器ID
	Mysql         []mysql_client.MySQLConfig
	Server        ServerConfig
	Discovery     service_discovery.ServiceDiscoveryServerConfig

	RpcServer     *rpcServerConfig
	RpcDiscovery  service_discovery.ServiceDiscoveryServerConfig
}

type AuthKeyServer struct {
	configPath     string
	config         *AuthKeyConfig
	server         *net2.TcpServer
	handshake      *handshake
	registry       *etcd3.EtcdReigistry
	rpcServer      *grpc_util.RPCServer

}

func NewAuthKeyServer(configPath string) *AuthKeyServer {
	return &AuthKeyServer{
		configPath:     configPath,
		config:         &AuthKeyConfig{},
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// AppInstance interface
func (s *AuthKeyServer) Initialize() error {
	var err error

	if _, err = toml.DecodeFile(s.configPath, s.config); err != nil {
		glog.Errorf("decode config file %s error: %v", s.configPath, err)
		return err
	}

	glog.Infof("config loaded: %v", s.config)

	// 初始化mysql_client、redis_client
	mysql_client.InstallMysqlClientManager(s.config.Mysql)

	// 初始化redis_dao、mysql_dao
	dao.InstallMysqlDAOManager(mysql_client.GetMysqlClientManager())

	s.handshake = newHandshake()
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
			NData: etcd3.NodeData{
				Addr: s.config.Discovery.RPCAddr,
				Metadata: map[string]string{},
				// Metadata: map[string]string{"weight": "1"},
			},
			Ttl: time.Duration(s.config.Discovery.TTL), // * time.Second,
		})
	if err != nil {
		glog.Fatal(err)
	}

	s.rpcServer = grpc_util.NewRpcServer(s.config.RpcServer.Addr, &s.config.RpcDiscovery)
	return nil
}

func (s *AuthKeyServer) RunLoop() {
	// TODO(@benqi): check error
	go s.registry.Register()
	go s.server.Serve()
	go s.rpcServer.Serve(func(s2 *grpc.Server) {
		mtproto.RegisterZRPCAuthKeyServer(s2, new(AuthKeyServiceImpl))
	})
}

func (s *AuthKeyServer) Destroy() {
	glog.Infof("sessionServer - destroy...")
	s.registry.Deregister()
	s.server.Stop()
	s.rpcServer.Stop()
	time.Sleep(1*time.Second)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// TcpConnectionCallback
func (s *AuthKeyServer) OnNewConnection(conn *net2.TcpConnection) {
	glog.Infof("onNewConnection %v", conn.RemoteAddr())
}

func (s *AuthKeyServer) OnConnectionDataArrived(conn *net2.TcpConnection, msg interface{}) error {
	// var err error
	zmsg, ok := msg.(*mtproto.ZProtoMessage)
	if !ok {
		glog.Error("invalid ZProtoMessage type")
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
		case mtproto.SESSION_HANDSHAKE:
			hrsp, err := s.handshake.onHandshake(conn, payload)
			if err != nil {
				glog.Error(err)
				return nil
			}
			if hrsp != nil {
				res := &mtproto.ZProtoMessage{
					SessionId: zmsg.SessionId,
					SeqNum:    1,
					Metadata:  zmsg.Metadata,
					Message:   hrsp,
				}
				return conn.Send(res)
			} else {
				return nil
			}
		default:
			return fmt.Errorf("invalid payload type: %v", msg)
		}
	default:
		return fmt.Errorf("invalid zmsg type: {%v}", zmsg)
	}
}

func (s *AuthKeyServer) OnConnectionClosed(conn *net2.TcpConnection) {
	glog.Infof("onConnectionClosed - %v", conn.RemoteAddr())
}
