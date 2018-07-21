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

package core

import (
	// "github.com/jmoiron/sqlx"
	// "sync"
	// "github.com/nebulaim/telegramd/baselib/redis_client"
)

const (
	TOKEN_TYPE_APNS = 1
	TOKEN_TYPE_GCM = 2
	TOKEN_TYPE_MPNS = 3
	TOKEN_TYPE_SIMPLE_PUSH = 4
	TOKEN_TYPE_UBUNTU_PHONE = 5
	TOKEN_TYPE_BLACKBERRY = 6
	// Android里使用
	TOKEN_TYPE_INTERNAL_PUSH = 7
	// web
	TOKEN_TYPE_WEB_PUSH = 10
	TOKEN_TYPE_MAXSIZE = 10
)

//type CoreModel interface {
//	InstallMysqlClients(dbClients sync.Map)
//	InstallRedisClients(map[string]*redis_client.RedisPool)
//}
//
//// type Instance func() Initializer
//
//var models = []CoreModel{}
//
//func RegisterCoreModel(model CoreModel) {
//	models = append(models, model)
//}
//
//func InstallMysqlClients(clients sync.Map) {
//	for _, m := range models {
//		m.InstallMysqlClients(clients)
//	}
//}
//
//func InstallRedisClients(clients map[string]*redis_client.RedisPool) {
//	for _, m := range models {
//		m.InstallRedisClients(clients)
//	}
//}
