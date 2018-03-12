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
	"github.com/golang/glog"
	"github.com/nebulaim/telegramd/baselib/base"
	"github.com/nebulaim/telegramd/baselib/logger"
	base2 "github.com/nebulaim/telegramd/biz_model/base"
	"github.com/nebulaim/telegramd/biz_model/model"
	"github.com/nebulaim/telegramd/mtproto"
	"golang.org/x/net/context"
	"sync"
	"time"
)

type SyncServiceImpl struct {
	// status *model.OnlineStatusModel

	mu sync.RWMutex
	s  *syncServer
	// TODO(@benqi): 多个连接
	// updates map[int32]chan *zproto.PushUpdatesNotify
}

func NewSyncService(sync *syncServer) *SyncServiceImpl {
	s := &SyncServiceImpl{s: sync}
	// s.status = status
	// s.updates = make(map[int32]chan *zproto.PushUpdatesNotify)
	return s
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// 推送给该用户所有设备
func (s *SyncServiceImpl) pushToUserUpdates(userId int32, updates *mtproto.Updates) {
	// pushRawData := updates.Encode()

	statusList, _ := model.GetOnlineStatusModel().GetOnlineByUserId(userId)
	ss := make(map[int32][]*model.SessionStatus)
	for _, status := range statusList {
		if _, ok := ss[status.ServerId]; !ok {
			ss[status.ServerId] = []*model.SessionStatus{}
		}
		// 会不会出问题？？
		ss[status.ServerId] = append(ss[status.ServerId], status)
	}

	rawData := updates.Encode()
	for k, ss3 := range ss {
		//// glog.Infof("DeliveryUpdates: k: {%v}, v: {%v}", k, ss3)
		//go s.withReadLock(func() {
		for _, ss4 := range ss3 {
			pushData := &mtproto.PushUpdatesData{
				ClientId: &mtproto.PushClientID{
					AuthKeyId:       ss4.AuthKeyId,
					SessionId:       ss4.SessionId,
					NetlibSessionId: ss4.NetlibSessionId,
				},
				// RawDataHeader:
				RawData: rawData,
			}
			s.s.sendToSessionServer(int(k), pushData)
		}
		//})
	}
}

// 推送给该用户剔除指定设备后的所有设备
func (s *SyncServiceImpl) pushToUserUpdatesNotMe(userId int32, netlibSessionId int64, updates *mtproto.Updates) {
	// pushRawData := updates.Encode()

	statusList, _ := model.GetOnlineStatusModel().GetOnlineByUserId(userId)
	ss := make(map[int32][]*model.SessionStatus)
	for _, status := range statusList {
		if _, ok := ss[status.ServerId]; !ok {
			ss[status.ServerId] = []*model.SessionStatus{}
		}
		// 会不会出问题？？
		ss[status.ServerId] = append(ss[status.ServerId], status)
	}

	rawData := updates.Encode()
	for k, ss3 := range ss {
		for _, ss4 := range ss3 {
			if ss4.NetlibSessionId != netlibSessionId {
				pushData := &mtproto.PushUpdatesData{
					ClientId: &mtproto.PushClientID{
						AuthKeyId:       ss4.AuthKeyId,
						SessionId:       ss4.SessionId,
						NetlibSessionId: ss4.NetlibSessionId,
					},
					// RawDataHeader:
					RawData: rawData,
				}
				s.s.sendToSessionServer(int(k), pushData)
			}
		}
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
func (s *SyncServiceImpl) SyncUpdateShortMessage(ctx context.Context, request *mtproto.SyncShortMessageRequest) (reply *mtproto.ClientUpdatesState, err error) {
	glog.Infof("syncUpdateShortMessage - request: {%v}", request)

	var (
		userId, peerId int32
		pts, ptsCount int32
		updateType int32
	)

	pushData := request.GetPushData()

	userId = pushData.GetPushUserId()
	shortMessage := pushData.GetPushData()
	peerId = request.GetSenderUserId()
	updateType = model.PTS_MESSAGE_OUTBOX

	pts = int32(model.GetSequenceModel().NextPtsId(base.Int32ToString(userId)))
	ptsCount = int32(1)
	pushData.GetPushData().SetPts(pts)
	shortMessage.SetPtsCount(ptsCount)

	// save pts
	model.GetUpdatesModel().AddPtsToUpdatesQueue(userId, pts, base2.PEER_USER, peerId, updateType, shortMessage.GetId(), 0)

	// push
	if pushData.GetPushType() == mtproto.SyncType_SYNC_TYPE_USER_NOTME {
		s.pushToUserUpdatesNotMe(userId, request.GetClientId().GetNetlibSessionId(), shortMessage.To_Updates())
	} else {
		s.pushToUserUpdates(pushData.GetPushUserId(), shortMessage.To_Updates())
	}

	reply = &mtproto.ClientUpdatesState{
		Pts:      pts,
		PtsCount: ptsCount,
		Date:     int32(time.Now().Unix()),
	}

	glog.Infof("DeliveryPushUpdateShortMessage - reply: %s", logger.JsonDebugData(reply))
	return
}

func (s *SyncServiceImpl) PushUpdateShortMessage(ctx context.Context, request *mtproto.UpdateShortMessageRequest) (reply *mtproto.VoidRsp, err error) {
	glog.Infof("pushUpdateShortMessage - request: {%v}", request)

	var (
		userId, peerId int32
		pts, ptsCount int32
		updateType int32
	)

	pushData := request.GetPushData()

	userId = pushData.GetPushUserId()
	shortMessage := pushData.GetPushData()
	peerId = request.GetSenderUserId()
	updateType = model.PTS_MESSAGE_INBOX

	pts = int32(model.GetSequenceModel().NextPtsId(base.Int32ToString(userId)))
	ptsCount = int32(1)
	pushData.GetPushData().SetPts(pts)
	shortMessage.SetPtsCount(ptsCount)

	// save pts
	model.GetUpdatesModel().AddPtsToUpdatesQueue(userId, pts, base2.PEER_USER, peerId, updateType, shortMessage.GetId(), 0)

	// push
	s.pushToUserUpdates(pushData.GetPushUserId(), shortMessage.To_Updates())

	reply = &mtproto.VoidRsp{}
	glog.Infof("pushUpdateShortMessage - reply: %s", logger.JsonDebugData(reply))
	return
}

func (s *SyncServiceImpl) SyncUpdateShortChatMessage(ctx context.Context, request *mtproto.SyncShortChatMessageRequest) (reply *mtproto.ClientUpdatesState, err error) {
	glog.Infof("syncUpdateShortChatMessage - request: {%v}", request)

	// TODO(@benqi): Check deliver valid!
	pushData := request.GetPushData()

	var (
		pts, ptsCount int32
		updateType int32
	)

	shortChatMessage := pushData.GetPushData()
	// var outgoing = request.GetSenderUserId() == pushData.GetPushUserId()
	updateType = model.PTS_MESSAGE_OUTBOX

	pts = int32(model.GetSequenceModel().NextPtsId(base.Int32ToString(pushData.GetPushUserId())))
	ptsCount = int32(1)
	pushData.GetPushData().SetPts(pts)
	shortChatMessage.SetPtsCount(ptsCount)

	// save pts
	model.GetUpdatesModel().AddPtsToUpdatesQueue(pushData.GetPushUserId(), pts, base2.PEER_CHAT, request.GetPeerChatId(), updateType, shortChatMessage.GetId(), 0)
	s.pushToUserUpdatesNotMe(pushData.GetPushUserId(), request.GetClientId().GetNetlibSessionId(), shortChatMessage.To_Updates())

	reply = &mtproto.ClientUpdatesState{
		Pts:      pts,
		PtsCount: ptsCount,
		Date:     int32(time.Now().Unix()),
	}
	glog.Infof("syncUpdateShortChatMessage - reply: %s", logger.JsonDebugData(reply))
	return
}

func (s *SyncServiceImpl) PushUpdateShortChatMessage(ctx context.Context, request *mtproto.UpdateShortChatMessageRequest) (reply *mtproto.VoidRsp, err error) {
	glog.Infof("pushUpdateShortChatMessage - request: {%v}", request)
	return nil, nil


	// TODO(@benqi): Check deliver valid!
	pushDatas := request.GetPushDatas()

	var (
		pts, ptsCount int32
		updateType int32
	)

	for _, pushData := range pushDatas {
		shortChatMessage := pushData.GetPushData()
		updateType = model.PTS_MESSAGE_INBOX

		pts = int32(model.GetSequenceModel().NextPtsId(base.Int32ToString(pushData.GetPushUserId())))
		ptsCount = int32(1)
		pushData.GetPushData().SetPts(pts)
		shortChatMessage.SetPtsCount(ptsCount)

		// save pts
		model.GetUpdatesModel().AddPtsToUpdatesQueue(pushData.GetPushUserId(), pts, base2.PEER_CHAT, request.GetPeerChatId(), updateType, shortChatMessage.GetId(), 0)
		s.pushToUserUpdates(pushData.GetPushUserId(), shortChatMessage.To_Updates())
	}

	reply = &mtproto.VoidRsp{}
	glog.Infof("pushUpdateShortChatMessage - reply: %s", logger.JsonDebugData(reply))
	return
}

/*
func (s *SyncServiceImpl) GetUpdatesData(ctx context.Context, request *zproto.GetUpdatesDataRequest) (reply *zproto.UpdatesDatasRsp, err error) {
	glog.Infof("GetUpdatesData - request: {%v}", request)

/ *
	// TODO(@benqi):
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("UpdatesGetDifference - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))
	difference := mtproto.NewTLUpdatesDifference()
	otherUpdates := []*mtproto.Update{}

	lastPts := request.GetPts()
	doList := dao.GetUserPtsUpdatesDAO(dao.DB_SLAVE).SelectByGtPts(md.UserId, request.GetPts())
	boxIDList := make([]int32, 0, len(doList))
	for _, do := range doList {
		switch do.UpdateType {
		case model.PTS_READ_HISTORY_OUTBOX:
			readHistoryOutbox := mtproto.NewTLUpdateReadHistoryOutbox()
			readHistoryOutbox.SetPeer(base.ToPeerByTypeAndID(do.PeerType, do.PeerId))
			readHistoryOutbox.SetMaxId(do.MaxMessageBoxId)
			readHistoryOutbox.SetPts(do.Pts)
			readHistoryOutbox.SetPtsCount(0)
			otherUpdates = append(otherUpdates, readHistoryOutbox.To_Update())
		case model.PTS_READ_HISTORY_INBOX:
			readHistoryInbox := mtproto.NewTLUpdateReadHistoryInbox()
			readHistoryInbox.SetPeer(base.ToPeerByTypeAndID(do.PeerType, do.PeerId))
			readHistoryInbox.SetMaxId(do.MaxMessageBoxId)
			readHistoryInbox.SetPts(do.Pts)
			readHistoryInbox.SetPtsCount(0)
			otherUpdates = append(otherUpdates, readHistoryInbox.To_Update())
		case model.PTS_MESSAGE_OUTBOX, model.PTS_MESSAGE_INBOX:
			boxIDList = append(boxIDList, do.MessageBoxId)
		}

		if do.Pts > lastPts {
			lastPts = do.Pts
		}
	}

	if len(boxIDList) > 0 {
		messages := model.GetMessageModel().GetMessagesByPeerAndMessageIdList2(md.UserId, boxIDList)
		// messages := model.GetMessageModel().GetMessagesByGtPts(md.UserId, request.Pts)
		userIdList := []int32{}
		chatIdList := []int32{}

		for _, m := range messages {
			switch m.GetConstructor()  {

			case mtproto.TLConstructor_CRC32_message:
				m2 := m.To_Message()
				userIdList = append(userIdList, m2.GetFromId())
				p := base.FromPeer(m2.GetToId())
				switch p.PeerType {
				case base.PEER_SELF, base.PEER_USER:
					userIdList = append(userIdList, p.PeerId)
				case base.PEER_CHAT:
					chatIdList = append(chatIdList, p.PeerId)
				case base.PEER_CHANNEL:
					// TODO(@benqi): add channel
				}
				//peer := base.FromPeer(m2.GetToId())
				//switch peer.PeerType {
				//case base.PEER_USER:
				//    userIdList = append(userIdList, peer.PeerId)
				//case base.PEER_CHAT:
				//    chatIdList = append(chatIdList, peer.PeerId)
				//case base.PEER_CHANNEL:
				//    // TODO(@benqi): add channel
				//}
			case mtproto.TLConstructor_CRC32_messageService:
				m2 := m.To_MessageService()
				userIdList = append(userIdList, m2.GetFromId())
				chatIdList = append(chatIdList, m2.GetToId().GetData2().GetChatId())
			case mtproto.TLConstructor_CRC32_messageEmpty:
			}
			difference.Data2.NewMessages = append(difference.Data2.NewMessages, m)
		}

		if len(userIdList) > 0 {
			usersList := model.GetUserModel().GetUserList(userIdList)
			for _, u := range usersList {
				if u.GetId() == md.UserId {
					u.SetSelf(true)
				} else {
					u.SetSelf(false)
				}
				u.SetContact(true)
				u.SetMutualContact(true)
				difference.Data2.Users = append(difference.Data2.Users, u.To_User())
			}
		}
	}

	difference.SetOtherUpdates(otherUpdates)

	state := mtproto.NewTLUpdatesState()
	// TODO(@benqi): Pts通过规则计算出来
	state.SetPts(lastPts)
	state.SetDate(int32(time.Now().Unix()))
	state.SetUnreadCount(0)
	state.SetSeq(int32(model.GetSequenceModel().CurrentSeqId(base2.Int32ToString(md.UserId))))

	difference.SetState(state.To_Updates_State())

	// dao.GetAuthUpdatesStateDAO(dao.DB_MASTER).UpdateState(request.GetPts(), request.GetQts(), request.GetDate(), md.AuthId, md.UserId)
	//glog.Infof("UpdatesGetDifference - reply: %s", difference)
	//return difference.To_Updates_Difference(), nil

* /
	glog.Infof("GetUpdatesData - reply {%v}", reply)
	return
}

//
//func (s *SyncServiceImpl) DeliveryUpdates2(ctx context.Context, request *mtproto.UpdatesRequest) (reply *mtproto.DeliveryRsp, err error) {
//	glog.Infof("DeliveryPushUpdates - request: {%v}", request)
//
//	var seq, replySeq int32
//	now := int32(time.Now().Unix())
//
//	// TODO(@benqi): Check deliver valid!
//	pushDatas := request.GetPushDatas()
//	for _, pushData := range pushDatas {
//		updates := pushData.GetPushData()
//		// pushRawData := updates.Encode()
//		statusList, _ := model.GetOnlineStatusModel().GetOnlineByUserId(pushData.GetPushUserId())
//		switch pushData.GetPushType() {
//		case mtproto.SyncType_SYNC_TYPE_USER:
//			for _, status := range statusList {
//				seq = int32(model.GetSequenceModel().NextSeqId(base.Int64ToString(status.AuthKeyId)))
//				if status.AuthKeyId == request.GetSenderAuthKeyId() {
//					replySeq = seq
//				}
//				updates.SetDate(now)
//				updates.SetSeq(seq)
//
//				//update := &zproto.PushUpdatesNotify{
//				//	AuthKeyId:       status.AuthKeyId,
//				//	SessionId:       status.SessionId,
//				//	NetlibSessionId: status.NetlibSessionId,
//				//	// RawData:         updates.Encode(),
//				//}
//				//
//				//go s.withReadLock(func() {
//				//	s.updates[status.ServerId] <- update
//				//})
//			}
//		case mtproto.SyncType_SYNC_TYPE_USER_NOTME:
//			for _, status := range statusList {
//				seq = int32(model.GetSequenceModel().NextSeqId(base.Int64ToString(status.AuthKeyId)))
//				if status.AuthKeyId == request.GetSenderAuthKeyId() {
//					replySeq = seq
//					continue
//				}
//				updates.SetDate(now)
//				updates.SetSeq(seq)
//
//				//update := &zproto.PushUpdatesNotify{
//				//	AuthKeyId:       status.AuthKeyId,
//				//	SessionId:       status.SessionId,
//				//	NetlibSessionId: status.NetlibSessionId,
//				//	// RawData:         updates.Encode(),
//				//}
//				//
//				//go s.withReadLock(func() {
//				//	s.updates[status.ServerId] <- update
//				//})
//			}
//		case mtproto.SyncType_SYNC_TYPE_AUTH_KEY:
//			for _, status := range statusList {
//				seq = int32(model.GetSequenceModel().NextSeqId(base.Int64ToString(status.AuthKeyId)))
//				if status.AuthKeyId != request.GetSenderAuthKeyId() {
//					replySeq = seq
//				}
//				updates.SetDate(now)
//				updates.SetSeq(seq)
//
//				//update := &zproto.PushUpdatesNotify{
//				//	AuthKeyId:       status.AuthKeyId,
//				//	SessionId:       status.SessionId,
//				//	NetlibSessionId: status.NetlibSessionId,
//				//	// RawData:         updates.Encode(),
//				//}
//				//
//				//go s.withReadLock(func() {
//				//	s.updates[status.ServerId] <- update
//				//})
//			}
//		case mtproto.SyncType_SYNC_TYPE_AUTH_KEY_USERNOTME:
//			for _, status := range statusList {
//				seq = int32(model.GetSequenceModel().NextSeqId(base.Int64ToString(status.AuthKeyId)))
//				if status.AuthKeyId == request.GetSenderAuthKeyId() {
//					replySeq = seq
//					continue
//				}
//				updates.SetDate(now)
//				updates.SetSeq(seq)
//
//				//update := &zproto.PushUpdatesNotify{
//				//	AuthKeyId:       status.AuthKeyId,
//				//	SessionId:       status.SessionId,
//				//	NetlibSessionId: status.NetlibSessionId,
//				//	// RawData:         updates.Encode(),
//				//}
//				//
//				//go s.withReadLock(func() {
//				//	s.updates[status.ServerId] <- update
//				//})
//			}
//		case mtproto.SyncType_SYNC_TYPE_AUTH_KEY_USER:
//			for _, status := range statusList {
//				seq = int32(model.GetSequenceModel().NextSeqId(base.Int64ToString(status.AuthKeyId)))
//				if status.AuthKeyId == request.GetSenderAuthKeyId() {
//					replySeq = seq
//				}
//				updates.SetDate(now)
//				updates.SetSeq(seq)
//
//				//update := &zproto.PushUpdatesNotify{
//				//	AuthKeyId:       status.AuthKeyId,
//				//	SessionId:       status.SessionId,
//				//	NetlibSessionId: status.NetlibSessionId,
//				//	// RawData:         updates.Encode(),
//				//}
//				//
//				//go s.withReadLock(func() {
//				//	s.updates[status.ServerId] <- update
//				//})
//			}
//		default:
//		}
//
//		//ss := make(map[int32][]*model.SessionStatus)
//		//for _, status := range statusList {
//		//	if _, ok := ss[status.ServerId]; !ok {
//		//		ss[status.ServerId] = []*model.SessionStatus{}
//		//	}
//		//	// 会不会出问题？？
//		//	ss[status.ServerId] = append(ss[status.ServerId], status)
//		//}
//		//
//		//for k, ss3 := range ss {
//		//	// glog.Infof("DeliveryUpdates: k: {%v}, v: {%v}", k, ss3)
//		//	go s.withReadLock(func() {
//		//		for _, ss4 := range ss3 {
//		//			update := &zproto.PushUpdatesData{
//		//				AuthKeyId:       ss4.AuthKeyId,
//		//				SessionId:       ss4.SessionId,
//		//				NetlibSessionId: ss4.NetlibSessionId,
//		//				// RawData:         pushRawData,
//		//			}
//		//
//		//
//		//
//		//			s.updates[k] <- update
//		//
//		//		}
//		//	})
//		//}
//
//		// 			updates := pushData.GetPushData()
//		// _ = updates
//	}
//
//	//seq = int32(model.GetSequenceModel().NextSeqId(base.Int64ToString(request.GetSenderAuthKeyId())))
//	reply = &mtproto.DeliveryRsp{
//		Seq:  replySeq,
//		Date: now,
//	}
//
//	glog.Infof("DeliveryPushUpdateShortMessage - reply: %s", logger.JsonDebugData(reply))
//	return
//}
*/
