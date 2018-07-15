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
	"github.com/nebulaim/telegramd/biz/core/message"
	"github.com/nebulaim/telegramd/biz/core/user"
	"github.com/nebulaim/telegramd/biz/core/chat"
)

// messages.getMessages#4222fa74 id:Vector<int> = messages.Messages;
func (s *MessagesServiceImpl) MessagesGetMessages(ctx context.Context, request *mtproto.TLMessagesGetMessages) (*mtproto.Messages_Messages, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("messages.getMessages#4222fa74 - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	messages :=message.GetMessagesByPeerAndMessageIdList2(md.UserId, request.Id)
	userIdList, chatIdList, _ := message.PickAllIDListByMessages(messages)
	userList := user.GetUsersBySelfAndIDList(md.UserId, userIdList)
	chatList := chat.GetChatListBySelfAndIDList(md.UserId, chatIdList)

	messagesMessages := &mtproto.TLMessagesMessages{Data2: &mtproto.Messages_Messages_Data{
		Messages: messages,
		Users: userList,
		Chats: chatList,
	}}

	glog.Infof("messages.getMessages#4222fa74 - reply: %s", logger.JsonDebugData(messagesMessages))
	return messagesMessages.To_Messages_Messages(), nil
}
