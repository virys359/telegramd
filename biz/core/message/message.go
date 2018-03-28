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

package message

import (
	"github.com/nebulaim/telegramd/mtproto"
	"github.com/nebulaim/telegramd/biz/base"
	"encoding/json"
	"time"
	"github.com/nebulaim/telegramd/biz/dal/dataobject"
	// "github.com/nebulaim/telegramd/biz/model"
	base2 "github.com/nebulaim/telegramd/baselib/base"
	"github.com/nebulaim/telegramd/biz/dal/dao"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/nebulaim/telegramd/biz/core/updates"
)

//type InboxMessageList struct {
//	// UserIds []int32
//	// Messages []*mtproto.Message
//}

type MessageBoxObserver interface {
	OnOutboxCreated(clearDraft bool, outbox *MessageBox)
	OnInboxCreated(outbox *MessageBox)
}

type MessageBox struct {
	UserId             int32
	MessageId          int32
	DialogMessageId    int64
	RandomId		   int64
	Message            *mtproto.Message
}

func InsertMessageToOutbox(fromId int32, peer *base.PeerUtil, clientRandomId int64, message2 *mtproto.Message) (box *MessageBox) {
	now := int32(time.Now().Unix())
	messageDO := &dataobject.MessagesDO{
		UserId:fromId,
		UserMessageBoxId: int32(updates.NextMessageBoxId(base2.Int32ToString(fromId))),
		DialogMessageId: base.NextSnowflakeId(),
		SenderUserId: fromId,
		MessageBoxType: MESSAGE_BOX_TYPE_OUTGOING,
		PeerType: int8(peer.PeerType),
		PeerId: peer.PeerId,
		RandomId: clientRandomId,
		Date2: now,
		Deleted: 0,
	}

	// var mentioned = false

	switch message2.GetConstructor() {
	case mtproto.TLConstructor_CRC32_messageEmpty:
		messageDO.MessageType = MESSAGE_TYPE_MESSAGE_EMPTY
	case mtproto.TLConstructor_CRC32_message:
		messageDO.MessageType = MESSAGE_TYPE_MESSAGE
		message := message2.To_Message()

		// mentioned = message.GetMentioned()
		message.SetId(messageDO.UserMessageBoxId)
	case mtproto.TLConstructor_CRC32_messageService:
		messageDO.MessageType = MESSAGE_TYPE_MESSAGE_SERVICE
		message := message2.To_MessageService()

		// mentioned = message.GetMentioned()
		message.SetId(messageDO.UserMessageBoxId)
	}

	messageData, _ := json.Marshal(message2)
	messageDO.MessageData = string(messageData)

	// TODO(@benqi): pocess clientRandomId dup
	dao.GetMessagesDAO(dao.DB_MASTER).Insert(messageDO)

	box = &MessageBox{
		UserId:          fromId,
		MessageId:       messageDO.UserMessageBoxId,
		DialogMessageId: messageDO.DialogMessageId,
		RandomId:        clientRandomId,
		Message:         message2,
	}

	// dialog
	// GetDialogModel().CreateOrUpdateByOutbox(fromId, peer.PeerType, peer.PeerId, messageDO.UserMessageBoxId, mentioned)
	//boxId = messageDO.UserMessageBoxId
	//dialogMessageId = messageDO.DialogMessageId
	return
}

func (this *MessageBox) InsertMessageToInbox(fromId int32, peer *base.PeerUtil) (MessageBoxList, error) {
	switch peer.PeerType {
	case base.PEER_USER:
		return this.insertUserMessageToInbox(fromId, peer)
	case base.PEER_CHAT:
		return this.insertChatMessageToInbox(fromId, peer)
	case base.PEER_CHANNEL:
		return this.insertChannelMessageToInbox(fromId, peer)
	default:
		//	panic("invalid peer")
		return nil, fmt.Errorf("invalid peer")
	}
}

func getPeerMessageId(userId, messageId, peerId int32) int32 {
	do := dao.GetMessagesDAO(dao.DB_SLAVE).SelectPeerMessageId(peerId, userId, messageId)
	if do == nil {
		return 0
	} else {
		return do.UserMessageBoxId
	}
}

// 发送到收件箱
func (this *MessageBox) insertUserMessageToInbox(fromId int32, peer *base.PeerUtil) (MessageBoxList, error) {
	now := int32(time.Now().Unix())
	messageDO := &dataobject.MessagesDO{
		UserId:peer.PeerId,
		UserMessageBoxId: int32(updates.NextMessageBoxId(base2.Int32ToString(peer.PeerId))),
		DialogMessageId: this.DialogMessageId,
		SenderUserId: this.UserId,
		MessageBoxType: MESSAGE_BOX_TYPE_INCOMING,
		PeerType: int8(peer.PeerType),
		PeerId: peer.PeerId,
		RandomId: this.RandomId,
		Date2: now,
		Deleted: 0,
	}

	inboxMessage := proto.Clone(this.Message).(*mtproto.Message)
	// var mentioned = false

	switch this.Message.GetConstructor() {
	case mtproto.TLConstructor_CRC32_messageEmpty:
		messageDO.MessageType = MESSAGE_TYPE_MESSAGE_EMPTY
	case mtproto.TLConstructor_CRC32_message:
		messageDO.MessageType = MESSAGE_TYPE_MESSAGE

		m2 := inboxMessage.To_Message()
		m2.SetOut(false)
		if m2.GetReplyToMsgId() != 0 {
			replyMsgId := getPeerMessageId(fromId, m2.GetReplyToMsgId(), peer.PeerId)
			m2.SetReplyToMsgId(replyMsgId)
		}
		m2.SetId(messageDO.UserMessageBoxId)
		// mentioned = m2.GetMentioned()
	case mtproto.TLConstructor_CRC32_messageService:
		messageDO.MessageType = MESSAGE_TYPE_MESSAGE_SERVICE

		m2 := inboxMessage.To_MessageService()
		m2.SetOut(false)
		m2.SetId(messageDO.UserMessageBoxId)

		// mentioned = m2.GetMentioned()
	}

	messageData, _ := json.Marshal(inboxMessage)
	messageDO.MessageData = string(messageData)

	// TODO(@benqi): rpocess clientRandomId dup
	dao.GetMessagesDAO(dao.DB_MASTER).Insert(messageDO)
	// dialog
	// model.GetDialogModel().CreateOrUpdateByInbox(peer.PeerId, peer.PeerType, fromId, messageDO.UserMessageBoxId, mentioned)

	box := &MessageBox{
		UserId:          fromId,
		MessageId:       messageDO.UserMessageBoxId,
		DialogMessageId: messageDO.DialogMessageId,
		RandomId:        this.RandomId,
		Message:         inboxMessage,
	}
	return []*MessageBox{box}, nil
	//return &InboxMessages{
	//	UserIds:  []int32{messageDO.UserId},
	//	Messages: []*mtproto.Message{inboxMessage},
	//}, nil
}

// 发送到收件箱
func (this *MessageBox) insertChatMessageToInbox(fromId int32, peer *base.PeerUtil) (MessageBoxList, error) {
	switch this.Message.GetConstructor() {
	case mtproto.TLConstructor_CRC32_message:
	case mtproto.TLConstructor_CRC32_messageService:
	default:
		panic("invalid messageEmpty type")
		// return
	}
	return []*MessageBox{}, nil
}

// 发送到收件箱
func (this *MessageBox) insertChannelMessageToInbox(fromId int32, peer *base.PeerUtil) (MessageBoxList, error) {
	switch this.Message.GetConstructor() {
	case mtproto.TLConstructor_CRC32_message:
	case mtproto.TLConstructor_CRC32_messageService:
	default:
		panic("invalid messageEmpty type")
		// return
	}
	return []*MessageBox{}, nil
}


type MessageBoxList []*MessageBox

func (this *MessageBoxList) ToMessageList() []*mtproto.Message {
	messageList := make([]*mtproto.Message, 0, len(*this))
	for _, box := range messageList {
		messageList = append(messageList, box)
	}
	return messageList
}

type MessageLogic struct {
	*mtproto.Message
	// creatorId    int32
	// Chat         *mtproto.TLChat
	// Participants []*mtproto.ChatParticipant
}
