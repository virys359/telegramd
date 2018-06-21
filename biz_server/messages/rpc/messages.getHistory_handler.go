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
	"github.com/nebulaim/telegramd/baselib/grpc_util"
	"github.com/nebulaim/telegramd/mtproto"
	"golang.org/x/net/context"
	"github.com/nebulaim/telegramd/biz/base"
	"math"
	"github.com/nebulaim/telegramd/biz/core/message"
	"github.com/nebulaim/telegramd/biz/core/user"
	"github.com/nebulaim/telegramd/biz/core/chat"
)

//const (
//	LOAD_HISTORY_TYPE_BACKWARD = 0
//	LOAD_HISTORY_TYPE_FORWARD = 1
//	LOAD_HISTORY_TYPE_FIRST_UNREAD = 2
//	LOAD_HISTORY_TYPE_AROUND_MESSAGE = 3
//	LOAD_HISTORY_TYPE_AROUND_DATE = 4
//)
//
//func calcLoadHistoryType(addOffset, limit int32) int {
//	if addOffset == 0 {
//		return LOAD_HISTORY_TYPE_BACKWARD
//	} else if addOffset == -limit + 5 {
//		return LOAD_HISTORY_TYPE_AROUND_DATE
//	} else if addOffset == -limit / 2 {
//		return LOAD_HISTORY_TYPE_AROUND_MESSAGE
//	} else if addOffset == -limit - 1 {
//		return 	LOAD_HISTORY_TYPE_FORWARD
//	} else if addOffset == -limit + 6 {
//		// TODO(@benqi): 	} else if (load_type == 2 && max_id != 0) {
//		return LOAD_HISTORY_TYPE_FIRST_UNREAD
//	} else {
//		// TODO(@benqi):
//		//if (lower_part < 0 && max_id != 0) {
//		//	TLRPC.Chat chat = getChat(-lower_part);
//		//	if (ChatObject.isChannel(chat)) {
//		//		req.add_offset = -1;
//		//		req.limit += 1;
//		//	}
//		//}
//	}
//	return LOAD_HISTORY_TYPE_BACKWARD
//}

// From android client
//
// load_type == 0 ? backward loading
// load_type == 1 ? forward loading
// load_type == 2 ? load from first unread
// load_type == 3 ? load around message
// load_type == 4 ? load around date
/*
  // @benqi: 这什么鬼规则啊？？？
  1. getHistory, ps: max_id:int min_id:int未使用
	TLRPC.TL_messages_getHistory req = new TLRPC.TL_messages_getHistory();
	req.peer = getInputPeer(lower_part);
	if (load_type == 4) {
		req.add_offset = -count + 5;
	} else if (load_type == 3) {
		req.add_offset = -count / 2;
	} else if (load_type == 1) {
		req.add_offset = -count - 1;
	} else if (load_type == 2 && max_id != 0) {
		req.add_offset = -count + 6;
	} else {
		if (lower_part < 0 && max_id != 0) {
			TLRPC.Chat chat = getChat(-lower_part);
			if (ChatObject.isChannel(chat)) {
				req.add_offset = -1;
				req.limit += 1;
			}
		}
	}
	req.limit = count;
	req.offset_id = max_id;
	req.offset_date = offset_date;

  2. Load dialog last message, ps: limit = 1
	TLRPC.TL_messages_getHistory req = new TLRPC.TL_messages_getHistory();
	req.peer = peer == null ? getInputPeer(lower_id) : peer;
	if (req.peer == null) {
		return;
	}
	req.limit = 1;

 */

// request: {"peer":{"constructor":2072935910,"data2":{"user_id":5,"access_hash":1006843769775067136}},"offset_id":1,"add_offset":-25,"limit":50}
// request: {"peer":{"constructor":2072935910,"data2":{"user_id":4,"access_hash":405858233924775823}},"offset_id":2147483647,"offset_date":2147483647,"limit":1,"max_id":2147483647,"min_id":1}
// request: {"peer":{"constructor":2072935910,"data2":{"user_id":4,"access_hash":405858233924775823}},"offset_id":2147483647,"offset_date":2147483647,"limit":1,"max_id":2147483647,"min_id":1}
// messages.getHistory#dcbb8260 peer:InputPeer offset_id:int offset_date:int add_offset:int limit:int max_id:int min_id:int hash:int = messages.Messages;
func (s *MessagesServiceImpl) MessagesGetHistory(ctx context.Context, request *mtproto.TLMessagesGetHistory) (*mtproto.Messages_Messages, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("messages.getHistory#dcbb8260 - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	peer := base.FromInputPeer(request.GetPeer())
	if peer.PeerType == base.PEER_SELF {
		peer.PeerType = base.PEER_USER
		peer.PeerId = md.UserId
	}

	// chatIdList := []int32{}
	// userIdList := []int32{md.UserId}

	offsetId := request.GetOffsetId()
	addOffset := request.GetAddOffset()
	limit := request.GetLimit()

	var messagesMessages *mtproto.Messages_Messages

	if peer.PeerType == base.PEER_CHANNEL {
		messages := []*mtproto.Message{}

		if limit == 1 {
			// 1. Load dialog last messag
			offsetId = math.MaxInt32
			messages = message.LoadBackwardHistoryMessages(md.UserId, peer.PeerType, peer.PeerId, offsetId, limit)
		} else {
			if addOffset < 0 {
				if addOffset + limit <= 0 {
					// LOAD_HISTORY_TYPE_FORWARD
					// Forward是按升序排
					messages = message.LoadForwardHistoryMessages(md.UserId, peer.PeerType, peer.PeerId, offsetId, -addOffset)
				} else {
					// LOAD_HISTORY_TYPE_FORWARD and LOAD_HISTORY_TYPE_BACKWARD
					// 按升序排
					messages1 := message.LoadForwardHistoryMessages(md.UserId, peer.PeerType, peer.PeerId, offsetId, -addOffset)
					messages = append(messages, messages1...)
					// 降序
					messages2 := message.LoadBackwardHistoryMessages(md.UserId, peer.PeerType, peer.PeerId, offsetId, limit + addOffset)
					messages = append(messages, messages2...)

					// @benqi: why??????
					if addOffset == -limit / 2 {
						for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
							messages[i], messages[j] = messages[j], messages[i]
						}
					}
				}
			} else {
				// 降序
				messages = message.LoadBackwardHistoryMessages(md.UserId, peer.PeerType, peer.PeerId, offsetId, addOffset + limit)
			}
			//// 2. getHistory
			//loadType := calcLoadHistoryType(addOffset, limit)
			//switch loadType {
			//case LOAD_HISTORY_TYPE_BACKWARD:
			//	messages = model.GetMessageModel().LoadBackwardHistoryMessages(md.UserId, peer.PeerType, peer.PeerId, offsetId, limit)
			//case LOAD_HISTORY_TYPE_FORWARD:
			//	// TODO(@benqi): 可能有问题，可能要按limit以及addOffset全部取出然后排除掉多余的offset
			//	// Forward是按升序排
			//	messages = model.GetMessageModel().LoadForwardHistoryMessages(md.UserId, peer.PeerType, peer.PeerId, offsetId, limit)
			//
			//case LOAD_HISTORY_TYPE_FIRST_UNREAD:
			//	// TODO(@benqi): 暂不实现
			//case LOAD_HISTORY_TYPE_AROUND_MESSAGE:
			//	// 按升序排
			//	messages1 := model.GetMessageModel().LoadForwardHistoryMessages(md.UserId, peer.PeerType, peer.PeerId, offsetId, limit/2)
			//	messages = append(messages, messages1...)
			//	// 降序
			//	messages2 := model.GetMessageModel().LoadBackwardHistoryMessages(md.UserId, peer.PeerType, peer.PeerId, offsetId, limit/2)
			//	messages = append(messages, messages2...)
			//case LOAD_HISTORY_TYPE_AROUND_DATE:
			//	// TODO(@benqi): 暂不实现
			//}
		}

		// TODO(@benqi): 查询出来超过limit条记录是否要处理？
		// messages = model.GetMessageModel().LoadBackwardHistoryMessages(md.UserId, peer.PeerType, peer.PeerId, request.GetOffsetId(), request.GetLimit())
		//for _, message := range messages {
		//	switch message.GetConstructor() {
		//	case mtproto.TLConstructor_CRC32_message:
		//		m := message.To_Message()
		//		userIdList = append(userIdList, m.GetFromId())
		//		p := helper.FromPeer(m.GetToId())
		//		switch p.PeerType {
		//		case helper.PEER_SELF, helper.PEER_USER:
		//			userIdList = append(userIdList, p.PeerId)
		//		case helper.PEER_CHAT:
		//			chatIdList = append(chatIdList, p.PeerId)
		//		case helper.PEER_CHANNEL:
		//			// TODO(@benqi): add channel
		//		}
		//	case mtproto.TLConstructor_CRC32_messageService:
		//		m := message.To_MessageService()
		//		userIdList = append(userIdList, m.GetFromId())
		//		chatIdList = append(chatIdList, m.GetToId().GetData2().GetChatId())
		//	}
		//}

		// messagesMessages := mtproto.NewTLMessagesMessages()
		if len(messages) == int(request.GetLimit()) {
			messaegesSlice := mtproto.NewTLMessagesMessagesSlice()
			messaegesSlice.SetCount(request.GetLimit())
			messaegesSlice.SetMessages(messages)
			//request.GetLimit())
			messagesMessages = messaegesSlice.To_Messages_Messages()
		} else {
			messages3 := mtproto.NewTLMessagesMessages()
			messages3.SetMessages(messages)
			messagesMessages = messages3.To_Messages_Messages()
		}
		// messagesMessages.SetMessages(messages)
		userIdList, chatIdList, _ := message.PickAllIDListByMessages(messages)
		if len(userIdList) > 0 {
			users := user.GetUsersBySelfAndIDList(md.UserId, userIdList)
			messagesMessages.Data2.Users = users
			//for _, u := range users {
			//	if u.GetId() == md.UserId {
			//		u.SetSelf(true)
			//	}
			//	u.SetContact(true)
			//	messagesMessages.Data2.Users = append(messagesMessages.Data2.Users, u.To_User())
			//}
		}

		if len(chatIdList) > 0 {
			messagesMessages.Data2.Chats = chat.GetChatListBySelfAndIDList(md.UserId, chatIdList)
		}
	} else {
		messages := []*mtproto.Message{}

		if limit == 1 {
			// 1. Load dialog last messag
			offsetId = math.MaxInt32
			messages = message.LoadBackwardHistoryMessages(md.UserId, peer.PeerType, peer.PeerId, offsetId, limit)
		} else {
			if addOffset < 0 {
				if addOffset + limit <= 0 {
					// LOAD_HISTORY_TYPE_FORWARD
					// Forward是按升序排
					messages = message.LoadForwardHistoryMessages(md.UserId, peer.PeerType, peer.PeerId, offsetId, -addOffset)
				} else {
					// LOAD_HISTORY_TYPE_FORWARD and LOAD_HISTORY_TYPE_BACKWARD
					// 按升序排
					messages1 := message.LoadForwardHistoryMessages(md.UserId, peer.PeerType, peer.PeerId, offsetId, -addOffset)
					messages = append(messages, messages1...)
					// 降序
					messages2 := message.LoadBackwardHistoryMessages(md.UserId, peer.PeerType, peer.PeerId, offsetId, limit + addOffset)
					messages = append(messages, messages2...)

					// @benqi: why??????
					if addOffset == -limit / 2 {
						for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
							messages[i], messages[j] = messages[j], messages[i]
						}
					}
				}
			} else {
				// 降序
				messages = message.LoadBackwardHistoryMessages(md.UserId, peer.PeerType, peer.PeerId, offsetId, addOffset + limit)
			}
			//// 2. getHistory
			//loadType := calcLoadHistoryType(addOffset, limit)
			//switch loadType {
			//case LOAD_HISTORY_TYPE_BACKWARD:
			//	messages = model.GetMessageModel().LoadBackwardHistoryMessages(md.UserId, peer.PeerType, peer.PeerId, offsetId, limit)
			//case LOAD_HISTORY_TYPE_FORWARD:
			//	// TODO(@benqi): 可能有问题，可能要按limit以及addOffset全部取出然后排除掉多余的offset
			//	// Forward是按升序排
			//	messages = model.GetMessageModel().LoadForwardHistoryMessages(md.UserId, peer.PeerType, peer.PeerId, offsetId, limit)
			//
			//case LOAD_HISTORY_TYPE_FIRST_UNREAD:
			//	// TODO(@benqi): 暂不实现
			//case LOAD_HISTORY_TYPE_AROUND_MESSAGE:
			//	// 按升序排
			//	messages1 := model.GetMessageModel().LoadForwardHistoryMessages(md.UserId, peer.PeerType, peer.PeerId, offsetId, limit/2)
			//	messages = append(messages, messages1...)
			//	// 降序
			//	messages2 := model.GetMessageModel().LoadBackwardHistoryMessages(md.UserId, peer.PeerType, peer.PeerId, offsetId, limit/2)
			//	messages = append(messages, messages2...)
			//case LOAD_HISTORY_TYPE_AROUND_DATE:
			//	// TODO(@benqi): 暂不实现
			//}
		}

		// TODO(@benqi): 查询出来超过limit条记录是否要处理？
		// messages = model.GetMessageModel().LoadBackwardHistoryMessages(md.UserId, peer.PeerType, peer.PeerId, request.GetOffsetId(), request.GetLimit())
		//for _, message := range messages {
		//	switch message.GetConstructor() {
		//	case mtproto.TLConstructor_CRC32_message:
		//		m := message.To_Message()
		//		userIdList = append(userIdList, m.GetFromId())
		//		p := helper.FromPeer(m.GetToId())
		//		switch p.PeerType {
		//		case helper.PEER_SELF, helper.PEER_USER:
		//			userIdList = append(userIdList, p.PeerId)
		//		case helper.PEER_CHAT:
		//			chatIdList = append(chatIdList, p.PeerId)
		//		case helper.PEER_CHANNEL:
		//			// TODO(@benqi): add channel
		//		}
		//	case mtproto.TLConstructor_CRC32_messageService:
		//		m := message.To_MessageService()
		//		userIdList = append(userIdList, m.GetFromId())
		//		chatIdList = append(chatIdList, m.GetToId().GetData2().GetChatId())
		//	}
		//}

		// messagesMessages := mtproto.NewTLMessagesMessages()
		if len(messages) == int(request.GetLimit()) {
			messaegesSlice := mtproto.NewTLMessagesMessagesSlice()
			messaegesSlice.SetCount(request.GetLimit())
			messaegesSlice.SetMessages(messages)
			//request.GetLimit())
			messagesMessages = messaegesSlice.To_Messages_Messages()
		} else {
			messages3 := mtproto.NewTLMessagesMessages()
			messages3.SetMessages(messages)
			messagesMessages = messages3.To_Messages_Messages()
		}
		// messagesMessages.SetMessages(messages)
		userIdList, chatIdList, _ := message.PickAllIDListByMessages(messages)
		if len(userIdList) > 0 {
			users := user.GetUsersBySelfAndIDList(md.UserId, userIdList)
			messagesMessages.Data2.Users = users
			//for _, u := range users {
			//	if u.GetId() == md.UserId {
			//		u.SetSelf(true)
			//	}
			//	u.SetContact(true)
			//	messagesMessages.Data2.Users = append(messagesMessages.Data2.Users, u.To_User())
			//}
		}

		if len(chatIdList) > 0 {
			messagesMessages.Data2.Chats = chat.GetChatListBySelfAndIDList(md.UserId, chatIdList)
		}
	}
	glog.Infof("messages.getHistory#dcbb8260 - reply: %s", logger.JsonDebugData(messagesMessages))
	return messagesMessages, nil
}
