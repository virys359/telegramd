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
	"github.com/nebulaim/telegramd/biz/core"
	update2 "github.com/nebulaim/telegramd/biz/core/update"
	"github.com/nebulaim/telegramd/proto/mtproto"
	"github.com/nebulaim/telegramd/server/sync/sync_client"
	"golang.org/x/net/context"
)

// messages.editChatTitle#dc452855 chat_id:int title:string = Updates;
func (s *MessagesServiceImpl) MessagesEditChatTitle(ctx context.Context, request *mtproto.TLMessagesEditChatTitle) (*mtproto.Updates, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("messages.editChatTitle#dc452855 - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	chatLogic, err := s.ChatModel.NewChatLogicById(request.ChatId)
	if err != nil {
		glog.Error("messages.editChatTitle#dc452855 - error: ", err)
		return nil, err
	}

	peer := &base.PeerUtil{
		PeerType: base.PEER_CHAT,
		PeerId:   chatLogic.GetChatId(),
	}

	err = chatLogic.EditChatTitle(md.UserId, request.Title)
	if err != nil {
		glog.Error("messages.editChatTitle#dc452855 - error: ", err)
		return nil, err
	}

	chatEditMessage := chatLogic.MakeChatEditTitleMessage(md.UserId, request.Title)

	randomId := core.GetUUID()
	outbox := s.MessageModel.CreateMessageOutboxByNew(md.UserId, peer, randomId, chatEditMessage, func(messageId int32) {
		s.UserModel.CreateOrUpdateByOutbox(md.UserId, peer.PeerType, peer.PeerId, messageId, false, false)
	})

	syncUpdates := update2.NewUpdatesLogic(md.UserId)
	updateChatParticipants := &mtproto.TLUpdateChatParticipants{Data2: &mtproto.Update_Data{
		Participants: chatLogic.GetChatParticipants().To_ChatParticipants(),
	}}
	syncUpdates.AddUpdate(updateChatParticipants.To_Update())
	syncUpdates.AddUpdateNewMessage(chatEditMessage)
	syncUpdates.AddUsers(s.UserModel.GetUsersBySelfAndIDList(md.UserId, []int32{md.UserId}))
	syncUpdates.AddChat(chatLogic.ToChat(md.UserId))

	state, _ := sync_client.GetSyncClient().SyncUpdatesData(md.AuthId, md.SessionId, md.UserId, syncUpdates.ToUpdates())
	syncUpdates.PushTopUpdateMessageId(outbox.MessageId, outbox.RandomId)
	syncUpdates.SetupState(state)

	replyUpdates := syncUpdates.ToUpdates()

	inboxList, _ := outbox.InsertMessageToInbox(md.UserId, peer, func(inBoxUserId, messageId int32) {
		s.UserModel.CreateOrUpdateByInbox(inBoxUserId, base.PEER_CHAT, peer.PeerId, messageId, false)
	})

	for _, inbox := range inboxList {
		updates := update2.NewUpdatesLogic(md.UserId)
		updates.AddUpdate(updateChatParticipants.To_Update())
		updates.AddUpdateNewMessage(inbox.Message)
		updates.AddUsers(s.UserModel.GetUsersBySelfAndIDList(md.UserId, []int32{md.UserId}))
		updates.AddChat(chatLogic.ToChat(inbox.UserId))
		sync_client.GetSyncClient().PushToUserUpdatesData(inbox.UserId, updates.ToUpdates())
	}

	glog.Infof("messages.editChatTitle#dc452855 - reply: {%v}", replyUpdates)
	return replyUpdates, nil
}
