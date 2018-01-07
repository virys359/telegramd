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

package rpc

import (
	"errors"
	"github.com/golang/glog"
	"github.com/nebulaim/telegramd/base/base"
	"github.com/nebulaim/telegramd/base/logger"
	base2 "github.com/nebulaim/telegramd/biz_model/base"
	"github.com/nebulaim/telegramd/biz_model/model"
	"github.com/nebulaim/telegramd/mtproto"
	"github.com/nebulaim/telegramd/zproto"
	"golang.org/x/net/context"
	"sync"
	"time"
)

type SyncServiceImpl struct {
	// status *model.OnlineStatusModel

	mu sync.RWMutex

	// TODO(@benqi): 多个连接
	updates map[int32]chan *zproto.PushUpdatesData
}

func NewSyncService() *SyncServiceImpl {
	s := &SyncServiceImpl{}
	// s.status = status
	s.updates = make(map[int32]chan *zproto.PushUpdatesData)
	return s
}

func (s *SyncServiceImpl) withReadLock(f func()) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	f()
}

func (s *SyncServiceImpl) withWriteLock(f func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	f()
}

func (s *SyncServiceImpl) unsafeExpire(sid int32) {
	if buf, ok := s.updates[sid]; ok {
		close(buf)
	}
	delete(s.updates, sid)
}

func (s *SyncServiceImpl) PushUpdatesStream(auth *zproto.ServerAuthReq, stream zproto.RPCSync_PushUpdatesStreamServer) error {
	// TODO(@benqi): chan数量
	var update chan *zproto.PushUpdatesData = make(chan *zproto.PushUpdatesData, 1000)

	var err error
	s.withWriteLock(func() {
		if _, ok := s.updates[auth.ServerId]; ok {
			err = errors.New("already connected")
			glog.Errorf("PushUpdatesStream - %s\n", err)
			return
		}
		s.updates[auth.ServerId] = update
	})

	if err != nil {
		return err
	}

	defer s.withWriteLock(func() { s.unsafeExpire(auth.ServerId) })

	for {
		select {
		case <-stream.Context().Done():
			err = stream.Context().Err()
			glog.Errorf("PushUpdatesStream - %s\n", err)
			return stream.Context().Err()
		case data := <-update:
			if err = stream.Send(data); err != nil {
				return err
			}

			glog.Infof("PushUpdatesStream: update: {%v}", data)
		}
	}
	return nil
}

func (s *SyncServiceImpl) DeliveryUpdates(ctx context.Context, deliver *zproto.DeliveryUpdatesToUsers) (reply *zproto.VoidRsp, err error) {
	glog.Infof("DeliveryUpdates: {%v}", deliver)

	statusList, err := model.GetOnlineStatusModel().GetOnlineByUserIdList(deliver.SendtoUserIdList)
	ss := make(map[int32][]*model.SessionStatus)
	for _, status := range statusList {
		if _, ok := ss[status.ServerId]; !ok {
			ss[status.ServerId] = []*model.SessionStatus{}
		}
		// 会不会出问题？？
		ss[status.ServerId] = append(ss[status.ServerId], status)
	}
	for k, ss3 := range ss {
		// glog.Infof("DeliveryUpdates: k: {%v}, v: {%v}", k, ss3)

		go s.withReadLock(func() {
			for _, ss4 := range ss3 {
				update := &zproto.PushUpdatesData{}
				update.AuthKeyId = ss4.AuthKeyId
				update.SessionId = ss4.SessionId
				update.NetlibSessionId = ss4.NetlibSessionId
				// update.RawDataHeader = deliver.RawDataHeader
				update.RawData = deliver.RawData

				glog.Infof("DeliveryUpdates: update: {%v}", update)
				s.updates[k] <- update
			}
		})
	}

	reply = &zproto.VoidRsp{}
	return
}

func (s *SyncServiceImpl) DeliveryUpdatesNotMe(ctx context.Context, deliver *zproto.DeliveryUpdatesToUsers) (reply *zproto.VoidRsp, err error) {
	glog.Infof("DeliveryUpdates: {%v}", deliver)

	statusList, err := model.GetOnlineStatusModel().GetOnlineByUserIdList(deliver.SendtoUserIdList)
	ss := make(map[int32][]*model.SessionStatus)
	for _, status := range statusList {
		if _, ok := ss[status.ServerId]; !ok {
			ss[status.ServerId] = []*model.SessionStatus{}
		}
		// 会不会出问题？？
		ss[status.ServerId] = append(ss[status.ServerId], status)
	}
	for k, ss3 := range ss {
		// glog.Infof("DeliveryUpdates: k: {%v}, v: {%v}", k, ss3)

		go s.withReadLock(func() {
			for _, ss4 := range ss3 {
				if ss4.NetlibSessionId != deliver.MyNetlibSessionId {
					update := &zproto.PushUpdatesData{}
					update.AuthKeyId = ss4.AuthKeyId
					update.SessionId = ss4.SessionId
					update.NetlibSessionId = ss4.NetlibSessionId
					// update.RawDataHeader = deliver.RawDataHeader
					update.RawData = deliver.RawData

					glog.Infof("DeliveryUpdates: update: {%v}", update)
					s.updates[k] <- update
				} else {
					glog.Infof("Not delivery me: {%v}", ss4)
				}
			}
		})
	}

	reply = &zproto.VoidRsp{}
	return
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// 推送给该用户所有设备
func (s *SyncServiceImpl) pushToUserUpdates(userId int32, updates *mtproto.Updates) {
	pushRawData := updates.Encode()

	statusList, _ := model.GetOnlineStatusModel().GetOnlineByUserId(userId)
	ss := make(map[int32][]*model.SessionStatus)
	for _, status := range statusList {
		if _, ok := ss[status.ServerId]; !ok {
			ss[status.ServerId] = []*model.SessionStatus{}
		}
		// 会不会出问题？？
		ss[status.ServerId] = append(ss[status.ServerId], status)
	}

	for k, ss3 := range ss {
		// glog.Infof("DeliveryUpdates: k: {%v}, v: {%v}", k, ss3)
		go s.withReadLock(func() {
			for _, ss4 := range ss3 {
				update := &zproto.PushUpdatesData{
					AuthKeyId:       ss4.AuthKeyId,
					SessionId:       ss4.SessionId,
					NetlibSessionId: ss4.NetlibSessionId,
					RawData:         pushRawData,
				}
				s.updates[k] <- update
			}
		})
	}
}

// 推送给该用户剔除指定设备后的所有设备
func (s *SyncServiceImpl) pushToUserUpdatesNotMe(userId int32, netlibSessionId int64, updates *mtproto.Updates) {
	pushRawData := updates.Encode()

	statusList, _ := model.GetOnlineStatusModel().GetOnlineByUserId(userId)
	ss := make(map[int32][]*model.SessionStatus)
	for _, status := range statusList {
		if _, ok := ss[status.ServerId]; !ok {
			ss[status.ServerId] = []*model.SessionStatus{}
		}
		// 会不会出问题？？
		ss[status.ServerId] = append(ss[status.ServerId], status)
	}

	for k, ss3 := range ss {
		// glog.Infof("DeliveryUpdates: k: {%v}, v: {%v}", k, ss3)
		go s.withReadLock(func() {
			for _, ss4 := range ss3 {
				if ss4.NetlibSessionId != netlibSessionId {
					update := &zproto.PushUpdatesData{
						AuthKeyId:       ss4.AuthKeyId,
						SessionId:       ss4.SessionId,
						NetlibSessionId: ss4.NetlibSessionId,
						RawData:         pushRawData,
					}

					s.updates[k] <- update
				}
			}
		})
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////////
func (s *SyncServiceImpl) DeliveryUpdateShortMessage(ctx context.Context, request *zproto.UpdateShortMessageRequest) (reply *zproto.DeliveryRsp, err error) {
	glog.Infof("DeliveryPushUpdateShortMessage - request: {%v}", request)

	// TODO(@benqi): Check deliver valid!
	pushDatas := request.GetPushDatas()
	isPeerSelf := request.GetSenderUserId() == request.GetPeerUserId()

	var (
		userId, peerId int32
		pts, ptsCount int32
		updateType int32
	)

	for _, pushData := range pushDatas {
		userId = pushData.GetPushUserId()
		shortMessage := pushData.GetPushData()
		var outgoing = request.GetSenderUserId() == userId
		if outgoing {
			// outbox
			//userId = deliver.GetSenderUserId()
			if isPeerSelf {
				peerId = request.GetSenderUserId()
			} else {
				peerId = request.GetPeerUserId()
			}
			updateType = model.PTS_MESSAGE_OUTBOX
		} else {
			// inbox
			peerId = request.GetSenderUserId()
			updateType = model.PTS_MESSAGE_INBOX
		}

		pts = int32(model.GetSequenceModel().NextPtsId(base.Int32ToString(userId)))
		ptsCount = int32(1)
		pushData.GetPushData().SetPts(pts)
		shortMessage.SetPtsCount(ptsCount)

		// save pts
		model.GetUpdatesModel().AddPtsToUpdatesQueue(userId, pts, base2.PEER_USER, peerId, updateType, shortMessage.GetId(), 0)

		// push
		if pushData.GetPushType() == zproto.SyncType_SYNC_TYPE_USER_NOTME {
			s.pushToUserUpdatesNotMe(userId, request.GetSenderNetlibSessionId(), shortMessage.To_Updates())
		} else {
			s.pushToUserUpdates(pushData.GetPushUserId(), shortMessage.To_Updates())
		}

		if userId == request.GetSenderUserId() {
			// reply
			reply = &zproto.DeliveryRsp{
				Pts:      pts,
				PtsCount: ptsCount,
				Date:     int32(time.Now().Unix()),
			}
		}
	}

	glog.Infof("DeliveryPushUpdateShortMessage - reply: %s", logger.JsonDebugData(reply))
	return
}

func (s *SyncServiceImpl) DeliveryUpdatShortChatMessage(ctx context.Context, request *zproto.UpdatShortChatMessageRequest) (reply *zproto.DeliveryRsp, err error) {
	glog.Infof("DeliveryPushUpdatShortChatMessage - request: {%v}", request)

	// TODO(@benqi): Check deliver valid!
	pushDatas := request.GetPushDatas()

	var (
		pts, ptsCount int32
		updateType int32
	)

	for _, pushData := range pushDatas {
		shortChatMessage := pushData.GetPushData()
		var outgoing = request.GetSenderUserId() == pushData.GetPushUserId()
		if outgoing {
			// outbox
			updateType = model.PTS_MESSAGE_OUTBOX
		} else {
			// inbox
			updateType = model.PTS_MESSAGE_INBOX
		}

		pts = int32(model.GetSequenceModel().NextPtsId(base.Int32ToString(pushData.GetPushUserId())))
		ptsCount = int32(1)
		pushData.GetPushData().SetPts(pts)
		shortChatMessage.SetPtsCount(ptsCount)

		// save pts
		model.GetUpdatesModel().AddPtsToUpdatesQueue(pushData.GetPushUserId(), pts, base2.PEER_CHAT, request.GetPeerChatId(), updateType, shortChatMessage.GetId(), 0)

		// push
		if pushData.GetPushType() == zproto.SyncType_SYNC_TYPE_USER_NOTME {
			s.pushToUserUpdatesNotMe(pushData.GetPushUserId(), request.GetSenderNetlibSessionId(), shortChatMessage.To_Updates())
		} else {
			s.pushToUserUpdates(pushData.GetPushUserId(), shortChatMessage.To_Updates())
		}

		// reply
		if pushData.GetPushUserId() == request.GetSenderUserId() {
			reply = &zproto.DeliveryRsp{
				Pts:      pts,
				PtsCount: ptsCount,
				Date:     int32(time.Now().Unix()),
			}
		}
	}

	glog.Infof("DeliveryPushUpdateShortMessage - reply: %s", logger.JsonDebugData(reply))
	return
}

func (s *SyncServiceImpl) DeliveryUpdates2(ctx context.Context, request *zproto.UpdatesRequest) (reply *zproto.DeliveryRsp, err error) {
	glog.Infof("DeliveryPushUpdates - request: {%v}", request)

	var seq, replySeq int32
	now := int32(time.Now().Unix())

	// TODO(@benqi): Check deliver valid!
	pushDatas := request.GetPushDatas()
	for _, pushData := range pushDatas {
		updates := pushData.GetPushData()
		// pushRawData := updates.Encode()
		statusList, _ := model.GetOnlineStatusModel().GetOnlineByUserId(pushData.GetPushUserId())
		switch pushData.GetPushType() {
		case zproto.SyncType_SYNC_TYPE_USER:
			for _, status := range statusList {
				seq = int32(model.GetSequenceModel().NextSeqId(base.Int64ToString(status.AuthKeyId)))
				if status.AuthKeyId == request.GetSenderAuthKeyId() {
					replySeq = seq
				}
				updates.SetDate(now)
				updates.SetSeq(seq)

				update := &zproto.PushUpdatesData{
					AuthKeyId:       status.AuthKeyId,
					SessionId:       status.SessionId,
					NetlibSessionId: status.NetlibSessionId,
					RawData:         updates.Encode(),
				}

				go s.withReadLock(func() {
					s.updates[status.ServerId] <- update
				})
			}
		case zproto.SyncType_SYNC_TYPE_USER_NOTME:
			for _, status := range statusList {
				seq = int32(model.GetSequenceModel().NextSeqId(base.Int64ToString(status.AuthKeyId)))
				if status.AuthKeyId == request.GetSenderAuthKeyId() {
					replySeq = seq
					continue
				}
				updates.SetDate(now)
				updates.SetSeq(seq)

				update := &zproto.PushUpdatesData{
					AuthKeyId:       status.AuthKeyId,
					SessionId:       status.SessionId,
					NetlibSessionId: status.NetlibSessionId,
					RawData:         updates.Encode(),
				}

				go s.withReadLock(func() {
					s.updates[status.ServerId] <- update
				})
			}
		case zproto.SyncType_SYNC_TYPE_AUTH_KEY:
			for _, status := range statusList {
				seq = int32(model.GetSequenceModel().NextSeqId(base.Int64ToString(status.AuthKeyId)))
				if status.AuthKeyId != request.GetSenderAuthKeyId() {
					replySeq = seq
				}
				updates.SetDate(now)
				updates.SetSeq(seq)

				update := &zproto.PushUpdatesData{
					AuthKeyId:       status.AuthKeyId,
					SessionId:       status.SessionId,
					NetlibSessionId: status.NetlibSessionId,
					RawData:         updates.Encode(),
				}

				go s.withReadLock(func() {
					s.updates[status.ServerId] <- update
				})
			}
		case zproto.SyncType_SYNC_TYPE_AUTH_KEY_USERNOTME:
			for _, status := range statusList {
				seq = int32(model.GetSequenceModel().NextSeqId(base.Int64ToString(status.AuthKeyId)))
				if status.AuthKeyId == request.GetSenderAuthKeyId() {
					replySeq = seq
					continue
				}
				updates.SetDate(now)
				updates.SetSeq(seq)

				update := &zproto.PushUpdatesData{
					AuthKeyId:       status.AuthKeyId,
					SessionId:       status.SessionId,
					NetlibSessionId: status.NetlibSessionId,
					RawData:         updates.Encode(),
				}

				go s.withReadLock(func() {
					s.updates[status.ServerId] <- update
				})
			}
		case zproto.SyncType_SYNC_TYPE_AUTH_KEY_USER:
			for _, status := range statusList {
				seq = int32(model.GetSequenceModel().NextSeqId(base.Int64ToString(status.AuthKeyId)))
				if status.AuthKeyId == request.GetSenderAuthKeyId() {
					replySeq = seq
				}
				updates.SetDate(now)
				updates.SetSeq(seq)

				update := &zproto.PushUpdatesData{
					AuthKeyId:       status.AuthKeyId,
					SessionId:       status.SessionId,
					NetlibSessionId: status.NetlibSessionId,
					RawData:         updates.Encode(),
				}

				go s.withReadLock(func() {
					s.updates[status.ServerId] <- update
				})
			}
		default:
		}

		//ss := make(map[int32][]*model.SessionStatus)
		//for _, status := range statusList {
		//	if _, ok := ss[status.ServerId]; !ok {
		//		ss[status.ServerId] = []*model.SessionStatus{}
		//	}
		//	// 会不会出问题？？
		//	ss[status.ServerId] = append(ss[status.ServerId], status)
		//}
		//
		//for k, ss3 := range ss {
		//	// glog.Infof("DeliveryUpdates: k: {%v}, v: {%v}", k, ss3)
		//	go s.withReadLock(func() {
		//		for _, ss4 := range ss3 {
		//			update := &zproto.PushUpdatesData{
		//				AuthKeyId:       ss4.AuthKeyId,
		//				SessionId:       ss4.SessionId,
		//				NetlibSessionId: ss4.NetlibSessionId,
		//				// RawData:         pushRawData,
		//			}
		//
		//
		//
		//			s.updates[k] <- update
		//
		//		}
		//	})
		//}

		// 			updates := pushData.GetPushData()
		// _ = updates
	}

	//seq = int32(model.GetSequenceModel().NextSeqId(base.Int64ToString(request.GetSenderAuthKeyId())))
	reply = &zproto.DeliveryRsp{
		Seq:  replySeq,
		Date: now,
	}

	glog.Infof("DeliveryPushUpdateShortMessage - reply: %s", logger.JsonDebugData(reply))
	return
}
