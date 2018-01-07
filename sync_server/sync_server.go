/*
 *  Copyright (c) 2017, https://github.com/nebulaim
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
	"google.golang.org/grpc"
	"github.com/nebulaim/telegramd/zproto"
	"github.com/nebulaim/telegramd/base/redis_client"
	"github.com/BurntSushi/toml"
	"fmt"
	"github.com/nebulaim/telegramd/biz_model/dal/dao"
	"github.com/nebulaim/telegramd/sync_server/rpc"
	"github.com/nebulaim/telegramd/base/mysql_client"
	"github.com/nebulaim/telegramd/grpc_util"
	"github.com/nebulaim/telegramd/grpc_util/service_discovery"
)

func init() {
	flag.Set("alsologtostderr", "true")
	flag.Set("log_dir", "false")

}

type RpcServerConfig struct {
	Addr string
}

type SyncServerConfig struct{
	Server 		*RpcServerConfig
	Discovery service_discovery.ServiceDiscoveryServerConfig
	Redis 		[]redis_client.RedisConfig
	Mysql     []mysql_client.MySQLConfig
}

// 整合各服务，方便开发调试
func main() {
	flag.Parse()

	syncServerConfig := &SyncServerConfig{}
	if _, err := toml.DecodeFile("./sync_server.toml", syncServerConfig); err != nil {
		fmt.Errorf("%s\n", err)
		return
	}

	glog.Info(syncServerConfig)

	// 初始化mysql_client、redis_client
	redis_client.InstallRedisClientManager(syncServerConfig.Redis)
	mysql_client.InstallMysqlClientManager(syncServerConfig.Mysql)

	// 初始化redis_dao、mysql_dao
	dao.InstallMysqlDAOManager(mysql_client.GetMysqlClientManager())
	dao.InstallRedisDAOManager(redis_client.GetRedisClientManager())

	authKeyServer := grpc_util.NewRpcServer(syncServerConfig.Server.Addr, &syncServerConfig.Discovery)
	authKeyServer.Serve(func(s *grpc.Server) {
		// cache := cache2.NewAuthKeyCacheManager()
		// mtproto.RegisterRPCAuthKeyServer(s, rpc.NewAuthKeyService(cache))
		zproto.RegisterRPCSyncServer(s, rpc.NewSyncService())
	})
}
