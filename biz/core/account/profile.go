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

package account

import (
	"github.com/nebulaim/telegramd/biz/dal/dao"
)

// not found, return 0
func GetUserIdByUserName(name string) int32 {
	do := dao.GetUsersDAO(dao.DB_SLAVE).SelectByUsername(name)
	if do == nil {
		return 0
	}
	return do.Id
}

func ChangeUserNameByUserId(id int32, name string) int64 {
	return dao.GetUsersDAO(dao.DB_MASTER).UpdateUsername(name, id)
}

func UpdateFirstAndLastName(id int32, firstName, lastName string) int64 {
	return dao.GetUsersDAO(dao.DB_MASTER).UpdateFirstAndLastName(firstName, lastName, id)
}

func UpdateAbout(id int32, about string) int64 {
	return dao.GetUsersDAO(dao.DB_MASTER).UpdateAbout(about, id)
}
