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
	// "github.com/nebulaim/telegramd/biz_server/sync_client"
	"github.com/nebulaim/telegramd/biz/base"
	"github.com/nebulaim/telegramd/biz/core/message"
	// "github.com/nebulaim/telegramd/biz/core/user"
	"github.com/nebulaim/telegramd/biz/core/user"
	"github.com/nebulaim/telegramd/biz_server/sync_client"
)

// messages.deleteChatUser#e0611f16 chat_id:int user_id:InputUser = Updates;
func (s *MessagesServiceImpl) MessagesDeleteChatUser(ctx context.Context, request *mtproto.TLMessagesDeleteChatUser) (*mtproto.Updates, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("messages.deleteChatUser#e0611f16 - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	var (
		err error
		deleteChatUserId int32
	)

	if request.GetUserId().GetConstructor() == mtproto.TLConstructor_CRC32_inputUserEmpty {
		err = mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_BAD_REQUEST)
		glog.Error("messages.sendMessage#fa88427a - invalid peer", err)
		return nil, err
	}

	switch request.GetUserId().GetConstructor() {
	case mtproto.TLConstructor_CRC32_inputUserEmpty:
	case mtproto.TLConstructor_CRC32_inputUserSelf:
		deleteChatUserId = md.UserId
	case mtproto.TLConstructor_CRC32_inputUser:
		deleteChatUserId = request.GetUserId().GetData2().GetUserId()
	}

	chatLogic, _ := chat.NewChatLogicById(request.GetChatId())

	peer := &base.PeerUtil{
		PeerType: base.PEER_CHAT,
		PeerId: chatLogic.GetChatId(),
	}

	// make delete user message
	deleteUserMessage := chatLogic.MakeDeleteUserMessage(md.UserId, deleteChatUserId)
	randomId := base.NextSnowflakeId()
	outbox := message.CreateMessageOutboxByNew(md.UserId, peer, randomId, deleteUserMessage, func(messageId int32) {
		user.CreateOrUpdateByOutbox(md.UserId, peer.PeerType, peer.PeerId, messageId, false, false)
	})
	inboxList, _ := outbox.InsertMessageToInbox(md.UserId, peer, func(inBoxUserId, messageId int32) {
		user.CreateOrUpdateByInbox(inBoxUserId, base.PEER_CHAT, peer.PeerId, messageId, false)
	})

	// update
	chatLogic.DeleteChatUser(deleteChatUserId)
	syncUpdates := update2.NewUpdatesLogic(md.UserId)
	updateChatParticipants := &mtproto.TLUpdateChatParticipants{Data2: &mtproto.Update_Data{
		Participants: chatLogic.GetChatParticipants().To_ChatParticipants(),
	}}
	syncUpdates.AddUpdate(updateChatParticipants.To_Update())
	syncUpdates.AddUpdateNewMessage(deleteUserMessage)
	syncUpdates.AddUsers(user.GetUsersBySelfAndIDList(md.UserId, chatLogic.GetChatParticipantIdList()))
	syncUpdates.AddChat(chatLogic.ToChat(md.UserId))

	state, _ := sync_client.GetSyncClient().SyncUpdatesData(md.AuthId, md.SessionId, md.UserId, syncUpdates.ToUpdates())
	syncUpdates.AddUpdateMessageId(outbox.MessageId, outbox.RandomId)
	syncUpdates.SetupState(state)
	replyUpdates := syncUpdates.ToUpdates()

	for _, inbox := range inboxList {
		updates := update2.NewUpdatesLogic(md.UserId)
		updates.AddUpdate(updateChatParticipants.To_Update())
		updates.AddUpdateNewMessage(inbox.Message)
		updates.AddUsers(user.GetUsersBySelfAndIDList(md.UserId, chatLogic.GetChatParticipantIdList()))
		updates.AddChat(chatLogic.ToChat(inbox.UserId))
		sync_client.GetSyncClient().PushToUserUpdatesData(inbox.UserId, updates.ToUpdates())
	}

	// TODO(@benqi): Add后需要将聊天历史消息导入

	glog.Infof("messages.deleteChatUser#e0611f16 - reply: {%v}", replyUpdates)
	return replyUpdates, nil

	// md := grpc_util.RpcMetadataFromIncoming(ctx)
	// glog.Infof("messages.deleteChatUser#e0611f16 - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	//// TODO(@benqi): Impl MessagesDeleteChatUser logic
	//chat := model.GetChatModel().GetChat(request.ChatId)
	//participants := model.GetChatModel().GetChatParticipants(chat.Id)
	//var deleteChatUserId int32
	//// peer := base.fr(request.UserId)
	//switch request.UserId.Payload.(type) {
	//case *mtproto.InputUser_InputUser:
	//	deleteChatUserId = request.GetUserId().GetInputUser().GetUserId()
	//case *mtproto.InputUser_InputUserSelf:
	//	deleteChatUserId = md.UserId
	//case *mtproto.InputUser_InputUserEmpty:
	//	panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_BAD_REQUEST), "InputPeer invalid"))
	//}
	//dao.GetChatUsersDAO(dao.DB_MASTER).DeleteChatUser(chat.Id, deleteChatUserId)
	//dao.GetChatsDAO(dao.DB_MASTER).UpdateParticipantCount(chat.ParticipantsCount-1, chat.Version+1, chat.Id)
	//chat.ParticipantsCount -= 1
	//chat.Version += 1
	//participants.Version = chat.Version
	//chatUserIdList := make([]int32, 0, chat.ParticipantsCount)
	//participantUsers := participants.GetParticipants()
	//var foundId int = 0
	//for i, participant := range participantUsers {
	//	switch participant.Payload.(type) {
	//	case *mtproto.ChatParticipant_ChatParticipant:
	//		if deleteChatUserId == participant.GetChatParticipant().GetUserId() {
	//			foundId = i
	//		}
	//		chatUserIdList = append(chatUserIdList, participant.GetChatParticipant().GetUserId())
	//	case *mtproto.ChatParticipant_ChatParticipantCreator:
	//		if deleteChatUserId == participant.GetChatParticipantCreator().GetUserId() {
	//			foundId = i
	//		}
	//		chatUserIdList = append(chatUserIdList, participant.GetChatParticipantCreator().GetUserId())
	//	case *mtproto.ChatParticipant_ChatParticipantAdmin:
	//		if deleteChatUserId == participant.GetChatParticipantAdmin().GetUserId() {
	//			foundId = i
	//		}
	//		chatUserIdList = append(chatUserIdList, participant.GetChatParticipantAdmin().GetUserId())
	//	}
	//}
	//if foundId != 0 {
	//	participantUsers = append(participantUsers[:foundId], participantUsers[foundId+1:]...)
	//	participants.Participants = participantUsers
	//}
	//peer := &base.PeerUtil{}
	//peer.PeerType = base.PEER_CHAT
	//peer.PeerId = chat.Id
	//messageService := &mtproto.TLMessageService{}
	//messageService.Out = true
	//messageService.Date = chat.Date
	//messageService.FromId = md.UserId
	//messageService.ToId = peer.ToPeer()
	//// mtproto.MakePeer(&mtproto.TLPeerChat{chat.Id})
	//action := &mtproto.TLMessageActionChatDeleteUser{}
	//action.UserId = deleteChatUserId
	//messageService.Action = action.ToMessageAction()
	//messageServiceId := model.GetMessageModel().CreateHistoryMessageService(md.UserId, peer, md.ClientMsgId, messageService)
	//messageService.Id = int32(messageServiceId)
	//users := model.GetUserModel().GetUserList(chatUserIdList)
	//updateUsers := make([]*mtproto.User, 0, len(users))
	//for _, u := range users {
	//	u.Self = true
	//	updates := &mtproto.TLUpdates{}
	//	// 2. MessageBoxes
	//	pts := model.GetMessageModel().CreateMessageBoxes(u.Id, md.UserId, peer.PeerType, peer.PeerId, false, messageServiceId)
	//	// 3. dialog
	//	model.GetDialogModel().CreateOrUpdateByLastMessage(u.Id, peer.PeerType, peer.PeerId, messageServiceId, false)
	//	if u.GetId() == md.UserId {
	//		updateMessageID := &mtproto.TLUpdateMessageID{}
	//		updateMessageID.Id = int32(messageServiceId)
	//		updateMessageID.RandomId = md.ClientMsgId
	//		updates.Updates = append(updates.Updates, updateMessageID.ToUpdate())
	//		updates.Seq = 0
	//	} else {
	//		// TODO(@benqi): seq++
	//		updates.Seq = 0
	//	}
	//	updateChatParticipants := &mtproto.TLUpdateChatParticipants{}
	//	updateChatParticipants.Participants = participants.ToChatParticipants()
	//	updates.Updates = append(updates.Updates, updateChatParticipants.ToUpdate())
	//	updateNewMessage := &mtproto.TLUpdateNewMessage{}
	//	updateNewMessage.Pts = pts
	//	updateNewMessage.PtsCount = 1
	//	updateNewMessage.Message = messageService.ToMessage()
	//	updates.Updates = append(updates.Updates, updateNewMessage.ToUpdate())
	//	updates.Users = updateUsers
	//	updates.Date = chat.Date
	//	if u.Id == md.UserId {
	//		// TODO(@benqi): Delete me
	//		updates.Chats = append(updates.Chats, chat.ToChat())
	//		reply = updates.ToUpdates()
	//		delivery2.GetDeliveryInstance().DeliveryUpdatesNotMe(
	//			md.AuthId,
	//			md.SessionId,
	//			md.NetlibSessionId,
	//			[]int32{u.Id},
	//			updates.ToUpdates().Encode())
	//	} else if u.Id == deleteChatUserId {
	//		// deleteChatUserId，chat设置成forbidden
	//		chat3 := &mtproto.TLChatForbidden{}
	//		chat3.Id = chat.Id
	//		chat3.Title = chat.Title
	//		updates.Chats = append(updates.Chats, chat3.ToChat())
	//		delivery2.GetDeliveryInstance().DeliveryUpdates(
	//			md.AuthId,
	//			md.SessionId,
	//			md.NetlibSessionId,
	//			[]int32{u.Id},
	//			updates.ToUpdates().Encode())
	//	} else {
	//		updates.Chats = append(updates.Chats, chat.ToChat())
	//		delivery2.GetDeliveryInstance().DeliveryUpdates(
	//			md.AuthId,
	//			md.SessionId,
	//			md.NetlibSessionId,
	//			[]int32{u.Id},
	//			updates.ToUpdates().Encode())
	//	}
	//	u.Self = false
	//}
	//for _, u := range users {
	//	// updates := &mtproto.TLUpdates{}
	//	if u.Id == md.UserId {
	//		u.Self = true
	//	}
	//	updateUsers = append(updateUsers, u.ToUser())
	//}
	//glog.Infof("messages.deleteChatUser#e0611f16 - reply: {%v}", reply)
	//return
	// return nil, fmt.Errorf("Not impl MessagesDeleteChatUser")
}
