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

// messages.deleteChatUser#e0611f16 chat_id:int user_id:InputUser = Updates;
func (s *MessagesServiceImpl) MessagesDeleteChatUser(ctx context.Context, request *mtproto.TLMessagesDeleteChatUser) (*mtproto.Updates, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("messages.deleteChatUser#e0611f16 - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	var (
		err              error
		deleteChatUserId int32
	)

	if request.GetUserId().GetConstructor() == mtproto.TLConstructor_CRC32_inputUserEmpty {
		err = mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_BAD_REQUEST)
		glog.Error("messages.deleteChatUser#e0611f16 - invalid peer", err)
		return nil, err
	}

	switch request.GetUserId().GetConstructor() {
	case mtproto.TLConstructor_CRC32_inputUserEmpty:
	case mtproto.TLConstructor_CRC32_inputUserSelf:
		deleteChatUserId = md.UserId
	case mtproto.TLConstructor_CRC32_inputUser:
		deleteChatUserId = request.GetUserId().GetData2().GetUserId()
	}

	chatLogic, _ := s.ChatModel.NewChatLogicById(request.GetChatId())

	peer := &base.PeerUtil{
		PeerType: base.PEER_CHAT,
		PeerId:   chatLogic.GetChatId(),
	}

	err = chatLogic.CheckDeleteChatUser(md.UserId, deleteChatUserId)
	if err != nil {
		glog.Error("messages.deleteChatUser#e0611f16 - invalid peer", err)
		return nil, err
	}

	// make delete user message
	deleteUserMessage := chatLogic.MakeDeleteUserMessage(md.UserId, deleteChatUserId)
	randomId := core.GetUUID()
	outbox := s.MessageModel.CreateMessageOutboxByNew(md.UserId, peer, randomId, deleteUserMessage, func(messageId int32) {
		s.UserModel.CreateOrUpdateByOutbox(md.UserId, peer.PeerType, peer.PeerId, messageId, false, false)
	})
	inboxList, _ := outbox.InsertMessageToInbox(md.UserId, peer, func(inBoxUserId, messageId int32) {
		s.UserModel.CreateOrUpdateByInbox(inBoxUserId, base.PEER_CHAT, peer.PeerId, messageId, false)
	})

	// update
	_ = chatLogic.DeleteChatUser(md.UserId, deleteChatUserId)
	syncUpdates := update2.NewUpdatesLogic(md.UserId)
	updateChatParticipants := &mtproto.TLUpdateChatParticipants{Data2: &mtproto.Update_Data{
		Participants: chatLogic.GetChatParticipants().To_ChatParticipants(),
	}}
	syncUpdates.AddUpdate(updateChatParticipants.To_Update())
	syncUpdates.AddUpdateNewMessage(deleteUserMessage)
	syncUpdates.AddUsers(s.UserModel.GetUsersBySelfAndIDList(md.UserId, chatLogic.GetChatParticipantIdList()))
	syncUpdates.AddChat(chatLogic.ToChat(md.UserId))

	state, _ := sync_client.GetSyncClient().SyncUpdatesData(md.AuthId, md.SessionId, md.UserId, syncUpdates.ToUpdates())
	syncUpdates.AddUpdateMessageId(outbox.MessageId, outbox.RandomId)
	syncUpdates.SetupState(state)
	replyUpdates := syncUpdates.ToUpdates()

	for _, inbox := range inboxList {
		updates := update2.NewUpdatesLogic(md.UserId)
		updates.AddUpdate(updateChatParticipants.To_Update())
		updates.AddUpdateNewMessage(inbox.Message)
		updates.AddUsers(s.UserModel.GetUsersBySelfAndIDList(md.UserId, chatLogic.GetChatParticipantIdList()))
		updates.AddChat(chatLogic.ToChat(inbox.UserId))
		sync_client.GetSyncClient().PushToUserUpdatesData(inbox.UserId, updates.ToUpdates())
	}

	glog.Infof("messages.deleteChatUser#e0611f16 - reply: {%v}", replyUpdates)
	return replyUpdates, nil
}
