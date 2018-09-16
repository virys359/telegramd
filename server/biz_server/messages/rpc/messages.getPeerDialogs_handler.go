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
	"github.com/nebulaim/telegramd/baselib/grpc_util"
	"github.com/nebulaim/telegramd/baselib/logger"
	"github.com/nebulaim/telegramd/biz/base"
	"github.com/nebulaim/telegramd/proto/mtproto"
	"golang.org/x/net/context"
	// "github.com/nebulaim/telegramd/server/sync/sync_client"
	"github.com/nebulaim/telegramd/server/sync/sync_client"
)

// messages.getPeerDialogs#2d9776b9 peers:Vector<InputPeer> = messages.PeerDialogs;
func (s *MessagesServiceImpl) MessagesGetPeerDialogs(ctx context.Context, request *mtproto.TLMessagesGetPeerDialogs) (*mtproto.Messages_PeerDialogs, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("messages.getPeerDialogs#2d9776b9 - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	peerDialogs := mtproto.NewTLMessagesPeerDialogs()

	messageIdList := []int32{}
	userIdList := []int32{md.UserId}
	chatIdList := []int32{}

	// dialogs := s.UserModel.GetPeersDialogs(md.UserId, request.GetPeers())
	dialogs := s.UserModel.GetPeersDialogs(md.UserId, nil)

	for _, dialog2 := range dialogs {
		dialog := dialog2.To_Dialog()
		messageIdList = append(messageIdList, dialog.GetTopMessage())
		peer := base.FromPeer(dialog.GetPeer())
		// TODO(@benqi): 先假设只有PEER_USER
		if peer.PeerType == base.PEER_USER {
			userIdList = append(userIdList, peer.PeerId)
		} else if peer.PeerType == base.PEER_SELF {
			userIdList = append(userIdList, md.UserId)
		} else if peer.PeerType == base.PEER_CHAT {
			chatIdList = append(chatIdList, peer.PeerId)
		} else if peer.PeerType == base.PEER_CHANNEL {

		}
		peerDialogs.Data2.Dialogs = append(peerDialogs.Data2.Dialogs, dialog.To_Dialog())
	}

	glog.Infof("messageIdList - %v", messageIdList)
	if len(messageIdList) > 0 {
		peerDialogs.SetMessages(s.MessageModel.GetUserMessagesByMessageIdList(md.UserId, messageIdList))
	}

	users := s.UserModel.GetUsersBySelfAndIDList(md.UserId, userIdList)
	peerDialogs.SetUsers(users)

	if len(chatIdList) > 0 {
		peerDialogs.Data2.Chats = s.ChatModel.GetChatListBySelfAndIDList(md.UserId, chatIdList)
	}

	state, _ := sync_client.GetSyncClient().SyncGetState(md.AuthId, md.UserId)
	peerDialogs.SetState(state)

	glog.Infof("messages.getPeerDialogs#2d9776b9 - reply: %s", logger.JsonDebugData(peerDialogs))
	return peerDialogs.To_Messages_PeerDialogs(), nil
}
