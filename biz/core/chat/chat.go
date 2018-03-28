/*
 *  Copyright (c) 2018, https://github.com/nebulaim
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

package chat

import (
	"github.com/nebulaim/telegramd/mtproto"
	"time"
	"github.com/nebulaim/telegramd/biz/base"
	"github.com/nebulaim/telegramd/biz/dal/dataobject"
	"math/rand"
	"github.com/nebulaim/telegramd/biz/dal/dao"
)

// chat#d91cdd54 flags:#
// 	creator:flags.0?true
// 	kicked:flags.1?true
// 	left:flags.2?true
// 	admins_enabled:flags.3?true
// 	admin:flags.4?true
// 	deactivated:flags.5?true
// 	id:int
// 	title:string
// 	photo:ChatPhoto
// 	participants_count:int
// 	date:int
// 	version:int
// 	migrated_to:flags.6?InputChannel = Chat;

type ChatObserver interface {
	OnChatCreated()
	OnChatAddUser()
	OnChatDeleteUser()
	OnChatEditAdmin()
	OnChatEditTitle()
	OnChatEditPhoto()
}

type ChatLogic struct {
	// creatorId    int32
	Chat         *mtproto.TLChat
	Participants []*mtproto.ChatParticipant
}

func chatToDO(creatorId int32, chat *mtproto.TLChat) *dataobject.ChatsDO {
	return &dataobject.ChatsDO{
		CreatorUserId:    creatorId,
		CreateRandomId:   rand.Int63(),
		Title:            chat.GetTitle(),
		ParticipantCount: chat.GetParticipantsCount(),
	}
}

func chatParticipantToDO(chat *mtproto.TLChat, participant *mtproto.ChatParticipant) *dataobject.ChatParticipantsDO {
	// uId := u.GetInputUser().GetUserId()
	chatParticipantsDO := &dataobject.ChatParticipantsDO{
		ChatId: chat.GetId(),
		InvitedAt: participant.GetData2().GetInviterId(),
	}

	return chatParticipantsDO

	// dao.GetChatParticipantsDAO(dao.DB_MASTER).Insert(chatParticipantsDO)

	//chatUserDO.ChatId = chatId
	//chatUserDO.CreatedAt = base.NowFormatYMDHMS()
	//chatUserDO.State = 0
	//chatUserDO.InvitedAt = int32(time.Now().Unix())
	//chatUserDO.InviterUserId = inviterId
	//chatUserDO.JoinedAt = chatUserDO.InvitedAt
	//chatUserDO.UserId = chatUserId
	//chatUserDO.ParticipantType = participantType
	//
	//if participantType == 2 {
	//	participant2 := mtproto.NewTLChatParticipantCreator()
	//	participant2.SetUserId(chatUserId)
	//
	//	participant = participant2.To_ChatParticipant()
	//} else if participantType == 1 {
	//	participant2 := mtproto.NewTLChatParticipantAdmin()
	//	participant2.SetUserId(chatUserId)
	//	participant2.SetDate(chatUserDO.InvitedAt)
	//	participant2.SetInviterId(inviterId)
	//
	//	participant = participant2.To_ChatParticipant()
	//} else if participantType == 0 {
	//	participant2 := mtproto.NewTLChatParticipant()
	//	participant2.SetUserId(chatUserId)
	//	participant2.SetDate(chatUserDO.InvitedAt)
	//	participant2.SetInviterId(inviterId)
	//	// participants.Participants = append(participants.Participants, participant.ToChatParticipant())
	//
	//	participant = participant2.To_ChatParticipant()
	//}
	// return
}

func NewChatLogicById(chatId int32) (*ChatLogic) {
	chatDO := dao.GetChatsDAO(dao.DB_SLAVE).Select(chatId)
	if chatDO == nil {
		panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_BAD_REQUEST), "InputPeer invalid"))
	}
	chat2 := &mtproto.TLChat{Data2: &mtproto.Chat_Data{
		Id:                chatDO.Id,
		Title:             chatDO.Title,
		Photo:             mtproto.NewTLChatPhotoEmpty().To_ChatPhoto(),
		Version:           chatDO.Version,
		ParticipantsCount: chatDO.ParticipantCount,
		Date:              int32(time.Now().Unix()),
	}}
	return &ChatLogic{
		Chat: chat2,
	}
}

func NewChatLogicByCreateChat(creatorId int32, userIds []int32, title string) (chatModule *ChatLogic) {
	// TODO(@benqi): 事务
	date := int32(time.Now().Unix())
	chat := &mtproto.TLChat{Data2: &mtproto.Chat_Data{
		Title: title,
		ParticipantsCount: 1 + int32(len(userIds)),
		Date: date,
		Version: 1,
	}}

	// insert
	chatDO := chatToDO(creatorId, chat)
	chatId := int32(dao.GetChatsDAO(dao.DB_MASTER).Insert(chatDO))
	chat.SetId(chatId)

	//chatDO := &dataobject.ChatsDO{}
	//chatDO.AccessHash = rand.Int63()
	//chatDO.CreatedAt = base.NowFormatYMDHMS()
	//chatDO.CreatorUserId = userId
	//// TODO(@benqi): 使用客户端message_id
	//chatDO.CreateRandomId = rand.Int63()
	//chatDO.Title = title
	//
	//chatDO.TitleChangerUserId = userId
	//chatDO.TitleChangedAt = chatDO.CreatedAt
	//// TODO(@benqi): 使用客户端message_id
	//chatDO.TitleChangeRandomId = chatDO.AccessHash
	//
	//chatDO.AvatarChangerUserId = userId
	//chatDO.AvatarChangedAt = chatDO.CreatedAt
	//// TODO(@benqi): 使用客户端message_id
	//chatDO.AvatarChangeRandomId = chatDO.AccessHash
	//// dao.GetChatsDA()
	//chatDO.ParticipantCount = chat.GetParticipantsCount()
	//
	//// TODO(@benqi): 事务！
	//chat.SetId(int32(dao.GetChatsDAO(dao.DB_MASTER).Insert(chatDO)))
	//

	participants := make([]*mtproto.ChatParticipant, 0, 1 + len(userIds))
	creatorParticipant := &mtproto.TLChatParticipantCreator{Data2: &mtproto.ChatParticipant_Data{
		UserId: creatorId,
	}}
	participants = append(participants, creatorParticipant.To_ChatParticipant())
	for _, id := range userIds {
		participant := &mtproto.TLChatParticipant{Data2: &mtproto.ChatParticipant_Data{
			UserId:    id,
			InviterId: creatorId,
			Date:      date,
		}}
		participantDO := chatParticipantToDO(chat, participant.To_ChatParticipant())
		dao.GetChatParticipantsDAO(dao.DB_MASTER).Insert(participantDO)

		participants = append(participants, participant.To_ChatParticipant())
	}
	chatModule = &ChatLogic{
		Chat:         chat,
		Participants: participants,
	}
	return
}

func (this *ChatLogic) MakeCreateChatMessage(creatorId int32) *mtproto.Message {
	idList := this.GetChatParticipantIdList()
	action := &mtproto.TLMessageActionChatCreate{Data2: &mtproto.MessageAction_Data{
		Title: this.Chat.GetTitle(),
		Users: idList,
	}}

	peer := &base.PeerUtil{
		PeerType: base.PEER_CHAT,
		PeerId:   this.Chat.GetId(),
	}

	message := &mtproto.TLMessageService{Data2: &mtproto.Message_Data{
		Date: this.Chat.GetDate(),
		FromId: creatorId,
		ToId: peer.ToPeer(),
		Action: action.To_MessageAction(),
	}}
	return message.To_Message()
}

func (this *ChatLogic) MakeAddUserMessage(inviterId, chatUserId int32) *mtproto.Message {
	// idList := this.GetChatParticipantIdList()
	action := &mtproto.TLMessageActionChatAddUser{Data2: &mtproto.MessageAction_Data{
		Title: this.Chat.GetTitle(),
		Users: []int32{chatUserId},
	}}

	peer := &base.PeerUtil{
		PeerType: base.PEER_CHAT,
		PeerId:   this.Chat.GetId(),
	}

	message := &mtproto.TLMessageService{Data2: &mtproto.Message_Data{
		Date: this.Chat.GetDate(),
		FromId: inviterId,
		ToId: peer.ToPeer(),
		Action: action.To_MessageAction(),
	}}
	return message.To_Message()
}

func (this *ChatLogic) MakeDeleteUserMessage(operatorId, chatUserId int32) *mtproto.Message {
	// idList := this.GetChatParticipantIdList()
	action := &mtproto.TLMessageActionChatDeleteUser{Data2: &mtproto.MessageAction_Data{
		Title:  this.Chat.GetTitle(),
		UserId: chatUserId,
	}}

	peer := &base.PeerUtil{
		PeerType: base.PEER_CHAT,
		PeerId:   this.Chat.GetId(),
	}

	message := &mtproto.TLMessageService{Data2: &mtproto.Message_Data{
		Date: this.Chat.GetDate(),
		FromId: operatorId,
		ToId: peer.ToPeer(),
		Action: action.To_MessageAction(),
	}}
	return message.To_Message()
}

func (this *ChatLogic) checkOrLoadChatParticipantList() {
	if len(this.Participants) == 0 {
		// TODO(@benqi): Load from DB
		chatId := this.Chat.GetId()
		chatUsersDOList := dao.GetChatParticipantsDAO(dao.DB_SLAVE).SelectByChatId(chatId)

		// updateChatParticipants := &mtproto.TLUpdateChatParticipants{}
		//participants := mtproto.NewTLChatParticipants()
		//participants.SetChatId(chatId)
		//participants.SetVersion(1)
		for _, chatUsersDO := range chatUsersDOList {
			// uId := u.GetInputUser().GetUserId()
			if chatUsersDO.ParticipantType == 2 {
				// chatUserDO.IsAdmin = 1
				participant := mtproto.NewTLChatParticipantCreator()
				participant.SetUserId(chatUsersDO.UserId)
				this.Participants = append(this.Participants, participant.To_ChatParticipant())
				// participants.SetParticipants(append(participants.Data2.Participants, participant.To_ChatParticipant()))
			} else if chatUsersDO.ParticipantType == 1 {
				participant := mtproto.NewTLChatParticipantAdmin()
				participant.SetUserId(chatUsersDO.UserId)
				participant.SetInviterId(chatUsersDO.InviterUserId)
				participant.SetDate(chatUsersDO.JoinedAt)
				this.Participants = append(this.Participants, participant.To_ChatParticipant())
				// participants.Data2.Participants = append(participants.Data2.Participants, participant.To_ChatParticipant())
			} else if chatUsersDO.ParticipantType == 0 {
				participant := mtproto.NewTLChatParticipant()
				participant.SetUserId(chatUsersDO.UserId)
				participant.SetInviterId(chatUsersDO.InviterUserId)
				participant.SetDate(chatUsersDO.JoinedAt)
				this.Participants = append(this.Participants, participant.To_ChatParticipant())
				// participants.Data2.Participants = append(participants.Data2.Participants, participant.To_ChatParticipant())
			}
		}
		// return participants
	}
}

func (this *ChatLogic) GetChatParticipantIdList() []int32 {
	this.checkOrLoadChatParticipantList()

	idList := make([]int32, 0, len(this.Participants))
	for _, v := range this.Participants {
		idList = append(idList, v.GetData2().GetUserId())
	}
	return idList
}

func (this *ChatLogic) GetChatParticipants() *mtproto.TLChatParticipants {
	this.checkOrLoadChatParticipantList()

	return &mtproto.TLChatParticipants{Data2: &mtproto.ChatParticipants_Data{
		ChatId: this.Chat.GetId(),
		Participants: this.Participants,
		Version: this.Chat.GetVersion(),
	}}
}

func (this *ChatLogic) AddChatUser(inviterId, userId int32) {
	this.checkOrLoadChatParticipantList()

	// TODO(@benqi): check userId exisits

	participant := &mtproto.TLChatParticipant{Data2: &mtproto.ChatParticipant_Data{
		UserId:    userId,
		InviterId: inviterId,
		Date:      int32(time.Now().Unix()),
	}}
	participantDO := chatParticipantToDO(this.Chat, participant.To_ChatParticipant())
	this.Participants = append(this.Participants, participant.To_ChatParticipant())
	dao.GetChatParticipantsDAO(dao.DB_MASTER).Insert(participantDO)
}

func (this *ChatLogic) DeleteChatUser(userId int32) {
	this.checkOrLoadChatParticipantList()

	// messageActionChatDeleteUser
	// dao.GetChatUsersDAO(dao.DB_MASTER).DeleteChatUser(chat.Id, deleteChatUserId)
}

func (this *ChatLogic) EditChatUser(userId int32) {
}

func (this *ChatLogic) EditChatTitle(userId int32) {
}

func (this *ChatLogic) EditChatPhoto(userId int32) {
}

func (this *ChatLogic) EditChatAdmin(userId int32) {
}
