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

package main

import (
	"flag"
	"github.com/golang/glog"

	phone "github.com/nebulaim/telegramd/biz_server/phone/rpc"
	"github.com/nebulaim/telegramd/mtproto"
	"github.com/nebulaim/telegramd/baselib/redis_client"
	"github.com/nebulaim/telegramd/baselib/mysql_client"
	"github.com/BurntSushi/toml"
	"fmt"
	"github.com/nebulaim/telegramd/biz/dal/dao"
	"github.com/nebulaim/telegramd/baselib/grpc_util"
	"github.com/nebulaim/telegramd/baselib/grpc_util/service_discovery"
	"google.golang.org/grpc"
)

func init() {
	flag.Set("alsologtostderr", "true")
	flag.Set("log_dir", "false")
}

type RpcServerConfig struct {
	Addr string
}

//type RpcClientConfig struct {
//	ServiceName string
//	Addr string
//}

type phoneServerConfig struct{
	Server 		*RpcServerConfig
	Discovery service_discovery.ServiceDiscoveryServerConfig

	// RpcClient	*RpcClientConfig
	Mysql		[]mysql_client.MySQLConfig
	Redis 		[]redis_client.RedisConfig
}

// 整合各服务，方便开发调试
func main() {
	flag.Parse()

	config := &phoneServerConfig{}
	if _, err := toml.DecodeFile("./auth.toml", config); err != nil {
		fmt.Errorf("%s\n", err)
		return
	}

	glog.Info(config)

	// 初始化mysql_client、redis_client
	redis_client.InstallRedisClientManager(config.Redis)
	mysql_client.InstallMysqlClientManager(config.Mysql)

	// 初始化redis_dao、mysql_dao
	dao.InstallMysqlDAOManager(mysql_client.GetMysqlClientManager())
	dao.InstallRedisDAOManager(redis_client.GetRedisClientManager())

	// Start server
	grpcServer := grpc_util.NewRpcServer(config.Server.Addr, &config.Discovery)
	grpcServer.Serve(func(s *grpc.Server) {
		mtproto.RegisterRPCPhoneServer(s, &phone.PhoneServiceImpl{})
	})
}
