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
	// "time"
	"github.com/nebulaim/telegramd/biz/core/chat"
	"github.com/nebulaim/telegramd/biz/core/message"
	"github.com/nebulaim/telegramd/biz/base"
	"github.com/nebulaim/telegramd/biz/core/updates"
	"github.com/nebulaim/telegramd/biz_server/sync_client"
	"github.com/nebulaim/telegramd/biz/core/user"
)


type createChatUpdates mtproto.TLUpdates

func makeCreateChatUpdates() *createChatUpdates {
	return nil
}

//func (this *createChatUpdates) addChat() {
//}

func (this *createChatUpdates) addChat() {
}

// sync: updateMessageId
// createChatMessageService
// participants
// users
// chats
//func makeCreateChatUpdates(selfUserId int32, participants *mtproto.ChatParticipants, message *mtproto.Message) *mtproto.TLUpdates {
//	userIdList, _, _ := model.PickAllIDListByMessages([]*mtproto.Message{message})
//	userList := model.GetUserModel().GetUsersBySelfAndIDList(selfUserId, userIdList)
//	updateNew := &mtproto.TLUpdateEditMessage{Data2: &mtproto.Update_Data{
//		Message_1: message,
//	}}
//
//	return &mtproto.TLUpdates{Data2: &mtproto.Updates_Data{
//		Updates: []*mtproto.Update{updateNew.To_Update()},
//		Users:   userList,
//		Date:    int32(time.Now().Unix()),
//		Seq:     0,
//	}}
//}

//type messagesCreateChatHandler struct {
//	md      *grpc_util.RpcMetadata
//	request *mtproto.TLMessagesCreateChat
//	result  *mtproto.Updates
//	err     error
//}
//
//func (h *messagesCreateChatHandler) onPreInit() bool {
//	return true
//}
//
//func (h *messagesCreateChatHandler) onExecute() bool {
//	return true
//}
//
//func (h *messagesCreateChatHandler) onPost() bool {
//	return true
//}

// messages.createChat#9cb126e users:Vector<InputUser> title:string = Updates;
func (s *MessagesServiceImpl) MessagesCreateChat(ctx context.Context, request *mtproto.TLMessagesCreateChat) (*mtproto.Updates, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("messages.createChat#9cb126e - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	// logic.NewChatLogicByCreateChat()
	//// TODO(@benqi): Impl MessagesCreateChat logic
	//randomId := md.ClientMsgId

	chatUserIdList := make([]int32, 0, len(request.GetUsers()))
	chatUserIdList = append(chatUserIdList, md.UserId)
	for _, u := range request.GetUsers() {
		switch u.GetConstructor() {
		case mtproto.TLConstructor_CRC32_inputUser:
			chatUserIdList = append(chatUserIdList, u.GetData2().GetUserId())
		default:
			// TODO(@benqi): chatUser不能是inputUser和inputUserSelf
			return nil, mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_BAD_REQUEST), "InputPeer invalid")
		}
	}

	chat := chat.NewChatLogicByCreateChat(md.UserId, chatUserIdList, request.GetTitle())

	peer := &base.PeerUtil{
		PeerType: base.PEER_CHAT,
		PeerId: chat.Chat.GetId(),
	}

	createChatMessage := chat.MakeCreateChatMessage(md.UserId)
	randomId := base.NextSnowflakeId()
	outbox := message.InsertMessageToOutbox(md.UserId, peer, randomId, createChatMessage)

	syncUpdates := updates.NewUpdatesLogic(md.UserId)
	updateChatParticipants := &mtproto.TLUpdateChatParticipants{Data2: &mtproto.Update_Data{
		Participants: chat.GetChatParticipants().To_ChatParticipants(),
	}}
	syncUpdates.AddUpdate(updateChatParticipants.To_Update())
	syncUpdates.AddUpdateNewMessage(createChatMessage)
	syncUpdates.AddUsers(user.GetUsersBySelfAndIDList(md.UserId, chat.GetChatParticipantIdList()))
	syncUpdates.AddChat(chat.Chat.To_Chat())

	state, _ := sync_client.GetSyncClient().SyncUpdatesData(md.AuthId, md.SessionId, md.UserId, syncUpdates.ToUpdates())
	syncUpdates.AddUpdateMessageId(outbox.MessageId, outbox.RandomId)
	syncUpdates.SetupState(state)
	reply := syncUpdates.ToUpdates()

	inboxList, _ := outbox.InsertMessageToInbox(md.UserId, peer)
	for _, inbox := range inboxList {
		updates := updates.NewUpdatesLogic(md.UserId)
		updateChatParticipants := &mtproto.TLUpdateChatParticipants{Data2: &mtproto.Update_Data{
			Participants: chat.GetChatParticipants().To_ChatParticipants(),
		}}
		updates.AddUpdate(updateChatParticipants.To_Update())
		updates.AddUpdateNewMessage(inbox.Message)
		updates.AddUsers(user.GetUsersBySelfAndIDList(md.UserId, chat.GetChatParticipantIdList()))
		updates.AddChat(chat.Chat.To_Chat())
		sync_client.GetSyncClient().PushToUserUpdatesData(inbox.UserId, updates.ToUpdates())
	}

	glog.Infof("messages.createChat#9cb126e - reply: {%v}", reply)
	return reply, nil

/*
	// logic2.InsertMessageToOutbox()
	// if err != nil {
	// 	return nil, err
	// }

	//createChatMessageList := chat.MakeCreateChatMessageList()
	//model.GetMessageModel().SendMessageToOutbox(md.UserId, peer, 0, createChatMessageList[0].Message)
	//
	//chat, _ := model.GetChatModel().CreateChat(md.UserId, request.Title, chatUserIdList, md.ClientMsgId)
	//
	//peer := &base.PeerUtil{
	//	PeerType: base.PEER_CHAT,
	//	PeerId:   chat.GetId(),
	//}

	// mtproto.MakePeer(&mtproto.TLPeerChat{chat.Id})
	action := &mtproto.TLMessageActionChatCreate{Data2: &mtproto.MessageAction_Data{
		Title: request.GetTitle(),
		Users: chatUserIdList,
	}}

	messageService := &mtproto.TLMessageService{Data2: &mtproto.Message_Data{
		Out: true,
		Date: chat.GetDate(),
		FromId: md.UserId,
		ToId: peer.ToPeer(),
		Action: action.To_MessageAction(),
	}}

	createChatMessageId, _ := model.GetMessageModel().SendMessageToOutbox(md.UserId, peer, 0, messageService.To_Message())
	messageService.SetId(createChatMessageId)

	//messageService.Out = true
	//messageService.Date = chat.Date
	//messageService.FromId = md.UserId
	//messageService.ToId = peer.ToPeer()
	//
	//messageServiceId := model.GetMessageModel().CreateHistoryMessageService(md.UserId, peer, md.ClientMsgId, messageService)
	//messageService.Id = int32(messageServiceId)
	_ = model.GetUserModel().GetUsersBySelfAndIDList(md.UserId, chatUserIdList)

	// sync
	// update dialog

	// 遍历所有参与者，创建服务消息并推送

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
	//		updateMessageID.RandomId = randomId
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
	//	updates.Chats = append(updates.Chats, chat.ToChat())
	//	updates.Date = chat.Date
	//	if u.Id == md.UserId {
	//		reply = updates.ToUpdates()
	//		delivery2.GetDeliveryInstance().DeliveryUpdatesNotMe(
	//			md.AuthId,
	//			md.SessionId,
	//			md.NetlibSessionId,
	//			[]int32{u.Id},
	//			updates.ToUpdates().Encode())
	//	} else {
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
	//glog.Infof("messages.createChat#9cb126e - reply: {%v}", reply)
	//return
	return nil, fmt.Errorf("Not impl MessagesCreateChat")
 */
}
