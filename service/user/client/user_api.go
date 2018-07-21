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

package user_client

import (
	"fmt"
	"github.com/nebulaim/telegramd/proto/mtproto"
)

type UserApi interface {
	Initialize(config string) error
	GetUserByID(selfId, id int32) (*mtproto.User, error)
	GetUserListByIDList(selfId int32, idList []int32) ([]*mtproto.User, error)
	GetUserByPhoneNumber(selfId int32, phone string) (*mtproto.User, error)
	GetSelfUserByPhoneNumber(phone string) (*mtproto.User, error)
}

type Instance func() UserApi

var adapters = make(map[string]Instance)

func Register(name string, adapter Instance) {
	if adapter == nil {
		panic("user_api: Register adapter is nil")
	}
	if _, ok := adapters[name]; ok {
		panic("user_api: Register called twice for adapter " + name)
	}
	adapters[name] = adapter
}

func NewUserApi(adapterName, config string) (adapter UserApi, err error) {
	instanceFunc, ok := adapters[adapterName]
	if !ok {
		err = fmt.Errorf("user_api: unknown adapter name %q (forgot to import?)", adapterName)
		return
	}
	adapter = instanceFunc()
	err = adapter.Initialize(config)
	if err != nil {
		adapter = nil
	}
	return
}
