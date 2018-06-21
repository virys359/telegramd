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

package updates

import (
	"github.com/nebulaim/telegramd/mtproto"
	// "github.com/nebulaim/telegramd/biz_model/logic/message"
	"time"
)

type UpdatesLogic struct {
	ownerUserId int32
	message     *mtproto.Message
	updates     []*mtproto.Update
	users       []*mtproto.User
	chats       []*mtproto.Chat
	date        int32
}

/////////////////////////////////////////////////////////////////////////////////////////
func NewUpdatesLogic(userId int32) *UpdatesLogic {
	return &UpdatesLogic{
		ownerUserId: userId,
		date:        int32(time.Now().Unix()),
	}
}

func NewUpdatesLogicByMessage(userId int32, message *mtproto.Message) *UpdatesLogic {
	return &UpdatesLogic{
		ownerUserId: userId,
		message:     message,
	}
}

func NewUpdatesLogicByUpdate(userId int32, update *mtproto.Update) *UpdatesLogic {
	return &UpdatesLogic{
		ownerUserId: userId,
		updates:     []*mtproto.Update{update},
	}
}

func NewUpdatesLogicByUpdates(userId int32, updateList []*mtproto.Update) *UpdatesLogic {
	return &UpdatesLogic{
		ownerUserId: userId,
		updates:     updateList,
	}
}

// updateShortMessage#914fbf11 flags:# out:flags.1?true mentioned:flags.4?true media_unread:flags.5?true silent:flags.13?true id:int user_id:int message:string pts:int pts_count:int date:int fwd_from:flags.2?MessageFwdHeader via_bot_id:flags.11?int reply_to_msg_id:flags.3?int entities:flags.7?Vector<MessageEntity> = Updates;
// message#44f9b43d flags:# out:flags.1?true mentioned:flags.4?true media_unread:flags.5?true silent:flags.13?true post:flags.14?true id:int from_id:flags.8?int to_id:Peer fwd_from:flags.2?MessageFwdHeader via_bot_id:flags.11?int reply_to_msg_id:flags.3?int date:int message:string media:flags.9?MessageMedia reply_markup:flags.6?ReplyMarkup entities:flags.7?Vector<MessageEntity> views:flags.10?int edit_date:flags.15?int post_author:flags.16?string grouped_id:flags.17?long = Message;
func messageToUpdateShortMessage(message2* mtproto.Message) (shortMessage *mtproto.TLUpdateShortMessage) {
	// TODO(@benqi): check message2.ToId
	var (
		userId int32
	)

	switch message2.GetConstructor() {
	case mtproto.TLConstructor_CRC32_message:
		message := message2.To_Message()
		if message.GetOut() {
			userId = message.GetToId().GetData2().GetUserId()
		} else {
			userId = message.GetFromId()
		}
		shortMessage = &mtproto.TLUpdateShortMessage{Data2: &mtproto.Updates_Data{
			Out:          message.GetOut(),
			Mentioned:    message.GetMentioned(),
			MediaUnread:  message.GetMediaUnread(),
			Silent:       message.GetSilent(),
			Id:           message.GetId(),
			UserId:       userId,
			Message:      message.GetMessage(),
			Date:         message.GetDate(),
			FwdFrom:      message.GetFwdFrom(),
			ViaBotId:     message.GetViaBotId(),
			ReplyToMsgId: message.GetReplyToMsgId(),
			Entities:     message.GetEntities(),
		}}
	case mtproto.TLConstructor_CRC32_messageService:
	default:
		// TODO(@benqi): error
	}
	return
}

func messageToUpdateShortChatMessage(message2* mtproto.Message) (shortMessage *mtproto.TLUpdateShortMessage) {
	return
}

//// updateShortSentMessage#11f1331c flags:# out:flags.1?true id:int pts:int pts_count:int date:int media:flags.9?MessageMedia entities:flags.7?Vector<MessageEntity> = Updates;
func messageToUpdateShortSentMessage(message2* mtproto.Message) (sentMessage *mtproto.TLUpdateShortSentMessage) {
	switch message2.GetConstructor() {
	case mtproto.TLConstructor_CRC32_message:
		message := message2.To_Message()
		sentMessage = &mtproto.TLUpdateShortSentMessage{Data2: &mtproto.Updates_Data{
			Out: message.GetOut(),
			Id:   message.GetId(),
			// Pts:,
			// PtsCount,
			Date:         message.GetDate(),
			Media: message.GetMedia(),
			Entities: message.GetEntities(),
		}}
	case mtproto.TLConstructor_CRC32_messageService:
	default:
		// TODO(@benqi): error
	}
	return
}

/////////////////////////////////////////////////////////////////////////////////////////
func (this *UpdatesLogic) ToUpdateTooLong() *mtproto.Updates {
	return mtproto.NewTLUpdatesTooLong().To_Updates()
}

func (this *UpdatesLogic) ToUpdateShortMessage() *mtproto.Updates {
	if this.message == nil {
		// TODO(@benqi): panic
	}

	shortMessage := messageToUpdateShortMessage(this.message)
	return shortMessage.To_Updates()
}

func (this *UpdatesLogic) ToUpdateShortChatMessage() *mtproto.Updates {
	if this.message == nil {
		// TODO(@benqi): panic
	}

	shortMessage := messageToUpdateShortChatMessage(this.message)
	return shortMessage.To_Updates()
}

func (this *UpdatesLogic) ToUpdateShort() *mtproto.Updates {
	if len(this.updates) != 1 {
		// TODO(@benqi): panic
	}

	updateShort := &mtproto.TLUpdateShort{Data2: &mtproto.Updates_Data{
		Update: this.updates[0],
		Date:   this.date,
	}}
	return updateShort.To_Updates()
}

func (this *UpdatesLogic) ToUpdatesCombined() *mtproto.Updates {
	updatesCombined := mtproto.NewTLUpdatesCombined()
	return updatesCombined.To_Updates()
}

func (this *UpdatesLogic) ToUpdates() *mtproto.Updates {
	updates := &mtproto.TLUpdates{Data2: &mtproto.Updates_Data{
		Updates: this.updates,
		Users:   this.users,
		Chats:   this.chats,
		Date:    this.date,
	}}
	return updates.To_Updates()
}

func (this *UpdatesLogic) ToUpdateShortSentMessage() *mtproto.Updates {
	if this.message == nil {
		// TODO(@benqi): panic
	}

	sentMessage := messageToUpdateShortSentMessage(this.message)
	return sentMessage.To_Updates()
}

/////////////////////////////////////////////////////////////////////////////////////////
func (this *UpdatesLogic) AddUpdateNewMessage(message *mtproto.Message) {
	updateNewMessage := &mtproto.TLUpdateNewMessage{Data2: &mtproto.Update_Data{
		Message_1: message,
	}}
	this.updates = append(this.updates, updateNewMessage.To_Update())
}

func (this *UpdatesLogic) AddUpdateNewChannelMessage(message *mtproto.Message) {
	updateNewChannelMessage := &mtproto.TLUpdateNewChannelMessage{Data2: &mtproto.Update_Data{
		Message_1: message,
	}}
	this.updates = append(this.updates, updateNewChannelMessage.To_Update())
}

//
//func (this *UpdatesLogic) AddUpdateNewMessageAndMessageId(message *logic.MessageBox) {
//	updateMessageID := &mtproto.TLUpdateMessageID{Data2: &mtproto.Update_Data{
//		Id_4:     message.MessageId,
//		RandomId: message.RandomId,
//	}}
//	this.updates = append(this.updates, updateMessageID.To_Update())
//	updateNewMessage := &mtproto.TLUpdateNewMessage{Data2: &mtproto.Update_Data{
//		Message_1: message.Message,
//	}}
//	this.updates = append(this.updates, updateNewMessage.To_Update())
//}
//
//func (this *UpdatesLogic) AddUpdateByMessageBox(message *logic.MessageBox) {
//	updateMessageID := &mtproto.TLUpdateMessageID{Data2: &mtproto.Update_Data{
//		Id_4:     message.MessageId,
//		RandomId: message.RandomId,
//	}}
//	this.updates = append(this.updates, updateMessageID.To_Update())
//	updateNewMessage := &mtproto.TLUpdateNewMessage{Data2: &mtproto.Update_Data{
//		Message_1: message.Message,
//	}}
//	this.updates = append(this.updates, updateNewMessage.To_Update())
//}
//

func (this *UpdatesLogic) AddUpdateMessageId(messageId int32, randomId int64) {
	updateMessageID := &mtproto.TLUpdateMessageID{Data2: &mtproto.Update_Data{
		Id_4:     messageId,
		RandomId: randomId,
	}}

	this.updates = append(this.updates, updateMessageID.To_Update())
}

func (this *UpdatesLogic) PushTopUpdateMessageId(messageId int32, randomId int64) {
	updateMessageID := &mtproto.TLUpdateMessageID{Data2: &mtproto.Update_Data{
		Id_4:     messageId,
		RandomId: randomId,
	}}

	updates2 := make([]*mtproto.Update, 0, 1 + len(this.updates))
	updates2 = append(updates2, updateMessageID.To_Update())
	this.updates = append(updates2, this.updates...)
	// this.updates = updates2
	// this.updates = append(this.updates, updateMessageID.To_Update())
}

/////////////////////////////////////////////////////////////////////////////////////////
func (this *UpdatesLogic) AddUpdates(updateList []*mtproto.Update) {
	this.updates = append(this.updates, updateList...)
}

func (this *UpdatesLogic) AddUpdate(update *mtproto.Update) {
	this.updates = append(this.updates, update)
}

func (this *UpdatesLogic) AddChats(chatList []*mtproto.Chat) {
	this.chats = append(this.chats, chatList...)
}

func (this *UpdatesLogic) AddChat(chat *mtproto.Chat) {
	this.chats = append(this.chats, chat)
}

func (this *UpdatesLogic) AddUsers(userList []*mtproto.User) {
	this.users = append(this.users, userList...)
}

func (this *UpdatesLogic) AddUser(user *mtproto.User) {
	this.users = append(this.users, user)
}

/////////////////////////////////////////////////////////////////////////////////////////
// TODO(@benqi): check error
func (this *UpdatesLogic) SetupState(state *mtproto.ClientUpdatesState) {
	pts := state.GetPts() - state.GetPtsCount() + 1

	// TODO(@benqi): setup seq
	for _, update := range this.updates {
		switch update.GetConstructor() {
		case mtproto.TLConstructor_CRC32_updateNewMessage,
			mtproto.TLConstructor_CRC32_updateDeleteMessages,
			mtproto.TLConstructor_CRC32_updateReadHistoryOutbox,
			mtproto.TLConstructor_CRC32_updateReadHistoryInbox,
			mtproto.TLConstructor_CRC32_updateWebPage,
			mtproto.TLConstructor_CRC32_updateReadMessagesContents,
			mtproto.TLConstructor_CRC32_updateEditMessage,

			// channel
			mtproto.TLConstructor_CRC32_updateNewChannelMessage,
			mtproto.TLConstructor_CRC32_updateDeleteChannelMessages,
			mtproto.TLConstructor_CRC32_updateEditChannelMessage,
			mtproto.TLConstructor_CRC32_updateChannelWebPage:

			update.Data2.Pts = pts
			update.Data2.PtsCount = 1
			pts += 1
		}
	}
}
