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
	"github.com/golang/glog"
	"github.com/nebulaim/telegramd/baselib/logger"
	"github.com/nebulaim/telegramd/baselib/grpc_util"
	"github.com/nebulaim/telegramd/proto/mtproto"
	"golang.org/x/net/context"
	"github.com/nebulaim/telegramd/biz/base"
	"github.com/nebulaim/telegramd/biz/dal/dao"
	"github.com/nebulaim/telegramd/server/sync/sync_client"
)

// messages.readHistory#b04f2510 peer:InputPeer max_id:int offset:int = messages.AffectedHistory;
func (s *MessagesServiceImpl) MessagesReadHistoryLayer2(ctx context.Context, request *mtproto.TLMessagesReadHistoryLayer2) (*mtproto.Messages_AffectedHistory, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("messages.readHistory#b04f2510 - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	peer := base.FromInputPeer(request.GetPeer())
	if peer.PeerType == base.PEER_SELF {
		// TODO(@benqi): 太土！
		peer.PeerType = base.PEER_USER
		peer.PeerId = md.UserId
	}

	// 消息已读逻辑
	// 1. inbox，设置unread_count为0以及read_inbox_max_id
	dao.GetUserDialogsDAO(dao.DB_MASTER).UpdateUnreadByPeer(request.GetMaxId(), md.UserId, int8(peer.PeerType), peer.PeerId)

	updateReadHistoryInbox := mtproto.NewTLUpdateReadHistoryInbox()
	updateReadHistoryInbox.SetPeer(peer.ToPeer())
	updateReadHistoryInbox.SetMaxId(request.MaxId)

	_, err := sync_client.GetSyncClient().SyncOneUpdateData3(md.ServerId, md.AuthId, md.SessionId, md.UserId, md.ClientMsgId, updateReadHistoryInbox.To_Update())
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	//affected := mtproto.NewTLMessagesAffectedMessages()
	//affected.SetPts(int32(state.Pts))
	//affected.SetPtsCount(state.PtsCount)

	// 2. outbox, 设置read_outbox_max_id
	if peer.PeerType == base.PEER_USER {
		outboxDO := dao.GetUserDialogsDAO(dao.DB_SLAVE).SelectByPeer(peer.PeerId, int8(peer.PeerType), md.UserId)
		dao.GetUserDialogsDAO(dao.DB_MASTER).UpdateReadOutboxMaxIdByPeer(outboxDO.TopMessage, peer.PeerId, int8(peer.PeerType), md.UserId)

		updateReadHistoryOutbox := mtproto.NewTLUpdateReadHistoryOutbox()
		outboxPeer := &mtproto.TLPeerUser{Data2: &mtproto.Peer_Data{
			UserId: md.UserId,
		}}
		updateReadHistoryOutbox.SetPeer(outboxPeer.To_Peer())
		updateReadHistoryOutbox.SetMaxId(outboxDO.TopMessage)

		sync_client.GetSyncClient().PushToUserOneUpdateData(peer.PeerId, updateReadHistoryOutbox.To_Update())
	} else {
		doList := dao.GetChatParticipantsDAO(dao.DB_SLAVE).SelectByChatId(peer.PeerId)
		for _, do := range doList {
			if do.UserId == md.UserId {
				continue
			}
			outboxDO := dao.GetUserDialogsDAO(dao.DB_SLAVE).SelectByPeer(do.UserId, int8(peer.PeerType), peer.PeerId)
			dao.GetUserDialogsDAO(dao.DB_MASTER).UpdateReadOutboxMaxIdByPeer(outboxDO.TopMessage, do.UserId, int8(peer.PeerType), peer.PeerId)

			updateReadHistoryOutbox := mtproto.NewTLUpdateReadHistoryOutbox()
			outboxPeer := &mtproto.TLPeerChat{Data2: &mtproto.Peer_Data{
				ChatId: peer.PeerId,
			}}
			updateReadHistoryOutbox.SetPeer(outboxPeer.To_Peer())
			updateReadHistoryOutbox.SetMaxId(outboxDO.TopMessage)

			sync_client.GetSyncClient().PushToUserOneUpdateData(do.UserId, updateReadHistoryOutbox.To_Update())
		}
	}

	err = mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_NOTRETURN_CLIENT)
	glog.Infof("messages.readHistory#b04f2510 - reply: {%v}", err)
	return nil, err
	// affected.To_Messages_AffectedMessages(), nil
}
