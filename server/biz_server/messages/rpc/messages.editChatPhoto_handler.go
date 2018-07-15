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
	"github.com/nebulaim/telegramd/biz/core/chat"
	"github.com/nebulaim/telegramd/biz/base"
	"time"
	"github.com/nebulaim/telegramd/biz/core/message"
	// "github.com/nebulaim/telegramd/biz/core/user"
	"github.com/nebulaim/telegramd/biz/core/user"
	"github.com/nebulaim/telegramd/server/sync/sync_client"
	update2 "github.com/nebulaim/telegramd/biz/core/update"
	"github.com/nebulaim/telegramd/server/nbfs/nbfs_client"
)


/*
	body: { messages_editChatPhoto
	  chat_id: 283362015 [INT],
	  photo: { inputChatUploadedPhoto
		file: { inputFile
		  id: 970161873153942262 [LONG],
		  parts: 9 [INT],
		  name: ".jpg" [STRING],
		  md5_checksum: "f1987132d8a3949420f0130d9e6afe08" [STRING],
		},
	  },
	},
 */
// messages.editChatPhoto#ca4c79d8 chat_id:int photo:InputChatPhoto = Updates;
func (s *MessagesServiceImpl) MessagesEditChatPhoto(ctx context.Context, request *mtproto.TLMessagesEditChatPhoto) (*mtproto.Updates, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("messages.editChatPhoto#ca4c79d8 - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	chatLogic, err := chat.NewChatLogicById(request.ChatId)
	if err != nil {
		glog.Error("messages.editChatTitle#dc452855 - error: ", err)
		return nil, err
	}

	peer := &base.PeerUtil{
		PeerType: base.PEER_CHAT,
		PeerId: chatLogic.GetChatId(),
	}

	var (
		photoId int64 = 0
		action *mtproto.MessageAction
	)

	chatPhoto := request.GetPhoto()
	switch chatPhoto.GetConstructor() {
	case mtproto.TLConstructor_CRC32_inputChatPhotoEmpty:
		photoId = 0
		action = mtproto.NewTLMessageActionChatDeletePhoto().To_MessageAction()
		// chatLogic.EditChatPhoto(md.UserId, photoId)
		// chatLogic.MakeMessageService(md.UserId, action.To_MessageAction())
	case mtproto.TLConstructor_CRC32_inputChatUploadedPhoto:
		file := chatPhoto.GetData2().GetFile()
		// photoId = helper.NextSnowflakeId()
		result, err := nbfs_client.UploadPhotoFile(md.AuthId, file) // photoId, file.GetData2().GetId(), file.GetData2().GetParts(), file.GetData2().GetName(), file.GetData2().GetMd5Checksum())
		if err != nil {
			glog.Errorf("UploadPhoto error: %v", err)
			return nil, err
		}
		photoId = result.PhotoId
		// user.SetUserPhotoID(md.UserId, uuid)
		// fileData := mediaData.GetFile().GetData2()
		photo := &mtproto.TLPhoto{ Data2: &mtproto.Photo_Data{
			Id:          photoId,
			HasStickers: false,
			AccessHash:  result.AccessHash, // photo2.GetFileAccessHash(file.GetData2().GetId(), file.GetData2().GetParts()),
			Date:        int32(time.Now().Unix()),
			Sizes:       result.SizeList,
		}}
		action2 := &mtproto.TLMessageActionChatEditPhoto{Data2: &mtproto.MessageAction_Data{
			Photo: photo.To_Photo(),
		}}
		action = action2.To_MessageAction()
	case mtproto.TLConstructor_CRC32_inputChatPhoto:
		// photo := chatPhoto.GetData2().GetId()
	}

	chatLogic.EditChatPhoto(md.UserId, photoId)
	editChatPhotoMessage := chatLogic.MakeMessageService(md.UserId, action)

	randomId := base.NextSnowflakeId()
	outbox := message.CreateMessageOutboxByNew(md.UserId, peer, randomId, editChatPhotoMessage, func(messageId int32) {
		user.CreateOrUpdateByOutbox(md.UserId, peer.PeerType, peer.PeerId, messageId, false, false)
	})

	syncUpdates := update2.NewUpdatesLogic(md.UserId)
	updateChatParticipants := &mtproto.TLUpdateChatParticipants{Data2: &mtproto.Update_Data{
		Participants: chatLogic.GetChatParticipants().To_ChatParticipants(),
	}}
	syncUpdates.AddUpdate(updateChatParticipants.To_Update())
	syncUpdates.AddUpdateNewMessage(editChatPhotoMessage)
	syncUpdates.AddUsers(user.GetUsersBySelfAndIDList(md.UserId, []int32{md.UserId}))
	syncUpdates.AddChat(chatLogic.ToChat(md.UserId))

	state, _ := sync_client.GetSyncClient().SyncUpdatesData(md.AuthId, md.SessionId, md.UserId, syncUpdates.ToUpdates())
	syncUpdates.PushTopUpdateMessageId(outbox.MessageId, outbox.RandomId)
	syncUpdates.SetupState(state)

	replyUpdates := syncUpdates.ToUpdates()

	inboxList, _ := outbox.InsertMessageToInbox(md.UserId, peer, func(inBoxUserId, messageId int32) {
		user.CreateOrUpdateByInbox(inBoxUserId, base.PEER_CHAT, peer.PeerId, messageId, false)
	})

	for _, inbox := range inboxList {
		updates := update2.NewUpdatesLogic(md.UserId)
		updates.AddUpdate(updateChatParticipants.To_Update())
		updates.AddUpdateNewMessage(inbox.Message)
		updates.AddUsers(user.GetUsersBySelfAndIDList(md.UserId, []int32{md.UserId}))
		updates.AddChat(chatLogic.ToChat(inbox.UserId))
		sync_client.GetSyncClient().PushToUserUpdatesData(inbox.UserId, updates.ToUpdates())
	}

	glog.Infof("messages.editChatPhoto#ca4c79d8 - reply: {%v}", replyUpdates)
	return replyUpdates, nil
}
