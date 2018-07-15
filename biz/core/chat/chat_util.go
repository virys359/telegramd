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
	"github.com/nebulaim/telegramd/proto/mtproto"
	"github.com/golang/glog"
	// photo2 "github.com/nebulaim/telegramd/biz/core/photo"
	"time"
	"github.com/nebulaim/telegramd/biz/base"
	"github.com/nebulaim/telegramd/biz/core/account"
	"github.com/nebulaim/telegramd/server/nbfs/nbfs_client"
)

//func CheckChatAccessHash(id int32, hash int64) bool {
//	return true
//}

// GetUsersBySelfAndIDList
func GetChatListBySelfAndIDList(selfUserId int32, idList []int32) (chats []*mtproto.Chat) {
	if len(idList) == 0 {
		return []*mtproto.Chat{}
	}

	chats = make([]*mtproto.Chat, 0, len(idList))

	// TODO(@benqi): 性能优化，从DB里一次性取出所有的chatList
	for _, id := range idList {
		chatData, err := NewChatLogicById(id)
		if err != nil {
			glog.Error("getChatListBySelfIDList - not find chat_id: ", id)
			chatEmpty := &mtproto.TLChatEmpty{Data2: &mtproto.Chat_Data{
				Id: id,
			}}
			chats = append(chats, chatEmpty.To_Chat())
		} else {
			chats = append(chats, chatData.ToChat(selfUserId))
		}
	}

	return
}

func GetChatBySelfID(selfUserId, chatId int32) (chat *mtproto.Chat) {
	chatData, err := NewChatLogicById(chatId)
	if err != nil {
		glog.Error("getChatBySelfID - not find chat_id: ", chatId)
		chatEmpty := &mtproto.TLChatEmpty{Data2: &mtproto.Chat_Data{
			Id: chatId,
		}}
		chat = chatEmpty.To_Chat()
	} else {
		chat = chatData.ToChat(selfUserId)
	}

	return
}

func GetChatFullBySelfId(selfUserId int32, chatData *chatLogicData) (*mtproto.TLChatFull) {
	sizes, _ := nbfs_client.GetPhotoSizeList(chatData.chat.PhotoId)
	// photo2 := photo2.MakeUserProfilePhoto(photoId, sizes)
	var photo *mtproto.Photo

	if chatData.GetPhotoId() == 0 {
		photoEmpty := &mtproto.TLPhotoEmpty{Data2: &mtproto.Photo_Data{
			Id: 0,
		}}
		photo = photoEmpty.To_Photo()
	} else {
		chatPhoto := &mtproto.TLPhoto{ Data2: &mtproto.Photo_Data{
			Id:          chatData.chat.PhotoId,
			HasStickers: false,
			AccessHash:  chatData.chat.PhotoId, // photo2.GetFileAccessHash(file.GetData2().GetId(), file.GetData2().GetParts()),
			Date:        int32(time.Now().Unix()),
			Sizes:       sizes,
		}}
		photo = chatPhoto.To_Photo()
	}

	peer := &base.PeerUtil{
		PeerType: base.PEER_CHAT,
		PeerId:   chatData.GetChatId(),
	}
	notifySettings := account.GetNotifySettings(selfUserId, peer)

	chatFull := &mtproto.TLChatFull{Data2: &mtproto.ChatFull_Data{
		Id:             chatData.GetChatId(),
		Participants:   chatData.GetChatParticipants().To_ChatParticipants(),
		ChatPhoto:      photo,
		NotifySettings: notifySettings,
		ExportedInvite: mtproto.NewTLChatInviteEmpty().To_ExportedChatInvite(), // TODO(@benqi):
		BotInfo:        []*mtproto.BotInfo{},
	}}

	return chatFull
}
