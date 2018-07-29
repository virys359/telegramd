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

package dialog

import (
	"github.com/nebulaim/telegramd/biz/core"
	"github.com/nebulaim/telegramd/biz/dal/dao/mysql_dao"
	"github.com/nebulaim/telegramd/biz/dal/dao"
)

type dialogsDAO struct {
	*mysql_dao.UserDialogsDAO
}

type DialogModel struct {
	dao *dialogsDAO
}

func (m *DialogModel) RegisterCallback(cb interface{}) {
}

func (m *DialogModel) InstallModel() {
	m.dao.UserDialogsDAO = dao.GetUserDialogsDAO(dao.DB_MASTER)
}

func (m *DialogModel) MakeDialogLogic(userId, peerType, peerId int32) *dialogLogic {
	// m.dao.UserDialogsDAO = dao.GetUserDialogsDAO(dao.DB_MASTER)
	return &dialogLogic{
		selfUserId: userId,
		peerType:   peerType,
		peerId:     peerId,
		dao:        m.dao,
	}
}

func (m *DialogModel) UpdateUnreadByPeer(userId int32, peerType int8, peerId int32, readInboxMaxId int32) {
	m.dao.UserDialogsDAO.UpdateUnreadByPeer(readInboxMaxId, userId, peerType, peerId)
}

func (m *DialogModel) UpdateReadOutboxMaxIdByPeer(userId int32, peerType int8, peerId int32, topMessage int32)  {
	m.dao.UserDialogsDAO.UpdateReadOutboxMaxIdByPeer(topMessage, userId, peerType, peerId)
}

func (m *DialogModel) GetTopMessage(userId int32, peerType int8, peerId int32) int32 {
	do := m.dao.UserDialogsDAO.SelectByPeer(userId, peerType, peerId)
	if do != nil {
		return do.TopMessage
	}
	return 0
}

func init() {
	core.RegisterCoreModel(&DialogModel{dao: &dialogsDAO{}})
}
