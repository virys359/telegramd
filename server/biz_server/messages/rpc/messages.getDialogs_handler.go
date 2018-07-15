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
	"github.com/nebulaim/telegramd/proto/mtproto"
	"golang.org/x/net/context"
	"math"
	"github.com/nebulaim/telegramd/biz/core/user"
	"github.com/nebulaim/telegramd/biz/core/message"
	"github.com/nebulaim/telegramd/biz/core/chat"
	"github.com/nebulaim/telegramd/biz/core/channel"
	"github.com/nebulaim/telegramd/biz/core/dialog"
)

// android client request source code
/*
	TLRPC.TL_messages_getDialogs req = new TLRPC.TL_messages_getDialogs();
	req.limit = count;
	req.exclude_pinned = true;
	if (UserConfig.dialogsLoadOffsetId != -1) {
		if (UserConfig.dialogsLoadOffsetId == Integer.MAX_VALUE) {
			dialogsEndReached = true;
			serverDialogsEndReached = true;
			loadingDialogs = false;
			NotificationCenter.getInstance().postNotificationName(NotificationCenter.dialogsNeedReload);
			return;
		}
		req.offset_id = UserConfig.dialogsLoadOffsetId;
		req.offset_date = UserConfig.dialogsLoadOffsetDate;
		if (req.offset_id == 0) {
			req.offset_peer = new TLRPC.TL_inputPeerEmpty();
		} else {
			if (UserConfig.dialogsLoadOffsetChannelId != 0) {
				req.offset_peer = new TLRPC.TL_inputPeerChannel();
				req.offset_peer.channel_id = UserConfig.dialogsLoadOffsetChannelId;
			} else if (UserConfig.dialogsLoadOffsetUserId != 0) {
				req.offset_peer = new TLRPC.TL_inputPeerUser();
				req.offset_peer.user_id = UserConfig.dialogsLoadOffsetUserId;
			} else {
				req.offset_peer = new TLRPC.TL_inputPeerChat();
				req.offset_peer.chat_id = UserConfig.dialogsLoadOffsetChatId;
			}
			req.offset_peer.access_hash = UserConfig.dialogsLoadOffsetAccess;
		}
	} else {
		boolean found = false;
		for (int a = dialogs.size() - 1; a >= 0; a--) {
			TLRPC.TL_dialog dialog = dialogs.get(a);
			if (dialog.pinned) {
				continue;
			}
			int lower_id = (int) dialog.id;
			int high_id = (int) (dialog.id >> 32);
			if (lower_id != 0 && high_id != 1 && dialog.top_message > 0) {
				MessageObject message = dialogMessage.get(dialog.id);
				if (message != null && message.getId() > 0) {
					req.offset_date = message.messageOwner.date;
					req.offset_id = message.messageOwner.id;
					int id;
					if (message.messageOwner.to_id.channel_id != 0) {
						id = -message.messageOwner.to_id.channel_id;
					} else if (message.messageOwner.to_id.chat_id != 0) {
						id = -message.messageOwner.to_id.chat_id;
					} else {
						id = message.messageOwner.to_id.user_id;
					}
					req.offset_peer = getInputPeer(id);
					found = true;
					break;
				}
			}
		}
		if (!found) {
			req.offset_peer = new TLRPC.TL_inputPeerEmpty();
		}
	}
 */
// 由客户端代码: offset_id为当前用户最后一条消息ID，offset_peer为最后一条消息的接收者peer
// offset_date
// messages.getDialogs#191ba9c5 flags:# exclude_pinned:flags.0?true offset_date:int offset_id:int offset_peer:InputPeer limit:int = messages.Dialogs;
func (s *MessagesServiceImpl) MessagesGetDialogs(ctx context.Context, request *mtproto.TLMessagesGetDialogs) (*mtproto.Messages_Dialogs, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("messages.getDialogs#191ba9c5 - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	offsetId := request.GetOffsetId()
	if offsetId == 0 {
		offsetId = math.MaxInt32
	}

	/*
		peer := helper.FromInputPeer(request.OffsetPeer)
		if peer.PeerType == helper.PEER_EMPTY {
			// 取出全部
		} else {
			// 通过message_boxs表检查offset_peer
			offsetMessageDO := dao.GetMessageBoxesDAO(dao.DB_SLAVE).SelectByUserIdAndMessageBoxId(md.UserId, offsetId)
			// TODO(@benqi): date, access_hash check
			if offsetMessageDO == nil || ( peer.PeerType != int32(offsetMessageDO.PeerType)  && peer.PeerId != offsetMessageDO.PeerId) {
				panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_BAD_REQUEST), "InputPeer invalid"))
			}
		}
	 */

	// dialogs := user.GetDialogsByOffsetId(md.UserId, !request.GetExcludePinned(), offsetId, request.GetLimit())
	dialogs := user.GetDialogsByOffsetId(md.UserId, false, offsetId, request.GetLimit())
	// glog.Infof("dialogs - {%v}", dialogs)

	// messageIdList, userIdList, chatIdList, channelIdList
	dialogItems := dialog.PickAllIDListByDialogs2(dialogs)

	messages := message.GetMessagesByPeerAndMessageIdList2(md.UserId, dialogItems.MessageIdList)
	for k, v := range dialogItems.ChannelMessageIdMap {
		m := message.GetChannelMessage(k, v)
		if m != nil {
			messages = append(messages, m)
		}
	}

	users := user.GetUsersBySelfAndIDList(md.UserId, dialogItems.UserIdList)
	chats := chat.GetChatListBySelfAndIDList(md.UserId, dialogItems.ChatIdList)
	chats = append(chats, channel.GetChannelListBySelfAndIDList(md.UserId, dialogItems.ChannelIdList)...)

	messageDialogs := mtproto.TLMessagesDialogs{Data2: &mtproto.Messages_Dialogs_Data{
		Dialogs:  dialogs,
		Messages: messages,
		Users:    users,
		Chats:    chats,
	}}

	glog.Infof("messages.getDialogs#191ba9c5 - reply: %s", logger.JsonDebugData(messageDialogs))
	return messageDialogs.To_Messages_Dialogs(), nil
}
