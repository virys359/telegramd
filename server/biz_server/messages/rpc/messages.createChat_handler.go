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

// messages.createChat#9cb126e users:Vector<InputUser> title:string = Updates;
func (s *MessagesServiceImpl) MessagesCreateChat(ctx context.Context, request *mtproto.TLMessagesCreateChat) (*mtproto.Updates, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("messages.createChat#9cb126e - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	// logic.NewChatLogicByCreateChat()
	//// TODO(@benqi): Impl MessagesCreateChat logic
	//randomId := md.ClientMsgId

	chatUserIdList := make([]int32, 0, len(request.GetUsers()))
	// chatUserIdList = append(chatUserIdList, md.UserId)
	for _, u := range request.GetUsers() {
		switch u.GetConstructor() {
		case mtproto.TLConstructor_CRC32_inputUser:
			chatUserIdList = append(chatUserIdList, u.GetData2().GetUserId())
		default:
			// TODO(@benqi): chatUser不能是inputUser和inputUserSelf
			err := mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_BAD_REQUEST)
			glog.Error("messages.createChat#9cb126e - error: ", err, "; InputPeer invalid")
			return nil, err
		}
	}

	chat := s.ChatModel.NewChatLogicByCreateChat(md.UserId, chatUserIdList, request.GetTitle())

	peer := &base.PeerUtil{
		PeerType: base.PEER_CHAT,
		PeerId:   chat.GetChatId(),
	}

	createChatMessage := chat.MakeCreateChatMessage(md.UserId)
	randomId := core.GetUUID()
	outbox := s.MessageModel.CreateMessageOutboxByNew(md.UserId, peer, randomId, createChatMessage, func(messageId int32) {
		s.UserModel.CreateOrUpdateByOutbox(md.UserId, peer.PeerType, peer.PeerId, messageId, false, false)
	})

	syncUpdates := update2.NewUpdatesLogic(md.UserId)
	updateChatParticipants := &mtproto.TLUpdateChatParticipants{Data2: &mtproto.Update_Data{
		Participants: chat.GetChatParticipants().To_ChatParticipants(),
	}}
	syncUpdates.AddUpdate(updateChatParticipants.To_Update())
	syncUpdates.AddUpdateNewMessage(createChatMessage)
	syncUpdates.AddUsers(s.UserModel.GetUsersBySelfAndIDList(md.UserId, chat.GetChatParticipantIdList()))
	syncUpdates.AddChat(chat.ToChat(md.UserId))

	state, _ := sync_client.GetSyncClient().SyncUpdatesData(md.AuthId, md.SessionId, md.UserId, syncUpdates.ToUpdates())
	syncUpdates.AddUpdateMessageId(outbox.MessageId, outbox.RandomId)
	syncUpdates.SetupState(state)
	reply := syncUpdates.ToUpdates()

	inboxList, _ := outbox.InsertMessageToInbox(md.UserId, peer, func(inBoxUserId, messageId int32) {
		s.UserModel.CreateOrUpdateByInbox(inBoxUserId, base.PEER_CHAT, peer.PeerId, messageId, false)
	})
	for _, inbox := range inboxList {
		updates := update2.NewUpdatesLogic(md.UserId)
		updateChatParticipants := &mtproto.TLUpdateChatParticipants{Data2: &mtproto.Update_Data{
			Participants: chat.GetChatParticipants().To_ChatParticipants(),
		}}
		updates.AddUpdate(updateChatParticipants.To_Update())
		updates.AddUpdateNewMessage(inbox.Message)
		updates.AddUsers(s.UserModel.GetUsersBySelfAndIDList(md.UserId, chat.GetChatParticipantIdList()))
		updates.AddChat(chat.ToChat(inbox.UserId))
		sync_client.GetSyncClient().PushToUserUpdatesData(inbox.UserId, updates.ToUpdates())
	}

	glog.Infof("messages.createChat#9cb126e - reply: {%v}", reply)
	return reply, nil
}
