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
	"time"
	"github.com/nebulaim/telegramd/mtproto"
	"github.com/nebulaim/telegramd/biz_model/dal/dao"
	"github.com/nebulaim/telegramd/biz_model/dal/dataobject"
	"github.com/nebulaim/telegramd/baselib/base"
	"encoding/json"
	"github.com/golang/glog"
)

const (
	PTS_UPDATE_TYPE_UNKNOWN = 0

	// pts
	PTS_UPDATE_NEW_MESSAGE = 1
	PTS_UPDATE_DELETE_MESSAGES = 2
	PTS_UPDATE_READ_HISTORY_OUTBOX = 3
	PTS_UPDATE_READ_HISTORY_INBOX = 4
	PTS_UPDATE_WEBPAGE = 5
	PTS_UPDATE_READ_MESSAGE_CONENTS = 6
	PTS_UPDATE_EDIT_MESSAGE = 7

	// channel pts
	PTS_UPDATE_NEW_CHANNEL_MESSAGE = 9
	PTS_UPDATE_DELETE_CHANNEL_MESSAGE = 9
	PTS_UPDATE_EDIT_CHANNEL_MESSAGE = 10
	PTS_UPDATE_EDIT_CHANNEL_WEBPAGE = 11
)

type updatesModel struct {
}

var (
	updatesInstance *updatesModel
	updatesInstanceOnce sync.Once
)

func GetUpdatesModel() *updatesModel {
	updatesInstanceOnce.Do(func() {
		updatesInstance = &updatesModel{}
	})
	return updatesInstance
}

func (m *updatesModel) GetUpdatesState(authKeyId int64, userId int32) *mtproto.TLUpdatesState {
	state := mtproto.NewTLUpdatesState()

	// TODO(@benqi): first sign in, state data???
	ptsDO := dao.GetUserPtsUpdatesDAO(dao.DB_SLAVE).SelectLastPts(userId)
	if ptsDO != nil {
		state.SetPts(ptsDO.Pts)
	} else {
		state.SetPts(1)
	}

	qtsDO := dao.GetUserQtsUpdatesDAO(dao.DB_SLAVE).SelectLastQts(userId)
	if qtsDO != nil {
		state.SetQts(qtsDO.Qts)
	} else {
		state.SetQts(0)
	}

	// state.SetSeq(int32(GetSequenceModel().CurrentSeqId(base.Int64ToString(authKeyId))))
	state.SetSeq(int32(GetSequenceModel().CurrentSeqId(base.Int32ToString(userId))))
	//
	//seqDO := dao.GetAuthSeqUpdatesDAO(dao.DB_SLAVE).SelectLastSeq(authKeyId, userId)
	//if seqDO != nil {
	//	state.SetSeq(seqDO.Seq)
	//} else {
	//	state.SetSeq(0)
	//}

	state.SetDate(int32(time.Now().Unix()))
	// TODO(@benqi): Calc unread
	state.SetUnreadCount(0)

	return state
}

//func (m *updatesModel) AddPtsToUpdatesQueue(userId, pts, peerType, peerId, updateType, messageBoxId, maxMessageBoxId int32, ) int32 {
//	do := &dataobject.UserPtsUpdatesDO{
//		UserId:          userId,
//		PeerType:		 int8(peerType),
//		PeerId:			 peerId,
//		Pts:             pts,
//		UpdateType:      updateType,
//		MessageBoxId:    messageBoxId,
//		MaxMessageBoxId: maxMessageBoxId,
//		Date2:           int32(time.Now().Unix()),
//	}
//
//	return int32(dao.GetUserPtsUpdatesDAO(dao.DB_MASTER).Insert(do))
//}

func (m *updatesModel) AddQtsToUpdatesQueue(userId, qts, updateType int32, updateData []byte) int32 {
	do := &dataobject.UserQtsUpdatesDO{
		UserId:     userId,
		UpdateType: updateType,
		UpdateData: updateData,
		Date2:      int32(time.Now().Unix()),
		Qts:        qts,
	}

	return int32(dao.GetUserQtsUpdatesDAO(dao.DB_MASTER).Insert(do))
}

func (m *updatesModel) AddSeqToUpdatesQueue(authId int64, userId, seq, updateType int32, updateData []byte) int32 {
	do := &dataobject.AuthSeqUpdatesDO{
		AuthId:     authId,
		UserId:     userId,
		UpdateType: updateType,
		UpdateData: updateData,
		Date2:      int32(time.Now().Unix()),
		Seq:        seq,
	}

	return int32(dao.GetAuthSeqUpdatesDAO(dao.DB_MASTER).Insert(do))
}

//func (m *updatesModel) GetAffectedMessage(userId, messageId int32) *mtproto.TLMessagesAffectedMessages {
//	doList := dao.GetMessageBoxesDAO(dao.DB_SLAVE).SelectPtsByGTMessageID(userId, messageId)
//	if len(doList) == 0 {
//		panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_OTHER2), fmt.Sprintf("GetAffectedMessage(%d, %d) empty", userId, messageId)))
//	}
//
//	affected := &mtproto.TLMessagesAffectedMessages{}
//	affected.Pts = doList[0].Pts
//	affected.PtsCount = int32(len(doList))
//	return affected
//}

//func updateToQueueData(update *mtproto.Update) int8 {
//	switch update.GetConstructor() {
//	case mtproto.TLConstructor_crc32_
//	}
//}

func getUpdateType(update *mtproto.Update) int8 {
	switch update.GetConstructor() {
	case mtproto.TLConstructor_CRC32_updateNewMessage:
		return PTS_UPDATE_NEW_MESSAGE
	case mtproto.TLConstructor_CRC32_updateDeleteMessages:
		return PTS_UPDATE_DELETE_MESSAGES
	case mtproto.TLConstructor_CRC32_updateReadHistoryOutbox:
		return PTS_UPDATE_READ_HISTORY_OUTBOX
	case mtproto.TLConstructor_CRC32_updateReadHistoryInbox:
		return PTS_UPDATE_READ_HISTORY_INBOX
	case mtproto.TLConstructor_CRC32_updateWebPage:
		return PTS_UPDATE_WEBPAGE
	case mtproto.TLConstructor_CRC32_updateReadMessagesContents:
		return PTS_UPDATE_READ_MESSAGE_CONENTS
	case mtproto.TLConstructor_CRC32_updateEditMessage:
		return PTS_UPDATE_EDIT_MESSAGE
	}
	return PTS_UPDATE_TYPE_UNKNOWN
}

func (m *updatesModel) AddToPtsQueue(userId, pts, ptsCount int32, update *mtproto.Update) int32 {
	// TODO(@benqi): check error
	updateData, _ := json.Marshal(update)

	do := &dataobject.UserPtsUpdatesDO{
		UserId:     userId,
		Pts:        pts,
		PtsCount:   ptsCount,
		UpdateType: getUpdateType(update),
		UpdateData: string(updateData),
		Date2:      int32(time.Now().Unix()),
	}

	return int32(dao.GetUserPtsUpdatesDAO(dao.DB_MASTER).Insert(do))
}

/*
func (m *updatesModel) GetUpdatesByGtPts(userId, pts int32) (otherUpdates []*mtproto.Update, boxIDList []int32, lastPts int32) {
	lastPts = pts
	doList := dao.GetUserPtsUpdatesDAO(dao.DB_SLAVE).SelectByGtPts(userId, pts)
	if len(doList) == 0 {
		otherUpdates = []*mtproto.Update{}
		boxIDList = []int32{}
	} else {
		boxIDList = make([]int32, 0, len(doList))
		otherUpdates = make([]*mtproto.Update, 0, len(doList))
		for _, do := range doList {
			switch do.UpdateType {
			//  case PTS_UPDATE_SHORT_MESSAGE, PTS_UPDATE_SHORT_CHAT_MESSAGE:
			case PTS_READ_HISTORY_OUTBOX:
				readHistoryOutbox := &mtproto.TLUpdateReadHistoryOutbox{Data2: &mtproto.Update_Data{
					Peer_39:  base2.ToPeerByTypeAndID(do.PeerType, do.PeerId),
					MaxId:    do.MaxMessageBoxId,
					Pts:      do.Pts,
					PtsCount: 0,
				}}
				otherUpdates = append(otherUpdates, readHistoryOutbox.To_Update())
			case PTS_READ_HISTORY_INBOX:
				readHistoryInbox := &mtproto.TLUpdateReadHistoryInbox{Data2: &mtproto.Update_Data{
					Peer_39:  base2.ToPeerByTypeAndID(do.PeerType, do.PeerId),
					MaxId:    do.MaxMessageBoxId,
					Pts:      do.Pts,
					PtsCount: 0,
				}}
				otherUpdates = append(otherUpdates, readHistoryInbox.To_Update())
			//case PTS_MESSAGE_OUTBOX, PTS_MESSAGE_INBOX:
			//	boxIDList = append(boxIDList, do.MessageBoxId)
			}

			if do.Pts > lastPts {
				lastPts = do.Pts
			}
		}
	}
	return
}
*/

func (m *updatesModel) GetUpdateListByGtPts(userId, pts int32) []*mtproto.Update {
	doList := dao.GetUserPtsUpdatesDAO(dao.DB_SLAVE).SelectByGtPts(userId, pts)
	if len(doList) == 0 {
		return []*mtproto.Update{}
	}

	updates := make([]*mtproto.Update, 0, len(doList))
	for _, do := range doList {
		update := new(mtproto.Update)
		err := json.Unmarshal([]byte(do.UpdateData), update)
		if err != nil {
			glog.Errorf("unmarshal pts's update(%d)error: %v", do.Id, err)
			continue
		}
		if getUpdateType(update) != do.UpdateType {
			glog.Errorf("update data error.")
			continue
		}
		updates = append(updates, update)
	}
	return updates
}
