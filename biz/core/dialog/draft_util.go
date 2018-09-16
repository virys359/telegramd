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
	"encoding/json"
	"github.com/nebulaim/telegramd/biz/base"
	"github.com/nebulaim/telegramd/biz/dal/dataobject"
	"github.com/nebulaim/telegramd/proto/mtproto"
)

func GetDraftID(peerType, peerId int32) (draftId int32) {
	switch peerType {
	case base.PEER_USER:
		draftId = peerId
	case base.PEER_CHAT, base.PEER_CHANNEL:
		draftId = -peerId
	}
	return
}

func (m *DialogModel) ClearDraft(userId, peerType, peerId int32) bool {
	affected := m.dao.DraftMessagesDAO.ClearDraft(userId, GetDraftID(peerType, peerId))
	return affected > 0
}

func (m *DialogModel) SaveDraftMessage(userId int32, peerType int32, peerId int32, message *mtproto.DraftMessage) {
	draftData, _ := json.Marshal(message)

	draftDO := &dataobject.DraftMessagesDO{
		UserId:           userId,
		DraftId:          GetDraftID(peerType, peerId),
		DraftType:        2,
		DraftMessageData: string(draftData),
	}

	m.dao.DraftMessagesDAO.InsertOrUpdate(draftDO)
}
