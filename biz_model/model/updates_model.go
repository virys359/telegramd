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
	//"github.com/nebulaim/telegramd/mtproto"
	//"github.com/nebulaim/telegramd/biz_model/dal/dao"
	//"time"
	"time"
	"github.com/nebulaim/telegramd/mtproto"
	"github.com/nebulaim/telegramd/biz_model/dal/dao"
	"github.com/nebulaim/telegramd/biz_model/dal/dataobject"
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

	seqDO := dao.GetAuthSeqUpdatesDAO(dao.DB_SLAVE).SelectLastSeq(authKeyId, userId)
	if seqDO != nil {
		state.SetSeq(seqDO.Seq)
	} else {
		state.SetSeq(0)
	}

	state.SetDate(int32(time.Now().Unix()))
	// TODO(@benqi): Calc unread
	state.SetUnreadCount(0)

	return state
}

func (m *updatesModel) AddPtsToUpdatesQueue(userId, pts, peerType, peerId, updateType, messageBoxId, maxMessageBoxId int32, ) int32 {
	do := &dataobject.UserPtsUpdatesDO{
		UserId:          userId,
		PeerType:		 int8(peerType),
		PeerId:			 peerId,
		Pts:             pts,
		UpdateType:      updateType,
		MessageBoxId:    messageBoxId,
		MaxMessageBoxId: maxMessageBoxId,
		Date2:           int32(time.Now().Unix()),
	}

	return int32(dao.GetUserPtsUpdatesDAO(dao.DB_MASTER).Insert(do))
}

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

