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

package model

import (
	"sync"
	"github.com/nebulaim/telegramd/biz_model/dal/dao"
)

type contactModel struct {
	// chatDAO *dao.UserDialogsDAO
}

var (
	contactInstance *contactModel
	contactInstanceOnce sync.Once
)

func GetContactModel() *contactModel {
	contactInstanceOnce.Do(func() {
		contactInstance = &contactModel{}
	})
	return contactInstance
}

func (m *contactModel) GetContactUserIDList(userId int32) []int32 {
	contactsDOList := dao.GetUserContactsDAO(dao.DB_SLAVE).SelectUserContacts(userId)
	idList := make([]int32, 0, len(contactsDOList))

	for _, do := range contactsDOList {
		idList = append(idList, do.ContactUserId)
	}
	return idList
}
