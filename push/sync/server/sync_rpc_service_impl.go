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

package server

import (
	"github.com/nebulaim/telegramd/biz/core/user"
	"github.com/nebulaim/telegramd/mtproto"
	"golang.org/x/net/context"
	"sync"
	"github.com/golang/glog"
	"github.com/nebulaim/telegramd/baselib/logger"
	"github.com/nebulaim/telegramd/baselib/base"
	"time"
	"fmt"
	update3 "github.com/nebulaim/telegramd/biz/core/update"
)

type SyncServiceImpl struct {
	// status *model.OnlineStatusModel
	mu sync.RWMutex
	s  *syncServer
	// TODO(@benqi): 多个连接
	// updates map[int32]chan *zproto.PushUpdatesNotify
}

func NewSyncService(sync2 *syncServer) *SyncServiceImpl {
	s := &SyncServiceImpl{s: sync2}
	// s.status = status
	// s.updates = make(map[int32]chan *zproto.PushUpdatesNotify)
	return s
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// 推送给该用户所有设备

func (s *SyncServiceImpl) pushUpdatesToSession(state *mtproto.ClientUpdatesState, updates *mtproto.UpdatesRequest) {
	statusList, _ := user.GetOnlineByUserId(updates.GetPushUserId())
	ss := make(map[int32][]*user.SessionStatus)
	for _, status := range statusList {
		if _, ok := ss[status.ServerId]; !ok {
			ss[status.ServerId] = []*user.SessionStatus{}
		}
		ss[status.ServerId] = append(ss[status.ServerId], status)
	}

	// TODO(@benqi): 预先计算是否需要同步？
	// updatesData := updates.GetUpdates().Encode()
	// hasOnlineClient := false
	var updatesData []byte

	encodeUpdateData := func() {
		// 序列化延时
		if updatesData == nil {
			updatesData = updates.GetUpdates().Encode()
		}
		// return updatesData
	}

	for k, ss3 := range ss {
		for _, ss4 := range ss3 {
			switch updates.GetPushType() {
			case mtproto.SyncType_SYNC_TYPE_USER_NOTME:
				if updates.GetSessionId() != ss4.SessionId {
					continue
					encodeUpdateData()
				} else {
					continue
				}
			case mtproto.SyncType_SYNC_TYPE_USER_ME:
				if updates.GetSessionId() == ss4.SessionId {
					continue
					encodeUpdateData()
				} else {
					continue
				}
			case mtproto.SyncType_SYNC_TYPE_USER:
				encodeUpdateData()
			default:
				continue
			}

			// push
			pushData := &mtproto.PushUpdatesData{
				AuthKeyId:   ss4.AuthKeyId,
				SessionId:   ss4.SessionId,
				State:       state,
				UpdatesData: updatesData,
			}
			s.s.sendToSessionServer(int(k), pushData)
		}
	}
}

func updateShortMessageToMessage(userId int32, shortMessage *mtproto.TLUpdateShortMessage) *mtproto.Message {
	var (
		fromId, peerId int32
	)
	if shortMessage.GetOut() {
		fromId = userId
		peerId = shortMessage.GetUserId()
	} else {
		fromId = shortMessage.GetUserId()
		peerId = userId
	}

	message := &mtproto.TLMessage{Data2: &mtproto.Message_Data{
		Out:          shortMessage.GetOut(),
		Mentioned:    shortMessage.GetMentioned(),
		MediaUnread:  shortMessage.GetMediaUnread(),
		Silent:       shortMessage.GetSilent(),
		Id:           shortMessage.GetId(),
		FromId:       fromId,
		ToId:         &mtproto.Peer{Constructor: mtproto.TLConstructor_CRC32_peerUser, Data2: &mtproto.Peer_Data{UserId: peerId}},
		Message:      shortMessage.GetMessage(),
		Date:         shortMessage.GetDate(),
		FwdFrom:      shortMessage.GetFwdFrom(),
		ViaBotId:     shortMessage.GetViaBotId(),
		ReplyToMsgId: shortMessage.GetReplyToMsgId(),
		Entities:     shortMessage.GetEntities(),
	}}
	return message.To_Message()
}

func updateShortChatMessageToMessage(shortMessage *mtproto.TLUpdateShortChatMessage) *mtproto.Message {
	message := &mtproto.TLMessage{Data2: &mtproto.Message_Data{
		Out:          shortMessage.GetOut(),
		Mentioned:    shortMessage.GetMentioned(),
		MediaUnread:  shortMessage.GetMediaUnread(),
		Silent:       shortMessage.GetSilent(),
		Id:           shortMessage.GetId(),
		FromId:       shortMessage.GetFromId(),
		ToId:         &mtproto.Peer{Constructor: mtproto.TLConstructor_CRC32_peerChat, Data2: &mtproto.Peer_Data{UserId: shortMessage.GetChatId()}},
		Message:      shortMessage.GetMessage(),
		Date:         shortMessage.GetDate(),
		FwdFrom:      shortMessage.GetFwdFrom(),
		ViaBotId:     shortMessage.GetViaBotId(),
		ReplyToMsgId: shortMessage.GetReplyToMsgId(),
		Entities:     shortMessage.GetEntities(),
	}}
	return message.To_Message()
}

func updateShortToUpdateNewMessage(userId int32, shortMessage *mtproto.TLUpdateShortMessage) *mtproto.Update {
	updateNew := &mtproto.TLUpdateNewMessage{ Data2: &mtproto.Update_Data{
		Message_1: updateShortMessageToMessage(userId, shortMessage),
		Pts:       shortMessage.GetPts(),
		PtsCount:  shortMessage.GetPtsCount(),
	}}
	return updateNew.To_Update()
}

func updateShortChatToUpdateNewMessage(userId int32, shortMessage *mtproto.TLUpdateShortChatMessage) *mtproto.Update {
	updateNew := &mtproto.TLUpdateNewMessage{ Data2: &mtproto.Update_Data{
		Message_1: updateShortChatMessageToMessage(shortMessage),
		Pts:       shortMessage.GetPts(),
		PtsCount:  shortMessage.GetPtsCount(),
	}}
	return updateNew.To_Update()
}

// rpc
// rpc SyncUpdatesData(UpdatesRequest) returns (ClientUpdatesState);
func processUpdatesRequest(request *mtproto.UpdatesRequest) (*mtproto.ClientUpdatesState, error) {
	var (
		pushUserId = request.GetPushUserId()
		pts, ptsCount int32
		seq = int32(0)
		updates = request.GetUpdates()
		date = int32(time.Now().Unix())
	)

	switch updates.GetConstructor() {
	case mtproto.TLConstructor_CRC32_updateShortMessage:
		shortMessage := updates.To_UpdateShortMessage()
		pts = int32(update3.NextPtsId(base.Int32ToString(pushUserId)))
		ptsCount = 1
		shortMessage.SetPts(pts)
		shortMessage.SetPtsCount(ptsCount)
		update3.AddToPtsQueue(pushUserId, pts, ptsCount, updateShortToUpdateNewMessage(pushUserId, shortMessage))
	case mtproto.TLConstructor_CRC32_updateShortChatMessage:
		shortMessage := updates.To_UpdateShortChatMessage()
		pts = int32(update3.NextPtsId(base.Int32ToString(pushUserId)))
		ptsCount = 1
		shortMessage.SetPts(pts)
		shortMessage.SetPtsCount(ptsCount)
		update3.AddToPtsQueue(pushUserId, pts, ptsCount, updateShortChatToUpdateNewMessage(pushUserId, shortMessage))
	case mtproto.TLConstructor_CRC32_updateShort:
		short := updates.To_UpdateShort()
		short.SetDate(date)
	case mtproto.TLConstructor_CRC32_updates:
		updates2 := updates.To_Updates()
		totalPtsCount := int32(0)
		for _, update := range updates2.GetUpdates() {
			switch update.GetConstructor() {
			case mtproto.TLConstructor_CRC32_updateNewMessage,
				 mtproto.TLConstructor_CRC32_updateDeleteMessages,
				 mtproto.TLConstructor_CRC32_updateReadHistoryOutbox,
				 mtproto.TLConstructor_CRC32_updateReadHistoryInbox,
				 mtproto.TLConstructor_CRC32_updateWebPage,
				 mtproto.TLConstructor_CRC32_updateReadMessagesContents,
				 mtproto.TLConstructor_CRC32_updateEditMessage:

				pts = int32(update3.NextPtsId(base.Int32ToString(pushUserId)))
				ptsCount = 1
				totalPtsCount += 1

				// @benqi: 以上都有Pts和PtsCount
				update.Data2.Pts = pts
				update.Data2.PtsCount = ptsCount
				update3.AddToPtsQueue(pushUserId, pts, ptsCount, update)
			}
		}

		// 有可能有多个
		ptsCount = totalPtsCount
		updates2.SetDate(date)
		updates2.SetSeq(seq)
	default:
		err := fmt.Errorf("invalid updates data: {%d}", updates.GetConstructor())
		// glog.Error(err)
		return nil, err
	}

	state := &mtproto.ClientUpdatesState{
		Pts:      pts,
		PtsCount: ptsCount,
		Date:     date,
	}

	return state, nil
}

func (s *SyncServiceImpl) SyncUpdatesData(ctx context.Context, request *mtproto.UpdatesRequest) (reply *mtproto.ClientUpdatesState, err error) {
	glog.Infof("syncUpdatesData - request: {%v}", request)

	reply, err = processUpdatesRequest(request)
	if err == nil {
		s.pushUpdatesToSession(reply, request)
		glog.Infof("syncUpdatesData - reply: %s", logger.JsonDebugData(reply))
	} else {
		glog.Error(err)
	}

	return
}

// rpc PushUpdatesData(UpdatesRequest) returns (VoidRsp);
func (s *SyncServiceImpl) PushUpdatesData(ctx context.Context, request *mtproto.UpdatesRequest) (reply *mtproto.VoidRsp, err error) {
	glog.Infof("syncUpdatesData - request: {%v}", request)

	var state *mtproto.ClientUpdatesState
	state, err = processUpdatesRequest(request)
	if err == nil {
		s.pushUpdatesToSession(state, request)
		glog.Infof("syncUpdatesData - reply: %s", logger.JsonDebugData(state))
		reply = &mtproto.VoidRsp{}
	} else {
		glog.Error(err)
	}

	return
}

// rpc PushUpdatesDataList(UpdatesListRequest) returns (VoidRsp);
//func (s *SyncServiceImpl) PushUpdatesDataList(ctx context.Context, request *mtproto.UpdatesListRequest) (reply *mtproto.VoidRsp, err error) {
//	return
//}
