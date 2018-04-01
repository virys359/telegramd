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
	"github.com/nebulaim/telegramd/grpc_util"
	"github.com/nebulaim/telegramd/mtproto"
	"golang.org/x/net/context"
	"github.com/nebulaim/telegramd/biz/core/chat"
	update2 "github.com/nebulaim/telegramd/biz/core/update"
	"github.com/nebulaim/telegramd/biz_server/sync_client"
	"github.com/nebulaim/telegramd/biz/base"
	"github.com/nebulaim/telegramd/biz/core/message"
	"github.com/nebulaim/telegramd/biz/core/user"
)

func makeUpdatesByChatMessage(chatLogic *chat.ChatLogic, selfUserId int32, box *message.MessageBox) *update2.UpdatesLogic {
	updates := update2.NewUpdatesLogic(selfUserId)
	updateChatParticipants := &mtproto.TLUpdateChatParticipants{Data2: &mtproto.Update_Data{
		Participants: chatLogic.GetChatParticipants().To_ChatParticipants(),
	}}
	updates.AddUpdate(updateChatParticipants.To_Update())
	updates.AddUpdateNewMessage(box.Message)
	updates.AddUsers(user.GetUsersBySelfAndIDList(selfUserId, chatLogic.GetChatParticipantIdList()))
	return updates
}

func makeUpdatesByChatMessageAndMessageId(chatLogic *chat.ChatLogic, selfUserId int32, box *message.MessageBox) *update2.UpdatesLogic {
	updates := update2.NewUpdatesLogic(selfUserId)
	updates.AddUpdateMessageId(box.MessageId, box.RandomId)
	updateChatParticipants := &mtproto.TLUpdateChatParticipants{Data2: &mtproto.Update_Data{
		Participants: chatLogic.GetChatParticipants().To_ChatParticipants(),
	}}
	updates.AddUpdate(updateChatParticipants.To_Update())
	updates.AddUpdateNewMessage(box.Message)
	updates.AddUsers(user.GetUsersBySelfAndIDList(selfUserId, chatLogic.GetChatParticipantIdList()))
	return updates
}

// messages.addChatUser#f9a0aa09 chat_id:int user_id:InputUser fwd_limit:int = Updates;
func (s *MessagesServiceImpl) MessagesAddChatUser(ctx context.Context, request *mtproto.TLMessagesAddChatUser) (*mtproto.Updates, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("MessagesAddChatUser - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	var (
		addChatUserId int32
	)

	switch request.GetUserId().GetConstructor() {
	case mtproto.TLConstructor_CRC32_inputUserEmpty:
	case mtproto.TLConstructor_CRC32_inputUserSelf:
	case mtproto.TLConstructor_CRC32_inputUser:
		addChatUserId = request.GetUserId().GetData2().GetUserId()
	}

	chatLogic := chat.NewChatLogicById(request.GetChatId())
	chatLogic.AddChatUser(md.UserId, addChatUserId)

	peer := &base.PeerUtil{
		PeerType: base.PEER_CHAT,
		PeerId: chatLogic.Chat.GetId(),
	}

	addUserMessage := chatLogic.MakeAddUserMessage(md.UserId, addChatUserId)
	randomId := base.NextSnowflakeId()
	outbox := message.InsertMessageToOutbox(md.UserId, peer, randomId, addUserMessage)
	// TODO(@benqi): update dialog

	syncUpdates := makeUpdatesByChatMessage(chatLogic, md.UserId, outbox)
	state, _ := sync_client.GetSyncClient().SyncUpdatesData(md.AuthId, md.SessionId, md.UserId, syncUpdates.ToUpdates())

	replyUpdates := makeUpdatesByChatMessageAndMessageId(chatLogic, md.UserId, outbox)
	// replyUpdates.AddUpdateMessageId(outbox.MessageId, outbox.RandomId)
	replyUpdates.SetupState(state)
	// reply := syncUpdates.ToUpdates()

	inboxList, _ := outbox.InsertMessageToInbox(md.UserId, peer)
	for _, inbox := range inboxList {
		// TODO(@benqi): update dialog
		pushUpdates := makeUpdatesByChatMessage(chatLogic, inbox.UserId, inbox)
		sync_client.GetSyncClient().PushToUserUpdatesData(inbox.UserId, pushUpdates.ToUpdates())
	}

	glog.Infof("messages.createChat#9cb126e - reply: {%v}", replyUpdates)
	return replyUpdates.ToUpdates(), nil
}
