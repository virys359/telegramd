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

// messages.addChatUser#f9a0aa09 chat_id:int user_id:InputUser fwd_limit:int = Updates;
func (s *MessagesServiceImpl) MessagesAddChatUser(ctx context.Context, request *mtproto.TLMessagesAddChatUser) (*mtproto.Updates, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("messages.addChatUser#f9a0aa09 - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	var (
		err           error
		addChatUserId int32
	)

	if request.GetUserId().GetConstructor() == mtproto.TLConstructor_CRC32_inputUserEmpty {
		err = mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_BAD_REQUEST)
		glog.Error("messages.sendMessage#fa88427a - invalid peer", err)
		return nil, err
	}

	switch request.GetUserId().GetConstructor() {
	case mtproto.TLConstructor_CRC32_inputUserEmpty:
	case mtproto.TLConstructor_CRC32_inputUserSelf:
		addChatUserId = md.UserId
	case mtproto.TLConstructor_CRC32_inputUser:
		addChatUserId = request.GetUserId().GetData2().GetUserId()
	}

	chatLogic, _ := s.ChatModel.NewChatLogicById(request.GetChatId())
	chatLogic.AddChatUser(md.UserId, addChatUserId)

	peer := &base.PeerUtil{
		PeerType: base.PEER_CHAT,
		PeerId:   chatLogic.GetChatId(),
	}

	addUserMessage := chatLogic.MakeAddUserMessage(md.UserId, addChatUserId)
	randomId := core.GetUUID()
	outbox := s.MessageModel.CreateMessageOutboxByNew(md.UserId, peer, randomId, addUserMessage, func(messageId int32) {
		s.UserModel.CreateOrUpdateByOutbox(md.UserId, peer.PeerType, peer.PeerId, messageId, false, false)
	})
	_ = outbox

	syncUpdates := update2.NewUpdatesLogic(md.UserId)
	updateChatParticipants := &mtproto.TLUpdateChatParticipants{Data2: &mtproto.Update_Data{
		Participants: chatLogic.GetChatParticipants().To_ChatParticipants(),
	}}
	syncUpdates.AddUpdate(updateChatParticipants.To_Update())
	syncUpdates.AddUpdateNewMessage(addUserMessage)
	syncUpdates.AddUsers(s.UserModel.GetUsersBySelfAndIDList(md.UserId, chatLogic.GetChatParticipantIdList()))
	syncUpdates.AddChat(chatLogic.ToChat(md.UserId))

	state, _ := sync_client.GetSyncClient().SyncUpdatesData(md.AuthId, md.SessionId, md.UserId, syncUpdates.ToUpdates())
	syncUpdates.AddUpdateMessageId(outbox.MessageId, outbox.RandomId)
	syncUpdates.SetupState(state)
	replyUpdates := syncUpdates.ToUpdates()

	inboxList, _ := outbox.InsertMessageToInbox(md.UserId, peer, func(inBoxUserId, messageId int32) {
		s.UserModel.CreateOrUpdateByInbox(inBoxUserId, base.PEER_CHAT, peer.PeerId, messageId, false)
	})

	for _, inbox := range inboxList {
		updates := update2.NewUpdatesLogic(md.UserId)
		updates.AddUpdate(updateChatParticipants.To_Update())
		updates.AddUpdateNewMessage(inbox.Message)
		updates.AddUsers(s.UserModel.GetUsersBySelfAndIDList(md.UserId, chatLogic.GetChatParticipantIdList()))
		updates.AddChat(chatLogic.ToChat(inbox.UserId))
		sync_client.GetSyncClient().PushToUserUpdatesData(inbox.UserId, updates.ToUpdates())
	}

	// TODO(@benqi): Add后需要将聊天历史消息导入

	glog.Infof("messages.addChatUser#f9a0aa09 - reply: {%v}", replyUpdates)
	return replyUpdates, nil
}
