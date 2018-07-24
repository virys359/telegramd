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

package updates

import "github.com/nebulaim/telegramd/biz/core"

/**
  1. peer_user和peer_chat使用user_dialogs存储
  2. channel和super_chat用channel_participants
*/

type updatesDAO struct {
	//*mysql_dao.UserContactsDAO
	//*mysql_dao.UsersDAO
	//*mysql_dao.UnregisteredContactsDAO
	//*mysql_dao.PopularContactsDAO

}

type UpdateModel struct {
	dao *updatesDAO
}

func (m *UpdateModel) InstallModel() {
	//m.dao.UserContactsDAO = dao.GetUserContactsDAO(dao.DB_MASTER)
	//m.dao.UsersDAO = dao.GetUsersDAO(dao.DB_MASTER)
	//m.dao.UnregisteredContactsDAO = dao.GetUnregisteredContactsDAO(dao.DB_MASTER)
	//m.dao.PopularContactsDAO = dao.GetPopularContactsDAO(dao.DB_MASTER)
}

func (m *UpdateModel) RegisterCallback(cb interface{}) {
}

func init() {
	core.RegisterCoreModel(&UpdateModel{dao: &updatesDAO{}})
}
