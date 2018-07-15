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

	account "github.com/nebulaim/telegramd/server/biz_server/account/rpc"
	auth "github.com/nebulaim/telegramd/server/biz_server/auth/rpc"
	bots "github.com/nebulaim/telegramd/server/biz_server/bots/rpc"
	channels "github.com/nebulaim/telegramd/server/biz_server/channels/rpc"
	contacts "github.com/nebulaim/telegramd/server/biz_server/contacts/rpc"
	help "github.com/nebulaim/telegramd/server/biz_server/help/rpc"
	langpack "github.com/nebulaim/telegramd/server/biz_server/langpack/rpc"
	messages "github.com/nebulaim/telegramd/server/biz_server/messages/rpc"
	payments "github.com/nebulaim/telegramd/server/biz_server/payments/rpc"
	phone "github.com/nebulaim/telegramd/server/biz_server/phone/rpc"
	photos "github.com/nebulaim/telegramd/server/biz_server/photos/rpc"
	stickers "github.com/nebulaim/telegramd/server/biz_server/stickers/rpc"
	updates "github.com/nebulaim/telegramd/server/biz_server/updates/rpc"
	users "github.com/nebulaim/telegramd/server/biz_server/users/rpc"
	"github.com/nebulaim/telegramd/proto/mtproto"
	"github.com/nebulaim/telegramd/baselib/redis_client"
	"github.com/nebulaim/telegramd/baselib/mysql_client"
	"github.com/BurntSushi/toml"
	"fmt"
	"github.com/nebulaim/telegramd/biz/dal/dao"
	"github.com/nebulaim/telegramd/baselib/grpc_util"
	"github.com/nebulaim/telegramd/baselib/grpc_util/service_discovery"
	"google.golang.org/grpc"
	"github.com/nebulaim/telegramd/server/sync/sync_client"
	"github.com/nebulaim/telegramd/server/nbfs/nbfs_client"
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

type BizServerConfig struct{
	Server 		*RpcServerConfig
	Discovery service_discovery.ServiceDiscoveryServerConfig

	// RpcClient	*RpcClientConfig
	Mysql		[]mysql_client.MySQLConfig
	Redis 		[]redis_client.RedisConfig
	NbfsRpcClient  *service_discovery.ServiceDiscoveryClientConfig
	SyncRpcClient1 *service_discovery.ServiceDiscoveryClientConfig
	SyncRpcClient2 *service_discovery.ServiceDiscoveryClientConfig
}

// 整合各服务，方便开发调试
func main() {
	flag.Parse()

	bizServerConfig := &BizServerConfig{}
	if _, err := toml.DecodeFile("./biz_server.toml", bizServerConfig); err != nil {
		fmt.Errorf("%s\n", err)
		return
	}

	glog.Info(bizServerConfig)

	// 初始化mysql_client、redis_client
	redis_client.InstallRedisClientManager(bizServerConfig.Redis)
	mysql_client.InstallMysqlClientManager(bizServerConfig.Mysql)

	// 初始化redis_dao、mysql_dao
	dao.InstallMysqlDAOManager(mysql_client.GetMysqlClientManager())
	dao.InstallRedisDAOManager(redis_client.GetRedisClientManager())

	nbfs_client.InstallNbfsClient(bizServerConfig.NbfsRpcClient)
	sync_client.InstallSyncClient(bizServerConfig.SyncRpcClient2)

	// InstallNbfsClient

	// Start server
	grpcServer := grpc_util.NewRpcServer(bizServerConfig.Server.Addr, &bizServerConfig.Discovery)
	grpcServer.Serve(func(s *grpc.Server) {
		// AccountServiceImpl
		mtproto.RegisterRPCAccountServer(s, &account.AccountServiceImpl{})

		// AuthServiceImpl
		mtproto.RegisterRPCAuthServer(s, &auth.AuthServiceImpl{})

		mtproto.RegisterRPCBotsServer(s, &bots.BotsServiceImpl{})
		mtproto.RegisterRPCChannelsServer(s, &channels.ChannelsServiceImpl{})

		// ContactsServiceImpl
		mtproto.RegisterRPCContactsServer(s, &contacts.ContactsServiceImpl{})

		mtproto.RegisterRPCHelpServer(s, &help.HelpServiceImpl{})
		mtproto.RegisterRPCLangpackServer(s, &langpack.LangpackServiceImpl{})

		// MessagesServiceImpl
		mtproto.RegisterRPCMessagesServer(s, &messages.MessagesServiceImpl{})

		mtproto.RegisterRPCPaymentsServer(s, &payments.PaymentsServiceImpl{})
		mtproto.RegisterRPCPhoneServer(s, &phone.PhoneServiceImpl{})
		mtproto.RegisterRPCPhotosServer(s, &photos.PhotosServiceImpl{})
		mtproto.RegisterRPCStickersServer(s, &stickers.StickersServiceImpl{})
		mtproto.RegisterRPCUpdatesServer(s, &updates.UpdatesServiceImpl{})

		mtproto.RegisterRPCUsersServer(s, &users.UsersServiceImpl{})
	})
}
