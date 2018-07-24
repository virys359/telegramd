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
	message2 "github.com/nebulaim/telegramd/biz/core/message"
	"github.com/nebulaim/telegramd/proto/mtproto"
	"github.com/nebulaim/telegramd/server/sync/sync_client"
	"golang.org/x/net/context"
	"time"
)

func (s *MessagesServiceImpl) makeUpdateEditMessageUpdates(selfUserId int32, message *mtproto.Message) *mtproto.TLUpdates {
	userIdList, _, _ := message2.PickAllIDListByMessages([]*mtproto.Message{message})
	userList := s.UserModel.GetUsersBySelfAndIDList(selfUserId, userIdList)

	updateNew := &mtproto.TLUpdateEditMessage{Data2: &mtproto.Update_Data{
		Message_1: message,
	}}
	return &mtproto.TLUpdates{Data2: &mtproto.Updates_Data{
		Updates: []*mtproto.Update{updateNew.To_Update()},
		Users:   userList,
		Date:    int32(time.Now().Unix()),
		Seq:     0,
	}}
}

func setEditMessageData(request *mtproto.TLMessagesEditMessage, message *mtproto.Message) {
	// edit message data
	data2 := message.GetData2()
	if request.GetMessage() != "" {
		data2.Message = request.GetMessage()
	}
	if request.GetReplyMarkup() != nil {
		data2.ReplyMarkup = request.GetReplyMarkup()
	}
	if request.GetEntities() != nil {
		data2.Entities = request.GetEntities()
	}
	data2.EditDate = int32(time.Now().Unix())
}

// messages.editMessage#5d1b8dd flags:# no_webpage:flags.1?true stop_geo_live:flags.12?true peer:InputPeer id:int message:flags.11?string reply_markup:flags.2?ReplyMarkup entities:flags.3?Vector<MessageEntity> geo_point:flags.13?InputGeoPoint = Updates;
func (s *MessagesServiceImpl) MessagesEditMessage(ctx context.Context, request *mtproto.TLMessagesEditMessage) (*mtproto.Updates, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("messages.editMessage#5d1b8dd - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	// SelectDialogMessageListByMessageId
	editOutbox := s.MessageModel.GetMessageByPeerAndMessageId(md.UserId, request.GetId())
	// TODO(@benqi): check invalid

	setEditMessageData(request, editOutbox)

	syncUpdates := s.makeUpdateEditMessageUpdates(md.UserId, editOutbox)
	state, err := sync_client.GetSyncClient().SyncUpdatesData(md.AuthId, md.SessionId, md.UserId, syncUpdates.To_Updates())
	if err != nil {
		return nil, err
	}
	SetupUpdatesState(state, syncUpdates)
	s.MessageModel.SaveMessage(editOutbox, md.UserId, request.GetId())

	// push edit peer message
	peerEditMessages := s.MessageModel.GetPeerDialogMessageListByMessageId(md.UserId, request.GetId())
	for i := 0; i < len(peerEditMessages.UserIds); i++ {
		editMessage := peerEditMessages.Messages[i]
		editUserId := peerEditMessages.UserIds[i]

		setEditMessageData(request, editMessage)
		editUpdates := s.makeUpdateEditMessageUpdates(editUserId, editMessage)
		sync_client.GetSyncClient().PushToUserUpdatesData(editUserId, editUpdates.To_Updates())
		s.MessageModel.SaveMessage(editMessage, editUserId, editMessage.GetData2().GetId())
	}

	glog.Infof("messages.editMessage#5d1b8dd - reply: %s", logger.JsonDebugData(syncUpdates))
	return syncUpdates.To_Updates(), nil
}
