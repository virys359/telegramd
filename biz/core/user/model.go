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

package user

import (
	"github.com/golang/glog"
	"github.com/nebulaim/telegramd/biz/core"
	"github.com/nebulaim/telegramd/biz/dal/dao"
	"github.com/nebulaim/telegramd/biz/dal/dao/mysql_dao"
)

type usersDAO struct {
	*mysql_dao.UsersDAO
	*mysql_dao.UserPresencesDAO
	*mysql_dao.UserContactsDAO
	*mysql_dao.UserDialogsDAO
	// redisClient *redis_client.RedisPool
}

type UserModel struct {
	dao             *usersDAO
	contactCallback core.ContactCallback
	photoCallback   core.PhotoCallback
}

func (m *UserModel) InstallModel() {
	m.dao.UsersDAO = dao.GetUsersDAO(dao.DB_MASTER)
	m.dao.UserPresencesDAO = dao.GetUserPresencesDAO(dao.DB_MASTER)
	m.dao.UserContactsDAO = dao.GetUserContactsDAO(dao.DB_MASTER)
	m.dao.UserDialogsDAO = dao.GetUserDialogsDAO(dao.DB_MASTER)
	// m.dao.redisClient = redis_client.GetRedisClient(dao.CACHE)
}

func (m *UserModel) RegisterCallback(cb interface{}) {
	switch cb.(type) {
	case core.ContactCallback:
		glog.Info("userModel - register core.ContactCallback")
		m.contactCallback = cb.(core.ContactCallback)
	case core.PhotoCallback:
		glog.Info("userModel - register core.PhotoCallback")
		m.photoCallback = cb.(core.PhotoCallback)
	}
}

func init() {
	core.RegisterCoreModel(&UserModel{dao: &usersDAO{}})
}
