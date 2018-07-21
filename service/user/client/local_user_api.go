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
	"github.com/golang/glog"
	"github.com/nebulaim/telegramd/proto/mtproto"
	"github.com/nebulaim/telegramd/biz/dal/dao/mysql_dao"
	"github.com/nebulaim/telegramd/baselib/mysql_client"
)

type localUserApi struct {
	// TODO(@benqi): add user cache
	*mysql_dao.UsersDAO
}

func localUserApiInstance() UserApi {
	return &localUserApi{}
}

func NewLocalUserApi(dbName string) (*localUserApi, error) {
	var err error

	dbClient := mysql_client.GetMysqlClient(dbName)
	if dbClient == nil {
		err = fmt.Errorf("invalid dbName: %s", dbName)
		glog.Error(err)
		return nil, err
	}

	return &localUserApi{UsersDAO: mysql_dao.NewUsersDAO(dbClient)}, nil
}

func (c *localUserApi) Initialize(config string) error {
	glog.Info("localUserApi - Initialize config: ", config)

	var err error

	dbName := config
	dbClient := mysql_client.GetMysqlClient(dbName)
	if dbClient == nil {
		err = fmt.Errorf("invalid dbName: %s", dbName)
		glog.Error(err)
	}
	c.UsersDAO = mysql_dao.NewUsersDAO(dbClient)

	return err
}

func (c *localUserApi) GetUserByID(selfId, id int32) (*mtproto.User, error) {
	return nil, nil
}

func (c *localUserApi) GetUserListByIDList(selfId int32, idList []int32) ([]*mtproto.User, error) {
	return nil, nil
}

func (c *localUserApi) GetUserByPhoneNumber(selfId int32, phone string) (*mtproto.User, error) {
	return nil, nil
}

func (c *localUserApi) GetSelfUserByPhoneNumber(phone string) (*mtproto.User, error) {
	return nil, nil
}

func init() {
	Register("local", localUserApiInstance)
}
