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

package channel

import (
	"github.com/nebulaim/telegramd/biz/dal/dao"
	"github.com/nebulaim/telegramd/server/nbfs/nbfs_client"
	"github.com/nebulaim/telegramd/biz/core/account"
	"github.com/nebulaim/telegramd/proto/mtproto"
	"time"
	"github.com/nebulaim/telegramd/biz/base"
	"github.com/golang/glog"
)

// GetUsersBySelfAndIDList
func GetChannelListBySelfAndIDList(selfUserId int32, idList []int32) (chats []*mtproto.Chat) {
	if len(idList) == 0 {
		return []*mtproto.Chat{}
	}

	chats = make([]*mtproto.Chat, 0, len(idList))

	// TODO(@benqi): 性能优化，从DB里一次性取出所有的chatList
	for _, id := range idList {
		chatData, err := NewChannelLogicById(id)
		if err != nil {
			glog.Error("getChatListBySelfIDList - not find chat_id: ", id)
			chatEmpty := &mtproto.TLChatEmpty{Data2: &mtproto.Chat_Data{
				Id: id,
			}}
			chats = append(chats, chatEmpty.To_Chat())
		} else {
			chats = append(chats, chatData.ToChannel(selfUserId))
		}
	}

	return
}

func CheckChannelUserName(id int32, userName string) bool {
	do := dao.GetUsersDAO(dao.DB_SLAVE).SelectByUsername(userName)
	if do == nil {
		return false
	}

	params := map[string]interface{}{
		"channel_id": id,
		"user_id": do.Id,
	}
	return dao.GetCommonDAO(dao.DB_SLAVE).CheckExists("channel_participants", params)
}

func GetChannelBySelfID(selfUserId, channelId int32) (chat *mtproto.Chat) {
	channelData, err := NewChannelLogicById(channelId)
	if err != nil {
		glog.Error("GetChannelBySelfID - not find chat_id: ", channelId)
		channelEmpty := &mtproto.TLChatEmpty{Data2: &mtproto.Chat_Data{
			Id: channelId,
		}}
		chat = channelEmpty.To_Chat()
	} else {
		chat = channelData.ToChannel(selfUserId)
	}

	return
}

func GetChannelParticipantIdList(channelId int32) []int32 {
	doList := dao.GetChannelParticipantsDAO(dao.DB_SLAVE).SelectByChannelId(channelId)
	idList := make([]int32, 0, len(doList))
	for i := 0; i < len(doList); i++  {
		if doList[i].State == 0 {
			idList = append(idList, doList[i].UserId)
		}
	}
	return idList
}

func GetChannelFullBySelfId(selfUserId int32, channelData *channelLogicData) (*mtproto.TLChannelFull) {
	sizes, _ := nbfs_client.GetPhotoSizeList(channelData.channel.PhotoId)
	// photo2 := photo2.MakeUserProfilePhoto(photoId, sizes)
	var photo *mtproto.Photo

	if channelData.GetPhotoId() == 0 {
		photoEmpty := &mtproto.TLPhotoEmpty{Data2: &mtproto.Photo_Data{
			Id: 0,
		}}
		photo = photoEmpty.To_Photo()
	} else {
		channelPhoto := &mtproto.TLPhoto{ Data2: &mtproto.Photo_Data{
			Id:          channelData.channel.PhotoId,
			HasStickers: false,
			AccessHash:  channelData.channel.PhotoId, // photo2.GetFileAccessHash(file.GetData2().GetId(), file.GetData2().GetParts()),
			Date:        int32(time.Now().Unix()),
			Sizes:       sizes,
		}}
		photo = channelPhoto.To_Photo()
	}

	peer := &base.PeerUtil{
		PeerType: base.PEER_CHANNEL,
		PeerId:   channelData.GetChannelId(),
	}

	// TODO(@benqi): chat notifySetting...
	notifySettings := account.GetNotifySettings(selfUserId, peer)

	channelFull := &mtproto.TLChannelFull{Data2: &mtproto.ChatFull_Data{
		// CanViewParticipants:
		Id:                channelData.channel.Id,
		About:             channelData.channel.Title,
		ParticipantsCount: channelData.channel.ParticipantCount,
		AdminsCount:       1,	// TODO(@benqi): calc adminscount
		ChatPhoto:         photo,
		NotifySettings:    notifySettings,
		// ExportedInvite:    mtproto.NewTLChatInviteEmpty().To_ExportedChatInvite(), // TODO(@benqi):
		BotInfo:           []*mtproto.BotInfo{},
	}}

	isAdmin := channelData.checkUserIsAdministrator(selfUserId)
	if isAdmin {
		channelFull.SetCanViewParticipants(true)
		channelFull.SetCanSetUsername(true)
	}

	exportedInvite := &mtproto.TLChatInviteExported{Data2: &mtproto.ExportedChatInvite_Data{
		Link: channelData.channel.Link,
	}}

	channelFull.SetExportedInvite(exportedInvite.To_ExportedChatInvite())
	return channelFull
}
