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

package message

import (
	"github.com/nebulaim/telegramd/mtproto"
	"github.com/nebulaim/telegramd/biz/base"
	"github.com/nebulaim/telegramd/biz/dal/dao"
	// "github.com/golang/glog"
	base2 "github.com/nebulaim/telegramd/baselib/base"
	"github.com/nebulaim/telegramd/biz/dal/dataobject"
	"time"
	// "github.com/nebulaim/telegramd/baselib/logger"
	"fmt"
	"encoding/json"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/glog"
	update2 "github.com/nebulaim/telegramd/biz/core/update"
	// "github.com/nebulaim/telegramd/biz/core/peer"
)

const (
	MESSAGE_TYPE_UNKNOWN = 0
	MESSAGE_TYPE_MESSAGE_EMPTY = 1
	MESSAGE_TYPE_MESSAGE = 2
	MESSAGE_TYPE_MESSAGE_SERVICE = 3
)
const (
	MESSAGE_BOX_TYPE_INCOMING = 0
	MESSAGE_BOX_TYPE_OUTGOING = 1
)

const (
	PTS_UNKNOWN = 0
	PTS_MESSAGE_OUTBOX = 1
	PTS_MESSAGE_INBOX = 2
	PTS_READ_HISTORY_OUTBOX = 3
	PTS_READ_HISTORY_INBOX = 4
)

/*
func (m *messageModel) getMessagesByIDList(idList []int32, order bool) (messages []*mtproto.Message) {
	// TODO(@benqi): Check messageDAO
	messageDAO := dao.GetMessagesDAO(dao.DB_SLAVE)

	var messagesDOList []dataobject.MessagesDO

	if order {
		// TODO(@benqi): 不推给DB，程序内排序
		messagesDOList = messageDAO.SelectOrderByIdList(idList)
	} else {
		messagesDOList = messageDAO.SelectByIdList(idList)
	}

	messages = []*mtproto.Message{}
	for _, messageDO := range messagesDOList {
		message := &mtproto.Message{
			Data2: &mtproto.Message_Data{},
		}
		switch messageDO.MessageType {
		case MESSAGE_TYPE_MESSAGE_EMPTY:
			message.Constructor = mtproto.TLConstructor_CRC32_messageEmpty
		case MESSAGE_TYPE_MESSAGE:
			message.Constructor = mtproto.TLConstructor_CRC32_message
			// err := proto.Unmarshal(messageDO.MessageData, message)
			err := json.Unmarshal([]byte(messageDO.MessageData), message)
			if err != nil {
				glog.Errorf("GetMessagesByIDList - Unmarshal message(%d)error: %v", messageDO.Id, err)
				continue
			}
			message.Data2.Id = messageDO.Id
			//messages = append(messages, message.To_Message())
		case MESSAGE_TYPE_MESSAGE_SERVICE:
			message.Constructor = mtproto.TLConstructor_CRC32_messageService
			// err := proto.Unmarshal(messageDO.MessageData, message)
			err := json.Unmarshal([]byte(messageDO.MessageData), message)
			if err != nil {
				glog.Errorf("GetMessagesByIDList - Unmarshal message(%d)error: %v", messageDO.Id, err)
				continue
			}
			message.Data2.Id = messageDO.Id
		default:
			glog.Error("Invalid messageType, db's data error: %s", logger.JsonDebugData(messageDO))
			continue
		}

		messages = append(messages, message)
	}
	glog.Infof("GetMessagesByIDList(%s) - %s", base2.JoinInt32List(idList, ","), logger.JsonDebugData(messages))
	return
}

func (m *messageModel) GetMessagesByPeerAndMessageIdList(userId int32, idList []int32) (messages []*mtproto.Message) {
	boxesList := dao.GetMessageBoxesDAO(dao.DB_SLAVE).SelectByMessageIdList(userId, idList)
	return m.getMessagesByMessageBoxes(boxesList, true)
}

func (m *messageModel) GetMessagesByPeerAndMessageIdList2(userId int32, idList []int32) (messages []*mtproto.Message) {
	boxesList := dao.GetMessageBoxesDAO(dao.DB_SLAVE).SelectByMessageIdList(userId, idList)
	return m.getMessagesByMessageBoxes(boxesList, false)
}

// Loadhistory
func (m *messageModel) LoadBackwardHistoryMessages(userId int32, peerType , peerId int32, offset int32, limit int32) []*mtproto.Message {
	// TODO(@benqi): chat and channel
	boxesList := dao.GetMessageBoxesDAO(dao.DB_SLAVE).SelectBackwardByPeerUserOffsetLimit(userId, peerId, int8(peerType), offset, limit)
	glog.Infof("GetMessagesByUserIdPeerOffsetLimit - boxesList: %v", boxesList)
	if len(boxesList) == 0 {
		return make([]*mtproto.Message, 0)
	}
	return m.getMessagesByMessageBoxes(boxesList, true)
}

func (m *messageModel) LoadForwardHistoryMessages(userId int32, peerType , peerId int32, offset int32, limit int32) []*mtproto.Message {
	// TODO(@benqi): chat and channel
	boxesList := dao.GetMessageBoxesDAO(dao.DB_SLAVE).SelectForwardByPeerUserOffsetLimit(userId, peerId, int8(peerType), offset, limit)
	glog.Infof("GetMessagesByUserIdPeerOffsetLimit - boxesList: %v", boxesList)
	if len(boxesList) == 0 {
		return make([]*mtproto.Message, 0)
	}
	return m.getMessagesByMessageBoxes(boxesList, true)
}

// TODO(@benqi): 这种方法比较土
func searchBoxByMessageId(messageId int32, boxes []dataobject.MessageBoxesDO) *dataobject.MessageBoxesDO {
	for _, box := range boxes {
		if messageId == box.MessageId {
			return &box
		}
	}
	return nil
}

func (m *messageModel) getMessagesByMessageBoxes(boxes []dataobject.MessageBoxesDO, order bool) []*mtproto.Message {
	glog.Infof("getMessagesByMessageBoxes - boxes: %s", logger.JsonDebugData(boxes))
	messageIdList := make([]int32, 0, len(boxes))
	for _, do := range boxes {
		messageIdList = append(messageIdList, do.MessageId)
	}
	messageList := m.getMessagesByIDList(messageIdList, order)
	for i, message := range messageList {
		boxDO := searchBoxByMessageId(message.Data2.Id, boxes)
		if boxDO == nil {
			glog.Error("getMessagesByMessageBoxes - Not found box, db's data error!!!")
			continue
		}
		if boxDO.MessageBoxType == MESSAGE_BOX_TYPE_OUTGOING {
			message.Data2.Out = true
		} else {
			message.Data2.Out = false
		}
		// message.Data2.Silent = true
		message.Data2.MediaUnread = boxDO.MediaUnread != 0
		message.Data2.Mentioned = false

		// 使用UserMessageBoxId作为messageBoxId
		message.Data2.Id = boxDO.UserMessageBoxId
		glog.Infof("getMessagesByMessageBoxes - message(%d): %s", i, logger.JsonDebugData(message))
	}

	return messageList
}

func (m *messageModel) GetMessagesByGtPts(userId int32, pts int32) []*mtproto.Message {
	boxDOList := dao.GetMessageBoxesDAO(dao.DB_SLAVE).SelectByGtPts(userId, pts)
	if len(boxDOList) == 0 {
		return make([]*mtproto.Message, 0)
	} else {
		return m.getMessagesByMessageBoxes(boxDOList, false)
	}
}

func (m *messageModel) GetLastPtsByUserId(userId int32) int32 {
	do := dao.GetMessageBoxesDAO(dao.DB_SLAVE).SelectLastPts(userId)
	if do == nil {
		return 0
	} else {
		return  do.Pts
	}
}

// CreateMessage
func (m *messageModel) CreateMessageBoxes(userId, fromId int32, peerType int32, peerId int32, incoming bool, messageId int32) (int32) {
	messageBox := &dataobject.MessageBoxesDO{
		UserId:       userId,
		SenderUserId: fromId,
		PeerType:     int8(peerType),
		PeerId:       peerId,
		MessageId:    messageId,
		Date2:        int32(time.Now().Unix()),
		CreatedAt:    base2.NowFormatYMDHMS(),
	}

	if incoming {
		messageBox.MessageBoxType = MESSAGE_BOX_TYPE_INCOMING
		outPts := GetSequenceModel().NextPtsId(base2.Int32ToString(messageBox.UserId))
		messageBox.Pts = int32(outPts)
	} else {
		messageBox.MessageBoxType = MESSAGE_BOX_TYPE_OUTGOING
		inPts := GetSequenceModel().NextPtsId(base2.Int32ToString(messageBox.UserId))
		messageBox.Pts = int32(inPts)
	}

	dao.GetMessageBoxesDAO(dao.DB_MASTER).Insert(messageBox)
	return messageBox.Pts
}

// CreateHistoryMessage2
func (m *messageModel) CreateHistoryMessage2(fromId int32, peer *helper.PeerUtil, randomId int64, date int32, message *mtproto.Message) (messageId int32) {
	// TODO(@benqi): 重复插入出错处理
	messageDO := &dataobject.MessagesDO{
		SenderUserId: fromId,
		PeerType:     peer.PeerType,
		PeerId:       peer.PeerId,
		RandomId:     randomId,
		Date2:        date,
	}

	switch message.GetConstructor() {
	case mtproto.TLConstructor_CRC32_messageEmpty:
		messageDO.MessageType = MESSAGE_TYPE_MESSAGE_EMPTY
	case mtproto.TLConstructor_CRC32_message:
		messageDO.MessageType = MESSAGE_TYPE_MESSAGE
	case mtproto.TLConstructor_CRC32_messageService:
		messageDO.MessageType = MESSAGE_TYPE_MESSAGE_SERVICE
	default:
		panic(fmt.Errorf("Invalid message_type: {%v}", message))
	}

	// TODO(@benqi): 测试阶段使用Json!!!
	// messageDO.MessageData, _ = proto.Marshal(message)
	messageData, _ := json.Marshal(message)
	messageDO.MessageData = string(messageData)
	messageId = int32(dao.GetMessagesDAO(dao.DB_MASTER).Insert(messageDO))
	return
}

//func (m *messageModel) MakeUpdatesByMessage(randomId int64, message *mtproto.TLMessageService) (updates *mtproto.Updates) {
//	//// 插入消息
//	peer := helper.FromPeer(message.ToId)
//	messageDO := &dataobject.MessagesDO{}
//
//	messageDO.UserId = message.FromId
//	messageDO.PeerType = peer.PeerType
//	messageDO.PeerId = peer.PeerId
//	messageDO.RandomId = randomId
//	messageDO.Date2 = message.Date
//
//	messageId := dao.GetMessagesDAO(dao.DB_MASTER).Insert(messageDO)
//	message.Id = int32(messageId)
//	return

type IDMessage struct {
	UserId int32
	MessageBoxId int32
}
*/

////////////////////////////////////////////////////////////////////////////////////////////////////
// Loadhistory
func  LoadBackwardHistoryMessages(userId int32, peerType , peerId int32, offset int32, limit int32) (messages []*mtproto.Message) {
	// TODO(@benqi): chat and channel

	if peerType == base.PEER_CHANNEL {
		boxDOList := dao.GetChannelMessageBoxesDAO(dao.DB_SLAVE).SelectBackwardByOffsetLimit(peerId, offset, limit)
		if len(boxDOList) == 0 {
			messages = []*mtproto.Message{}
		} else {
			messages = make([]*mtproto.Message, 0, len(boxDOList))
			for i := 0; i < len(boxDOList); i++ {
				// TODO(@benqi): check data
				messageDO := dao.GetMessageDatasDAO(dao.DB_SLAVE).SelectByMessageId(boxDOList[i].MessageId)
				if messageDO == nil {
					continue
				}
				m, _ := doToChannelMessage(messageDO)
				if m != nil {
					messages = append(messages, m)
				}
			}
		}
	} else {
		var (
			doList []dataobject.MessagesDO
		)

		switch peerType {
		case base.PEER_USER:
			// doList = dao.GetMessagesDAO(dao.DB_SLAVE).SelectForwardByPeerUserOffsetLimit(userId, peerId, int8(peerType), offset, limit)
			doList = dao.GetMessagesDAO(dao.DB_SLAVE).SelectBackwardByPeerUserOffsetLimit(userId, peerId, int8(peerType), offset, limit)
		case base.PEER_CHAT:
			// doList = dao.GetMessagesDAO(dao.DB_SLAVE).SelectForwardByPeerOffsetLimit(userId, int8(peerType), peerId, offset, limit)
			doList = dao.GetMessagesDAO(dao.DB_SLAVE).SelectBackwardByPeerOffsetLimit(userId, int8(peerType), peerId, offset, limit)
		case base.PEER_CHANNEL:
			// boxDOList := dao.GetChannelMessageBoxesDAO(dao.DB_SLAVE).SelectBackwardByOffsetLimit(peerId, offset, limit)
			// _ = boxDOList
		default:
		}

		glog.Infof("GetMessagesByUserIdPeerOffsetLimit - boxesList: %v", doList)
		if len(doList) == 0 {
			messages = []*mtproto.Message{}
		} else {
			messages = make([]*mtproto.Message, 0, len(doList))
			for _, do := range doList {
				// TODO(@benqi): check data
				m, _ := messageDOToMessage(&do)
				messages = append(messages, m)
			}
		}
	}

	return
}

func LoadForwardHistoryMessages(userId int32, peerType , peerId int32, offset int32, limit int32) (messages []*mtproto.Message) {
	// TODO(@benqi): chat and channel

	if peerType == base.PEER_CHANNEL {
		boxDOList := dao.GetChannelMessageBoxesDAO(dao.DB_SLAVE).SelectForwardByOffsetLimit(peerId, offset, limit)
		if len(boxDOList) == 0 {
			messages = []*mtproto.Message{}
		} else {
			messages = make([]*mtproto.Message, 0, len(boxDOList))
			for i := 0; i < len(boxDOList); i++ {
				// TODO(@benqi): check data
				messageDO := dao.GetMessageDatasDAO(dao.DB_SLAVE).SelectByMessageId(boxDOList[i].MessageId)
				if messageDO == nil {
					continue
				}
				m, _ := doToChannelMessage(messageDO)
				if m != nil {
					messages = append(messages, m)
				}
			}
		}
	} else {
		var (
			doList []dataobject.MessagesDO
		)

		switch peerType {
		case base.PEER_USER:
			doList = dao.GetMessagesDAO(dao.DB_SLAVE).SelectForwardByPeerUserOffsetLimit(userId, peerId, int8(peerType), offset, limit)
		case base.PEER_CHAT:
			doList = dao.GetMessagesDAO(dao.DB_SLAVE).SelectForwardByPeerOffsetLimit(userId, int8(peerType), peerId, offset, limit)
		case base.PEER_CHANNEL:
			//boxDOList := dao.GetChannelMessageBoxesDAO(dao.DB_SLAVE).SelectForwardByOffsetLimit(peerId, offset, limit)
			//_ = boxDOList
		default:
		}


		glog.Infof("GetMessagesByUserIdPeerOffsetLimit - boxesList: %v", doList)
		if len(doList) == 0 {
			messages = []*mtproto.Message{}
		} else {
			messages = make([]*mtproto.Message, 0, len(doList))
			for _, do := range doList {
				// TODO(@benqi): check data
				m, _ := messageDOToMessage(&do)
				messages = append(messages, m)
			}
		}
	}

	return
}

////////////////////////////////////////////////////////////////////////////////////////////////////
func GetMessageByPeerAndMessageId(userId int32, messageId int32) (message *mtproto.Message) {
	do := dao.GetMessagesDAO(dao.DB_SLAVE).SelectByMessageId(userId, messageId)
	if do != nil {
		message, _ = messageDOToMessage(do)
	}
	return
}

func GetMessageBoxListByMessageIdList(userId int32, idList []int32) ([]*MessageBox) {
	doList := dao.GetMessagesDAO(dao.DB_SLAVE).SelectByMessageIdList(userId, idList)
	boxList := make([]*MessageBox, 0, len(doList))
	for _, do := range doList {
		message, _ := messageDOToMessage(&do)
		box := &MessageBox{
			UserId:    do.UserId,
			MessageId: do.UserMessageBoxId,
			Message:   message,
		}
		boxList = append(boxList, box)
	}
	return boxList
}

func GetPeerDialogMessageListByMessageId(userId int32, messageId int32) (messages *InboxMessages) {
	doList := dao.GetMessagesDAO(dao.DB_SLAVE).SelectPeerDialogMessageListByMessageId(userId, messageId)
	messages = &InboxMessages{
		UserIds: make([]int32, 0, len(doList)),
		Messages: make([]*mtproto.Message, 0, len(doList)),
	}
	for _, do := range doList {
		// TODO(@benqi): check data
		m, _ := messageDOToMessage(&do)
		messages.Messages = append(messages.Messages, m)
		messages.UserIds = append(messages.UserIds, do.UserId)
	}
	return
}

func GetMessagesByPeerAndMessageIdList2(userId int32, idList []int32) (messages []*mtproto.Message) {
	if len(idList) == 0 {
		messages = []*mtproto.Message{}
	} else {
		doList := dao.GetMessagesDAO(dao.DB_SLAVE).SelectByMessageIdList(userId, idList)
		messages = make([]*mtproto.Message, 0, len(doList))
		for i := 0; i < len(doList); i++ {
			// TODO(@benqi): check data
			m, _ := messageDOToMessage(&doList[i])
			if m != nil {
				messages = append(messages, m)
			}
		}
	}
	return
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// sendMessage
// 所有收件箱信息
type InboxMessages struct {
	UserIds []int32
	Messages []*mtproto.Message
}

//type MessageBox struct {
//	UserId             int32
//	MessageId          int32
//	DialogMessageId    int64
//	// SenderUserId    int32
//	// MessageBoxType  int8
//	// PeerType        int8
//	// PeerId          int32
//	Message            *mtproto.Message
//}

// SendMessage
// 发送到发件箱
func SendMessageToOutbox(fromId int32, peer *base.PeerUtil, clientRandomId int64, message2 *mtproto.Message) (boxId int32, dialogMessageId int64) {
	now := int32(time.Now().Unix())
	messageDO := &dataobject.MessagesDO{
		UserId:fromId,
		UserMessageBoxId: int32(update2.NextMessageBoxId(base2.Int32ToString(fromId))),
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

	// dialog
	// GetDialogModel().CreateOrUpdateByOutbox(fromId, peer.PeerType, peer.PeerId, messageDO.UserMessageBoxId, mentioned)

	boxId = messageDO.UserMessageBoxId
	dialogMessageId = messageDO.DialogMessageId
	return
}

// 发送到收件箱
func SendMessageToInbox(fromId int32, peer *base.PeerUtil, clientRandomId, dialogMessageId int64, outboxMessage *mtproto.Message) (*InboxMessages, error) {
	switch peer.PeerType {
	case base.PEER_USER:
		return sendUserMessageToInbox(fromId, peer, clientRandomId, dialogMessageId, outboxMessage)
	case base.PEER_CHAT:
		return sendChatMessageToInbox(fromId, peer, clientRandomId, dialogMessageId, outboxMessage)
	case base.PEER_CHANNEL:
		return sendChannelMessageToInbox(fromId, peer, clientRandomId, dialogMessageId, outboxMessage)
	default:
		panic("invalid peer")
		return nil, fmt.Errorf("")
	}
}

// 发送到收件箱
func sendUserMessageToInbox(fromId int32, peer *base.PeerUtil, clientRandomId, dialogMessageId int64, outboxMessage *mtproto.Message) (*InboxMessages, error) {
	now := int32(time.Now().Unix())
	messageDO := &dataobject.MessagesDO{
		UserId:peer.PeerId,
		UserMessageBoxId: int32(update2.NextMessageBoxId(base2.Int32ToString(peer.PeerId))),
		DialogMessageId: dialogMessageId,
		SenderUserId: fromId,
		MessageBoxType: MESSAGE_BOX_TYPE_INCOMING,
		PeerType: int8(peer.PeerType),
		PeerId: peer.PeerId,
		RandomId: clientRandomId,
		Date2: now,
		Deleted: 0,
	}

	inboxMessage := proto.Clone(outboxMessage).(*mtproto.Message)
	var mentioned = false

	switch outboxMessage.GetConstructor() {
	case mtproto.TLConstructor_CRC32_messageEmpty:
		messageDO.MessageType = MESSAGE_TYPE_MESSAGE_EMPTY
	case mtproto.TLConstructor_CRC32_message:
		messageDO.MessageType = MESSAGE_TYPE_MESSAGE

		m2 := inboxMessage.To_Message()
		m2.SetOut(false)
		if m2.GetReplyToMsgId() != 0 {
			replyMsgId := GetPeerMessageId(fromId, m2.GetReplyToMsgId(), peer.PeerId)
			m2.SetReplyToMsgId(replyMsgId)
		}
		m2.SetId(messageDO.UserMessageBoxId)
		mentioned = m2.GetMentioned()
	case mtproto.TLConstructor_CRC32_messageService:
		messageDO.MessageType = MESSAGE_TYPE_MESSAGE_SERVICE

		m2 := inboxMessage.To_MessageService()
		m2.SetOut(false)
		m2.SetId(messageDO.UserMessageBoxId)

		mentioned = m2.GetMentioned()
	}

	_ = mentioned

	messageData, _ := json.Marshal(inboxMessage)
	messageDO.MessageData = string(messageData)

	// TODO(@benqi): rpocess clientRandomId dup
	dao.GetMessagesDAO(dao.DB_MASTER).Insert(messageDO)
	// dialog
	// GetDialogModel().CreateOrUpdateByInbox(peer.PeerId, peer.PeerType, fromId, messageDO.UserMessageBoxId, mentioned)

	return &InboxMessages{
		UserIds:  []int32{messageDO.UserId},
		Messages: []*mtproto.Message{inboxMessage},
	}, nil
}

// 发送到收件箱
func sendChatMessageToInbox(fromId int32, peer *base.PeerUtil, clientRandomId, dialogMessageId int64, outboxMessage *mtproto.Message) (*InboxMessages, error) {
	switch outboxMessage.GetConstructor() {
	case mtproto.TLConstructor_CRC32_message:
	case mtproto.TLConstructor_CRC32_messageService:
	default:
		panic("invalid messageEmpty type")
		// return
	}
	return &InboxMessages{
		// UserIds: []int32{peer.PeerId},
		// Messages: []*mtproto.Message{inboxMessage},
	}, nil
}

// 发送到收件箱
func sendChannelMessageToInbox(fromId int32, peer *base.PeerUtil, clientRandomId, dialogMessageId int64, outboxMessage *mtproto.Message) (*InboxMessages, error) {
	switch outboxMessage.GetConstructor() {
	case mtproto.TLConstructor_CRC32_message:
	case mtproto.TLConstructor_CRC32_messageService:
	default:
		panic("invalid messageEmpty type")
		// return
	}
	return &InboxMessages{
		// UserIds: []int32{peer.PeerId},
		// Messages: []*mtproto.Message{inboxMessage},
	}, nil
}

/////////////////////////////////////
func GetMessageIdListByDialog(userId int32, peer *base.PeerUtil) []int32 {
	doList := dao.GetMessagesDAO(dao.DB_SLAVE).SelectDialogMessageIdList(userId, peer.PeerId, int8(peer.PeerType))
	idList := make([]int32, 0, len(doList))
	for i := 0; i < len(doList); i++ {
		idList = append(idList, doList[i].UserMessageBoxId)
	}
	return idList
}

/////////////////////////////////////
func GetPeerMessageId(userId, messageId, peerId int32) int32 {
	do := dao.GetMessagesDAO(dao.DB_SLAVE).SelectPeerMessageId(peerId, userId, messageId)
	if do == nil {
		return 0
	} else {
		return do.UserMessageBoxId
	}
}

func SaveMessage(message *mtproto.Message, userId, messageId int32) error {
	var err error
	messageData, err := json.Marshal(message)
	dao.GetMessagesDAO(dao.DB_MASTER).UpdateMessagesData(string(messageData), userId, messageId)
	return err
}


func DeleteByMessageIdList(userId int32, idList []int32) {
	dao.GetMessagesDAO(dao.DB_MASTER).DeleteMessagesByMessageIdList(userId, idList)
}

/*
// SendMessage
func (m *messageModel) SendMessage(fromId int32, peerType int32, peerId int32, clientRandomId int64, message *mtproto.Message) (ids []*IDMessage) {
	switch peerType {
	case helper.PEER_USER:
		ids = m.sendUserMessage(fromId, peerId, clientRandomId, message)
	case helper.PEER_CHAT:
		ids = m.sendChatMessage(fromId, peerId, clientRandomId, message)
	case helper.PEER_CHANNEL:
		ids = m.sendChannelMessage(fromId, peerId, clientRandomId, message)
	default:
		glog.Errorf("SendMessage - invalid peerType: %d", peerType)
	}

	return
/// *
//	// TODO(@benqi): 重复插入出错处理
//	messageDO := &dataobject.MessagesDO{
//		SenderUserId: fromId,
//		PeerType:     peer.PeerType,
//		PeerId:       peer.PeerId,
//		RandomId:     randomId,
//		Date2:        date,
//	}
//
//	// TODO(@benqi): 测试阶段使用Json!!!
//	// messageDO.MessageData, _ = proto.Marshal(message)
//	messageData, _ := json.Marshal(message)
//	messageDO.MessageData = string(messageData)
//	messageId = int32(dao.GetMessagesDAO(dao.DB_MASTER).Insert(messageDO))
// * /
	// Insert
}

func (m *messageModel) insertMessage(fromId, peerType, peerId int32, clientRandomId int64, message *mtproto.Message) int32 {
	messageDO := &dataobject.MessagesDO{
		// UserId:       userId,
		SenderUserId: fromId,
		PeerType:     helper.PEER_USER,
		PeerId:       peerId,
		RandomId:     clientRandomId,
		Date2:        message.GetData2().GetDate(),
	}

	switch message.GetConstructor() {
	case mtproto.TLConstructor_CRC32_messageEmpty:
		messageDO.MessageType = MESSAGE_TYPE_MESSAGE_EMPTY
	case mtproto.TLConstructor_CRC32_message:
		messageDO.MessageType = MESSAGE_TYPE_MESSAGE
	case mtproto.TLConstructor_CRC32_messageService:
		messageDO.MessageType = MESSAGE_TYPE_MESSAGE_SERVICE
	default:
		panic(fmt.Errorf("Invalid message_type: {%v}", message))
	}

	// TODO(@benqi): 测试阶段使用Json!!!
	// messageDO.MessageData, _ = proto.Marshal(message)
	messageData, _ := json.Marshal(message)
	messageDO.MessageData = string(messageData)

	return int32(dao.GetMessagesDAO(dao.DB_MASTER).Insert(messageDO))
}

// CreateMessage
func (m *messageModel) insertMessageBox(userId, fromId, peerType, peerId int32, messageBoxType int8, messageId int32) (int32) {
	messageBox := &dataobject.MessageBoxesDO{
		UserId:       userId,
		UserMessageBoxId: int32(GetSequenceModel().NextMessageBoxId(base2.Int32ToString(userId))),
		SenderUserId: fromId,
		MessageBoxType: messageBoxType,
		PeerType:     int8(peerType),
		PeerId:       peerId,
		MessageId:    messageId,
		Date2:        int32(time.Now().Unix()),
		CreatedAt:    base2.NowFormatYMDHMS(),
	}

	// TODO(@benqi): check UserMessageBoxId
	if messageBox.UserMessageBoxId == 0 {
		glog.Errorf("insertMessageBox - error: NextMessageBoxId is 0")
	} else {
		dao.GetMessageBoxesDAO(dao.DB_MASTER).Insert(messageBox)
	}

	return messageBox.UserMessageBoxId
}

func (m *messageModel) sendUserMessage(fromId, peerId int32, clientRandomId int64, message *mtproto.Message) (ids []*IDMessage) {
	var boxId int32

	// 存历史消息
	messageId := m.insertMessage(fromId, helper.PEER_USER, peerId, clientRandomId, message)
	if fromId == peerId {
		// PeerSelf

		// 存message_box
		boxId = m.insertMessageBox(fromId, fromId, helper.PEER_USER, peerId, MESSAGE_BOX_TYPE_INCOMING, messageId)
		// dialog
		_  = GetDialogModel().CreateOrUpdateByLastMessage(fromId, helper.PEER_USER, peerId, boxId, message.GetData2().GetMentioned(), true)
		ids = append(ids, &IDMessage{UserId: fromId, MessageBoxId: boxId})
	} else {
		// PeerUser

		// outbox
		// 存历史消息
		// messageId = m.insertMessage(fromId, helper.PEER_USER, peerId, clientRandomId, message)
		// 存message_box
		boxId = m.insertMessageBox(fromId, fromId, helper.PEER_USER, peerId, MESSAGE_BOX_TYPE_OUTGOING, messageId)
		// dialog
		_  = GetDialogModel().CreateOrUpdateByLastMessage(fromId, helper.PEER_USER, peerId, boxId, message.GetData2().GetMentioned(), false)
		ids = append(ids, &IDMessage{UserId: fromId, MessageBoxId: boxId})

		// inbox
		// 存历史消息
		// messageId = m.insertMessage(peerId, fromId, helper.PEER_USER, peerId, clientRandomId, message)
		// 存message_box
		boxId = m.insertMessageBox(peerId, fromId, helper.PEER_USER, peerId, MESSAGE_BOX_TYPE_INCOMING, messageId)
		// dialog
		_  = GetDialogModel().CreateOrUpdateByLastMessage(peerId, helper.PEER_USER, fromId, boxId, message.GetData2().GetMentioned(), true)
		ids = append(ids, &IDMessage{UserId: peerId, MessageBoxId: boxId})
	}
	return
}

func (m *messageModel) sendChatMessage(fromId, peerId int32, clientRandomId int64, message *mtproto.Message) (ids []*IDMessage) {
	return
}

func (m *messageModel) sendChannelMessage(fromId, peerId int32, clientRandomId int64, message *mtproto.Message) (ids []*IDMessage) {
	return
}

func (m *messageModel) GetPeerMessageBoxID(userId, boxID, peerId int32) int32 {
	var id = int32(0)
	do := dao.GetMessageBoxesDAO(dao.DB_SLAVE).SelectByUserIdAndMessageBoxId(userId, boxID)
	if do != nil {
		do2 := dao.GetMessageBoxesDAO(dao.DB_SLAVE).SelectMessageBoxIdByUserIdAndMessageId(peerId, do.MessageId)
		if do2 != nil {
			id = do2.UserMessageBoxId
		}
	}
	return id
}
*/
