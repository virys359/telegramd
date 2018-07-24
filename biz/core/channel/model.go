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
	"github.com/golang/glog"
	// "github.com/nebulaim/telegramd/server/nbfs/nbfs_client"
	"github.com/nebulaim/telegramd/biz/base"
	"github.com/nebulaim/telegramd/biz/core"
	"github.com/nebulaim/telegramd/biz/dal/dao"
	"github.com/nebulaim/telegramd/biz/dal/dao/mysql_dao"
	"github.com/nebulaim/telegramd/proto/mtproto"
	// "github.com/nebulaim/telegramd/server/nbfs/nbfs_client"
)

type channelsDAO struct {
	*mysql_dao.CommonDAO
	*mysql_dao.UsersDAO
	*mysql_dao.ChannelsDAO
	*mysql_dao.ChannelParticipantsDAO
}

type ChannelModel struct {
	dao           *channelsDAO
	photoCallback core.PhotoCallback
	notifySetting core.NotifySettingCallback
}

func (m *ChannelModel) InstallModel() {
	m.dao.CommonDAO = dao.GetCommonDAO(dao.DB_MASTER)
	m.dao.UsersDAO = dao.GetUsersDAO(dao.DB_MASTER)
	m.dao.ChannelsDAO = dao.GetChannelsDAO(dao.DB_MASTER)
	m.dao.ChannelParticipantsDAO = dao.GetChannelParticipantsDAO(dao.DB_MASTER)
}

func (m *ChannelModel) RegisterCallback(cb interface{}) {
	switch cb.(type) {
	case core.PhotoCallback:
		glog.Info("chatModel - register core.PhotoCallback")
		m.photoCallback = cb.(core.PhotoCallback)
	case core.NotifySettingCallback:
		glog.Info("chatModel - register core.NotifySettingCallback")
		m.notifySetting = cb.(core.NotifySettingCallback)
	}
}

// GetUsersBySelfAndIDList
func (m *ChannelModel) GetChannelListBySelfAndIDList(selfUserId int32, idList []int32) (chats []*mtproto.Chat) {
	if len(idList) == 0 {
		return []*mtproto.Chat{}
	}

	chats = make([]*mtproto.Chat, 0, len(idList))

	// TODO(@benqi): 性能优化，从DB里一次性取出所有的chatList
	for _, id := range idList {
		chatData, err := m.NewChannelLogicById(id)
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

func (m *ChannelModel) CheckChannelUserName(id int32, userName string) bool {
	do := m.dao.UsersDAO.SelectByUsername(userName)
	if do == nil {
		return false
	}

	params := map[string]interface{}{
		"channel_id": id,
		"user_id":    do.Id,
	}
	return m.dao.CommonDAO.CheckExists("channel_participants", params)
}

func (m *ChannelModel) GetChannelBySelfID(selfUserId, channelId int32) (chat *mtproto.Chat) {
	channelData, err := m.NewChannelLogicById(channelId)
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

func (m *ChannelModel) GetChannelParticipantIdList(channelId int32) []int32 {
	doList := m.dao.ChannelParticipantsDAO.SelectByChannelId(channelId)
	idList := make([]int32, 0, len(doList))
	for i := 0; i < len(doList); i++ {
		if doList[i].State == 0 {
			idList = append(idList, doList[i].UserId)
		}
	}
	return idList
}

func (m *ChannelModel) GetChannelFullBySelfId(selfUserId int32, channelData *channelLogicData) *mtproto.TLChannelFull {
	// TODO(@benqi): nbfs_client
	var photo *mtproto.Photo

	// TODO(@benqi):
	if channelData.GetPhotoId() == 0 {
		photoEmpty := &mtproto.TLPhotoEmpty{Data2: &mtproto.Photo_Data{
			Id: 0,
		}}
		photo = photoEmpty.To_Photo()
	} else {
		// sizes, _ := nbfs_client.GetPhotoSizeList(channelData.channel.PhotoId)
		// photo2 := photo2.MakeUserProfilePhoto(photoId, sizes)
		//channelPhoto := &mtproto.TLPhoto{ Data2: &mtproto.Photo_Data{
		//	Id:          channelData.channel.PhotoId,
		//	HasStickers: false,
		//	AccessHash:  channelData.channel.PhotoId, // photo2.GetFileAccessHash(file.GetData2().GetId(), file.GetData2().GetParts()),
		//	Date:        int32(time.Now().Unix()),
		//	Sizes:       sizes,
		//}}
		photo = m.photoCallback.GetPhoto(channelData.channel.PhotoId)
		// channelPhoto.To_Photo()
	}

	peer := &base.PeerUtil{
		PeerType: base.PEER_CHANNEL,
		PeerId:   channelData.GetChannelId(),
	}

	// var notifySettings *mtproto.PeerNotifySettings

	// TODO(@benqi): chat notifySetting...
	//if notifySettingFunc == nil {
	//	notifySettings = &mtproto.PeerNotifySettings{
	//		Constructor: mtproto.TLConstructor_CRC32_peerNotifySettings,
	//		Data2: &mtproto.PeerNotifySettings_Data{
	//			ShowPreviews: true,
	//			Silent:       false,
	//			MuteUntil:    0,
	//			Sound:        "default",
	//		},
	//	}
	//} else {
	notifySettings := m.notifySetting.GetNotifySettings(selfUserId, peer)
	//}

	channelFull := &mtproto.TLChannelFull{Data2: &mtproto.ChatFull_Data{
		// CanViewParticipants:
		Id:                channelData.channel.Id,
		About:             channelData.channel.Title,
		ParticipantsCount: channelData.channel.ParticipantCount,
		AdminsCount:       1, // TODO(@benqi): calc adminscount
		ChatPhoto:         photo,
		NotifySettings:    notifySettings,
		// ExportedInvite:    mtproto.NewTLChatInviteEmpty().To_ExportedChatInvite(), // TODO(@benqi):
		BotInfo: []*mtproto.BotInfo{},
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

func init() {
	core.RegisterCoreModel(&ChannelModel{dao: &channelsDAO{}})
}
