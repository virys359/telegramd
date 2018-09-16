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
	"fmt"
	base2 "github.com/nebulaim/telegramd/baselib/base"
	"github.com/nebulaim/telegramd/biz/base"
	"github.com/nebulaim/telegramd/biz/dal/dataobject"
	"github.com/nebulaim/telegramd/proto/mtproto"
	"math/rand"
	"time"
	// "github.com/nebulaim/telegramd/server/nbfs/nbfs_client"
	"encoding/base64"
	"github.com/nebulaim/telegramd/baselib/crypto"
	// photo2 "github.com/nebulaim/telegramd/biz/core/photo"
	// "github.com/nebulaim/telegramd/server/nbfs/nbfs_client"
	"github.com/nebulaim/telegramd/biz/core"
)

const (
	kChannelParticipant        = 0
	kChannelParticipantSelf    = 1
	kChannelParticipantCreator = 2
	kChannelParticipantAdmin   = 3
	kChannelParticipantBanned  = 4
)

//channelParticipant#15ebac1d user_id:int date:int = ChannelParticipant;
//channelParticipantSelf#a3289a6d user_id:int inviter_id:int date:int = ChannelParticipant;
//channelParticipantCreator#e3e2e1f9 user_id:int = ChannelParticipant;
//channelParticipantAdmin#a82fa898 flags:# can_edit:flags.0?true user_id:int inviter_id:int promoted_by:int date:int admin_rights:ChannelAdminRights = ChannelParticipant;
//channelParticipantBanned#222c1886 flags:# left:flags.0?true user_id:int kicked_by:int date:int banned_rights:ChannelBannedRights = ChannelParticipant;

type channelLogicData struct {
	channel      *dataobject.ChannelsDO
	participants []dataobject.ChannelParticipantsDO
	dao          *channelsDAO
	cb           core.PhotoCallback
}

func makeChannelParticipantByDO(do *dataobject.ChannelParticipantsDO) (participant *mtproto.ChannelParticipant) {
	participant = &mtproto.ChannelParticipant{Data2: &mtproto.ChannelParticipant_Data{
		UserId:    do.UserId,
		InviterId: do.InviterUserId,
		Date:      do.JoinedAt,
	}}

	switch do.ParticipantType {
	case kChannelParticipant:
		participant.Constructor = mtproto.TLConstructor_CRC32_channelParticipant
	case kChannelParticipantCreator:
		participant.Constructor = mtproto.TLConstructor_CRC32_channelParticipantCreator
	case kChannelParticipantAdmin:
		participant.Constructor = mtproto.TLConstructor_CRC32_channelParticipantAdmin
	default:
		panic("channelParticipant type error.")
	}

	return
}

func MakeChannelParticipant2ByDO(selfId int32, do *dataobject.ChannelParticipantsDO) (participant *mtproto.ChannelParticipant) {
	participant = &mtproto.ChannelParticipant{Data2: &mtproto.ChannelParticipant_Data{
		UserId:    do.UserId,
		InviterId: do.InviterUserId,
		Date:      do.JoinedAt,
	}}

	switch do.ParticipantType {
	case kChannelParticipant:
		if do.UserId == selfId {
			participant.Constructor = mtproto.TLConstructor_CRC32_channelParticipantSelf
		} else {
			participant.Constructor = mtproto.TLConstructor_CRC32_channelParticipant
		}
	case kChannelParticipantCreator:
		participant.Constructor = mtproto.TLConstructor_CRC32_channelParticipantCreator
	case kChannelParticipantAdmin:
		participant.Constructor = mtproto.TLConstructor_CRC32_channelParticipantAdmin
	case kChannelParticipantBanned:
		participant.Constructor = mtproto.TLConstructor_CRC32_channelParticipantBanned
	default:
		panic("channelParticipant type error.")
	}

	return
}

func (m *ChannelModel) NewChannelLogicById(channelId int32) (channelData *channelLogicData, err error) {
	channelDO := m.dao.ChannelsDAO.Select(channelId)
	if channelDO == nil {
		err = mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_CHAT_ID_INVALID)
	} else {
		channelData = &channelLogicData{
			channel: channelDO,
			dao:     m.dao,
			cb:      m.photoCallback,
		}
	}
	return
}

func (m *ChannelModel) NewChannelLogicByCreateChannel(creatorId int32, title, about string) (channelData *channelLogicData) {
	// TODO(@benqi): 事务
	channelData = &channelLogicData{
		channel: &dataobject.ChannelsDO{
			CreatorUserId: creatorId,
			AccessHash:    rand.Int63(),
			// TODO(@benqi): use message_id is randomid
			// RandomId:         helper.NextSnowflakeId(),
			ParticipantCount: 1,
			Title:            title,
			About:            about,
			PhotoId:          0,
			Version:          1,
			Date:             int32(time.Now().Unix()),
		},
		// participants: make([]dataobject.ChannelParticipantsDO, 1+len(userIds)),
		dao: m.dao,
		cb:  m.photoCallback,
	}
	channelData.channel.Id = int32(m.dao.ChannelsDAO.Insert(channelData.channel))

	participant := &dataobject.ChannelParticipantsDO{
		ChannelId:       channelData.channel.Id,
		UserId:          creatorId,
		ParticipantType: kChannelParticipantCreator,
	}
	m.dao.ChannelParticipantsDAO.Insert(participant)

	return
}

func (m *channelLogicData) GetPhotoId() int64 {
	return m.channel.PhotoId
}

func (m *channelLogicData) GetChannelId() int32 {
	return m.channel.Id
}

func (m *channelLogicData) GetVersion() int32 {
	return m.channel.Version
}

func (m *channelLogicData) ExportedChatInvite() string {
	if m.channel.Link == "" {
		// TODO(@benqi): 检查唯一性
		m.channel.Link = "https://nebula.im/joinchat/" + base64.StdEncoding.EncodeToString(crypto.GenerateNonce(16))
		m.dao.ChannelsDAO.UpdateLink(m.channel.Link, int32(time.Now().Unix()), m.channel.Id)
	}
	return m.channel.Link
}

// TODO(@benqi): 性能优化
func (m *channelLogicData) checkUserIsAdministrator(userId int32) bool {
	m.checkOrLoadChannelParticipantList()
	for i := 0; i < len(m.participants); i++ {
		if m.participants[i].ParticipantType == kChannelParticipantCreator ||
			m.participants[i].ParticipantType == kChannelParticipantAdmin {
			return true
		}
	}
	return false
}

func (m *channelLogicData) checkOrLoadChannelParticipantList() {
	if len(m.participants) == 0 {
		m.participants = m.dao.ChannelParticipantsDAO.SelectByChannelId(m.channel.Id)
	}
}

func (m *channelLogicData) MakeMessageService(fromId int32, action *mtproto.MessageAction) *mtproto.Message {
	peer := &base.PeerUtil{
		PeerType: base.PEER_CHANNEL,
		PeerId:   m.channel.Id,
	}

	message := &mtproto.TLMessageService{Data2: &mtproto.Message_Data{
		Date:   m.channel.Date,
		FromId: fromId,
		ToId:   peer.ToPeer(),
		Post:   true,
		Action: action,
	}}
	return message.To_Message()
}

func (m *channelLogicData) MakeCreateChannelMessage(creatorId int32) *mtproto.Message {
	action := &mtproto.TLMessageActionChannelCreate{Data2: &mtproto.MessageAction_Data{
		Title: m.channel.Title,
	}}
	return m.MakeMessageService(creatorId, action.To_MessageAction())
}

func (m *channelLogicData) MakeAddUserMessage(inviterId, channelUserId int32) *mtproto.Message {
	action := &mtproto.TLMessageActionChatAddUser{Data2: &mtproto.MessageAction_Data{
		Title: m.channel.Title,
		Users: []int32{channelUserId},
	}}

	return m.MakeMessageService(inviterId, action.To_MessageAction())
}

func (m *channelLogicData) MakeDeleteUserMessage(operatorId, channelUserId int32) *mtproto.Message {
	action := &mtproto.TLMessageActionChatDeleteUser{Data2: &mtproto.MessageAction_Data{
		Title:  m.channel.Title,
		UserId: channelUserId,
	}}

	return m.MakeMessageService(operatorId, action.To_MessageAction())
}

func (m *channelLogicData) MakeChannelEditTitleMessage(operatorId int32, title string) *mtproto.Message {
	action := &mtproto.TLMessageActionChatEditTitle{Data2: &mtproto.MessageAction_Data{
		Title: title,
	}}

	return m.MakeMessageService(operatorId, action.To_MessageAction())
}

func (m *channelLogicData) GetChannelParticipantList() []*mtproto.ChannelParticipant {
	m.checkOrLoadChannelParticipantList()

	participantList := make([]*mtproto.ChannelParticipant, 0, len(m.participants))
	for i := 0; i < len(m.participants); i++ {
		if m.participants[i].State == 0 {
			participantList = append(participantList, makeChannelParticipantByDO(&m.participants[i]))
		}
	}
	return participantList
}

func (m *channelLogicData) GetChannelParticipantIdList() []int32 {
	m.checkOrLoadChannelParticipantList()

	idList := make([]int32, 0, len(m.participants))
	for i := 0; i < len(m.participants); i++ {
		if m.participants[i].State == 0 {
			idList = append(idList, m.participants[i].UserId)
		}
	}
	return idList
}

func (m *channelLogicData) GetChannelParticipants() *mtproto.TLChannelsChannelParticipants {
	m.checkOrLoadChannelParticipantList()

	return &mtproto.TLChannelsChannelParticipants{Data2: &mtproto.Channels_ChannelParticipants_Data{
		// ChatId: this.channel.Id,
		Participants: m.GetChannelParticipantList(),
		// Version: this.channel.Version,
	}}
}

func (m *channelLogicData) AddChannelUser(inviterId, userId int32) error {
	m.checkOrLoadChannelParticipantList()

	// TODO(@benqi): check userId exisits
	var founded = -1
	for i := 0; i < len(m.participants); i++ {
		if userId == m.participants[i].UserId {
			if m.participants[i].State == 1 {
				founded = i
			} else {
				return fmt.Errorf("userId exisits")
			}
		}
	}

	var now = int32(time.Now().Unix())

	if founded != -1 {
		m.participants[founded].State = 0
		m.dao.ChannelParticipantsDAO.Update(inviterId, now, now, m.participants[founded].Id)
	} else {
		channelParticipant := &dataobject.ChannelParticipantsDO{
			ChannelId:       m.channel.Id,
			UserId:          userId,
			ParticipantType: kChannelParticipant,
			InviterUserId:   inviterId,
			InvitedAt:       now,
			JoinedAt:        now,
		}
		channelParticipant.Id = int32(m.dao.ChannelParticipantsDAO.Insert(channelParticipant))
		m.participants = append(m.participants, *channelParticipant)
	}

	// update chat
	m.channel.ParticipantCount += 1
	m.channel.Version += 1
	m.channel.Date = now
	m.dao.ChannelsDAO.UpdateParticipantCount(m.channel.ParticipantCount, now, m.channel.Id)

	return nil
}

func (m *channelLogicData) findChatParticipant(selfUserId int32) (int, *dataobject.ChannelParticipantsDO) {
	for i := 0; i < len(m.participants); i++ {
		if m.participants[i].UserId == selfUserId {
			return i, &m.participants[i]
		}
	}
	return -1, nil
}

func (m *channelLogicData) ToChannel(selfUserId int32) *mtproto.Chat {
	// TODO(@benqi): kicked:flags.1?true left:flags.2?true admins_enabled:flags.3?true admin:flags.4?true deactivated:flags.5?true

	var forbidden = false
	//for i := 0; i < len(this.participants); i++ {
	//	if this.participants[i].UserId == selfUserId && this.participants[i].State == 1 {
	//		forbidden = true
	//		break
	//	}
	//}

	if forbidden {
		channel := &mtproto.TLChannelForbidden{Data2: &mtproto.Chat_Data{
			Id:    m.channel.Id,
			Title: m.channel.Title,
		}}
		return channel.To_Chat()
	} else {
		// channel#450b7115 flags:#
		// 	creator:flags.0?true
		// 	left:flags.2?true
		// 	editor:flags.3?true
		// 	broadcast:flags.5?true
		// 	verified:flags.7?true
		// 	megagroup:flags.8?true
		// 	restricted:flags.9?true
		// 	democracy:flags.10?true
		// 	signatures:flags.11?true
		// 	min:flags.12?true
		//  id:int
		// 	access_hash:flags.13?long
		// 	title:string
		// 	username:flags.6?string
		// 	photo:ChatPhoto
		// 	date:int
		// 	version:int
		// 	restriction_reason:flags.9?string
		// 	admin_rights:flags.14?ChannelAdminRights
		// 	banned_rights:flags.15?ChannelBannedRights
		// 	participants_count:flags.17?int = Chat;
		channel := &mtproto.TLChannel{Data2: &mtproto.Chat_Data{
			Creator:    m.channel.CreatorUserId == selfUserId,
			Id:         m.channel.Id,
			AccessHash: rand.Int63(),
			Title:      m.channel.Title,
			// AdminsEnabled:     this.channel.AdminsEnabled == 1,
			// ParticipantsCount: this.channel.ParticipantCount,
			Date:    m.channel.Date,
			Version: m.channel.Version,
		}}

		if m.channel.PhotoId == 0 {
			channel.SetPhoto(mtproto.NewTLChatPhotoEmpty().To_ChatPhoto())
		} else {
			channel.SetPhoto(m.cb.GetChatPhoto(m.channel.PhotoId))
		}
		return channel.To_Chat()
	}
}

func (m *channelLogicData) CheckDeleteChannelUser(operatorId, deleteUserId int32) error {
	// operatorId is creatorUserId，allow delete all user_id
	// other delete me
	if operatorId != m.channel.CreatorUserId && operatorId != deleteUserId {
		return mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_NO_EDIT_CHAT_PERMISSION)
	}

	m.checkOrLoadChannelParticipantList()
	var found = -1
	for i := 0; i < len(m.participants); i++ {
		if deleteUserId == m.participants[i].UserId {
			if m.participants[i].State == 0 {
				found = i
			}
			break
		}
	}

	if found == -1 {
		return mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_PARTICIPANT_NOT_EXISTS)
	}

	return nil
}

func (m *channelLogicData) DeleteChannelUser(operatorId, deleteUserId int32) error {
	// operatorId is creatorUserId，allow delete all user_id
	// other delete me
	if operatorId != m.channel.CreatorUserId && operatorId != deleteUserId {
		return mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_NO_EDIT_CHAT_PERMISSION)
	}

	m.checkOrLoadChannelParticipantList()
	var found = -1
	for i := 0; i < len(m.participants); i++ {
		if deleteUserId == m.participants[i].UserId {
			if m.participants[i].State == 0 {
				found = i
			}
			break
		}
	}

	if found == -1 {
		return mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_PARTICIPANT_NOT_EXISTS)
	}

	m.participants[found].State = 1
	m.dao.ChannelParticipantsDAO.DeleteChannelUser(m.participants[found].Id)

	// delete found.
	// this.participants = append(this.participants[:found], this.participants[found+1:]...)

	var now = int32(time.Now().Unix())
	m.channel.ParticipantCount = int32(len(m.participants) - 1)
	m.channel.Version += 1
	m.channel.Date = now
	m.dao.ChannelsDAO.UpdateParticipantCount(m.channel.ParticipantCount, now, m.channel.Id)

	return nil
}

func (m *channelLogicData) EditChannelTitle(editUserId int32, title string) error {
	m.checkOrLoadChannelParticipantList()

	_, participant := m.findChatParticipant(editUserId)

	if participant == nil || participant.State == 1 {
		return mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_PARTICIPANT_NOT_EXISTS)
	}

	// check editUserId is creator or admin
	if m.channel.AdminsEnabled != 0 && participant.ParticipantType == kChannelParticipant {
		return mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_NO_EDIT_CHAT_PERMISSION)
	}

	if m.channel.Title == title {
		return mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_CHAT_NOT_MODIFIED)
	}

	m.channel.Title = title
	m.channel.Date = int32(time.Now().Unix())
	m.channel.Version += 1

	m.dao.ChannelsDAO.UpdateTitle(title, m.channel.Date, m.channel.Id)
	return nil
}

func (m *channelLogicData) EditChannelPhoto(editUserId int32, photoId int64) error {
	m.checkOrLoadChannelParticipantList()

	_, participant := m.findChatParticipant(editUserId)

	if participant == nil || participant.State == 1 {
		return mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_PARTICIPANT_NOT_EXISTS)
	}

	// check editUserId is creator or admin
	if m.channel.AdminsEnabled != 0 && participant.ParticipantType == kChannelParticipant {
		return mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_NO_EDIT_CHAT_PERMISSION)
	}

	if m.channel.PhotoId == photoId {
		return mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_CHAT_NOT_MODIFIED)
	}

	m.channel.PhotoId = photoId
	m.channel.Date = int32(time.Now().Unix())
	m.channel.Version += 1

	m.dao.ChannelsDAO.UpdatePhotoId(photoId, m.channel.Date, m.channel.Id)
	return nil
}

func (m *channelLogicData) EditChannelAdmin(operatorId, editChannelAdminId int32, isAdmin bool) error {
	// operatorId is creator
	if operatorId != m.channel.CreatorUserId {
		return mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_NO_EDIT_CHAT_PERMISSION)
	}

	// editChatAdminId not creator
	if editChannelAdminId == operatorId {
		return mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_BAD_REQUEST)
	}

	m.checkOrLoadChannelParticipantList()

	// check exists
	_, participant := m.findChatParticipant(editChannelAdminId)
	if participant == nil || participant.State == 1 {
		return mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_PARTICIPANT_NOT_EXISTS)
	}

	if isAdmin && participant.ParticipantType == kChannelParticipantAdmin || !isAdmin && participant.ParticipantType == kChannelParticipant {
		return mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_CHAT_NOT_MODIFIED)
	}

	if isAdmin {
		participant.ParticipantType = kChannelParticipantAdmin
	} else {
		participant.ParticipantType = kChannelParticipant
	}
	m.dao.ChannelParticipantsDAO.UpdateParticipantType(participant.ParticipantType, participant.Id)

	// update version
	m.channel.Date = int32(time.Now().Unix())
	m.channel.Version += 1
	m.dao.ChannelsDAO.UpdateVersion(m.channel.Date, m.channel.Id)

	return nil
}

func (m *channelLogicData) ToggleChannelAdmins(userId int32, adminsEnabled bool) error {
	// check is creator
	if userId != m.channel.CreatorUserId {
		return mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_NO_EDIT_CHAT_PERMISSION)
	}

	var (
		channelAdminsEnabled = m.channel.AdminsEnabled == 1
	)

	// Check modified
	if channelAdminsEnabled == adminsEnabled {
		return mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_CHAT_NOT_MODIFIED)
	}

	m.channel.AdminsEnabled = base2.BoolToInt8(adminsEnabled)
	m.channel.Date = int32(time.Now().Unix())
	m.channel.Version += 1

	m.dao.ChannelsDAO.UpdateAdminsEnabled(m.channel.AdminsEnabled, m.channel.Date, m.channel.Id)

	return nil
}
