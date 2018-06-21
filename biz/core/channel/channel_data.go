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
	"github.com/nebulaim/telegramd/mtproto"
	"time"
	"github.com/nebulaim/telegramd/biz/base"
	"github.com/nebulaim/telegramd/biz/dal/dataobject"
	"math/rand"
	"github.com/nebulaim/telegramd/biz/dal/dao"
	"fmt"
	photo2 "github.com/nebulaim/telegramd/biz/core/photo"
	base2 "github.com/nebulaim/telegramd/baselib/base"
	"github.com/nebulaim/telegramd/biz/nbfs_client"
	"github.com/nebulaim/telegramd/baselib/crypto"
	"encoding/base64"
)

const (
	kChannelParticipant = 0
	kChannelParticipantSelf = 1
	kChannelParticipantCreator = 2
	kChannelParticipantAdmin = 3
	kChannelParticipantBanned = 4
)

//channelParticipant#15ebac1d user_id:int date:int = ChannelParticipant;
//channelParticipantSelf#a3289a6d user_id:int inviter_id:int date:int = ChannelParticipant;
//channelParticipantCreator#e3e2e1f9 user_id:int = ChannelParticipant;
//channelParticipantAdmin#a82fa898 flags:# can_edit:flags.0?true user_id:int inviter_id:int promoted_by:int date:int admin_rights:ChannelAdminRights = ChannelParticipant;
//channelParticipantBanned#222c1886 flags:# left:flags.0?true user_id:int kicked_by:int date:int banned_rights:ChannelBannedRights = ChannelParticipant;

type channelLogicData struct {
	channel      *dataobject.ChannelsDO
	participants []dataobject.ChannelParticipantsDO
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

func NewChannelLogicById(channelId int32) (channelData *channelLogicData, err error) {
	channelDO := dao.GetChannelsDAO(dao.DB_SLAVE).Select(channelId)
	if channelDO == nil {
		err = mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_CHAT_ID_INVALID)
	} else {
		channelData = &channelLogicData{
			channel: channelDO,
		}
	}
	return
}

func NewChannelLogicByCreateChannel(creatorId int32, userIds []int32, title, about string) (channelData *channelLogicData) {
	// TODO(@benqi): 事务
	channelData = &channelLogicData{
		channel: &dataobject.ChannelsDO{
			CreatorUserId: creatorId,
			AccessHash:    rand.Int63(),
			// TODO(@benqi): use message_id is randomid
			// RandomId:         helper.NextSnowflakeId(),
			ParticipantCount: int32(1 + len(userIds)),
			Title:            title,
			About:            about,
			PhotoId:          0,
			Version:          1,
			Date:             int32(time.Now().Unix()),
		},
		participants: make([]dataobject.ChannelParticipantsDO, 1+len(userIds)),
	}
	channelData.channel.Id = int32(dao.GetChannelsDAO(dao.DB_MASTER).Insert(channelData.channel))

	channelData.participants = make([]dataobject.ChannelParticipantsDO, 1 + len(userIds))
	channelData.participants[0].ChannelId = channelData.channel.Id
	channelData.participants[0].UserId = creatorId
	channelData.participants[0].ParticipantType = kChannelParticipantCreator
	dao.GetChannelParticipantsDAO(dao.DB_MASTER).Insert(&channelData.participants[0])

	for i := 0; i < len(userIds); i++ {
		channelData.participants[i+1].ChannelId = channelData.channel.Id
		channelData.participants[i+1].UserId = userIds[i]
		channelData.participants[i+1].ParticipantType = kChannelParticipant
		channelData.participants[i+1].InviterUserId = creatorId
		channelData.participants[i+1].InvitedAt = channelData.channel.Date
		channelData.participants[i+1].JoinedAt = channelData.channel.Date
		dao.GetChannelParticipantsDAO(dao.DB_MASTER).Insert(&channelData.participants[i+1])
	}
	return
}

func (this *channelLogicData) GetPhotoId() int64 {
	return this.channel.PhotoId
}

func (this *channelLogicData) GetChannelId() int32 {
	return this.channel.Id
}

func (this *channelLogicData) GetVersion() int32 {
	return this.channel.Version
}

func (this *channelLogicData) ExportedChatInvite() string {
	if this.channel.Link == "" {
		// TODO(@benqi): 检查唯一性
		this.channel.Link = "https://nebula.im/joinchat/" + base64.StdEncoding.EncodeToString(crypto.GenerateNonce(16))
		dao.GetChannelsDAO(dao.DB_MASTER).UpdateLink(this.channel.Link, int32(time.Now().Unix()), this.channel.Id)
	}
	return this.channel.Link
}

// TODO(@benqi): 性能优化
func (this *channelLogicData) checkUserIsAdministrator(userId int32) bool {
	this.checkOrLoadChannelParticipantList()
	for i := 0; i < len(this.participants); i++  {
		if this.participants[i].ParticipantType == kChannelParticipantCreator ||
			this.participants[i].ParticipantType == kChannelParticipantAdmin {
			return true
		}
	}
	return false
}

func (this *channelLogicData) checkOrLoadChannelParticipantList() {
	if len(this.participants) == 0 {
		this.participants = dao.GetChannelParticipantsDAO(dao.DB_SLAVE).SelectByChannelId(this.channel.Id)
	}
}

func (this *channelLogicData) MakeMessageService(fromId int32, action *mtproto.MessageAction) *mtproto.Message {
	peer := &base.PeerUtil{
		PeerType: base.PEER_CHANNEL,
		PeerId:   this.channel.Id,
	}

	message := &mtproto.TLMessageService{Data2: &mtproto.Message_Data{
		Date:   this.channel.Date,
		FromId: fromId,
		ToId:   peer.ToPeer(),
		Post:   true,
		Action: action,
	}}
	return message.To_Message()
}

func (this *channelLogicData) MakeCreateChannelMessage(creatorId int32) *mtproto.Message {
	action := &mtproto.TLMessageActionChannelCreate{Data2: &mtproto.MessageAction_Data{
		Title: this.channel.Title,
	}}
	return this.MakeMessageService(creatorId, action.To_MessageAction())
}

func (this *channelLogicData) MakeAddUserMessage(inviterId, channelUserId int32) *mtproto.Message {
	action := &mtproto.TLMessageActionChatAddUser{Data2: &mtproto.MessageAction_Data{
		Title: this.channel.Title,
		Users: []int32{channelUserId},
	}}

	return this.MakeMessageService(inviterId, action.To_MessageAction())
}

func (this *channelLogicData) MakeDeleteUserMessage(operatorId, channelUserId int32) *mtproto.Message {
	action := &mtproto.TLMessageActionChatDeleteUser{Data2: &mtproto.MessageAction_Data{
		Title:  this.channel.Title,
		UserId: channelUserId,
	}}

	return this.MakeMessageService(operatorId, action.To_MessageAction())
}

func (this *channelLogicData) MakeChannelEditTitleMessage(operatorId int32, title string) *mtproto.Message {
	action := &mtproto.TLMessageActionChatEditTitle{Data2: &mtproto.MessageAction_Data{
		Title:  title,
	}}

	return this.MakeMessageService(operatorId, action.To_MessageAction())
}

func (this *channelLogicData) GetChannelParticipantList() []*mtproto.ChannelParticipant {
	this.checkOrLoadChannelParticipantList()

	participantList := make([]*mtproto.ChannelParticipant, 0, len(this.participants))
	for i := 0; i < len(this.participants); i++ {
		if this.participants[i].State == 0 {
			participantList = append(participantList, makeChannelParticipantByDO(&this.participants[i]))
		}
	}
	return participantList
}


func (this *channelLogicData) GetChannelParticipantIdList() []int32 {
	this.checkOrLoadChannelParticipantList()

	idList := make([]int32, 0, len(this.participants))
	for i := 0; i < len(this.participants); i++  {
		if this.participants[i].State == 0 {
			idList = append(idList, this.participants[i].UserId)
		}
	}
	return idList
}

func (this *channelLogicData) GetChannelParticipants() *mtproto.TLChannelsChannelParticipants {
	this.checkOrLoadChannelParticipantList()

	return &mtproto.TLChannelsChannelParticipants{Data2: &mtproto.Channels_ChannelParticipants_Data{
		// ChatId: this.channel.Id,
		Participants: this.GetChannelParticipantList(),
		// Version: this.channel.Version,
	}}
}

func (this *channelLogicData) AddChannelUser(inviterId, userId int32) error {
	this.checkOrLoadChannelParticipantList()

	// TODO(@benqi): check userId exisits
	var founded = -1
	for i := 0; i < len(this.participants); i++ {
		if userId == this.participants[i].UserId {
			if this.participants[i].State == 1 {
				founded = i
			} else {
				return fmt.Errorf("userId exisits")
			}
		}
	}

	var now = int32(time.Now().Unix())

	if founded != -1 {
		this.participants[founded].State = 0
		dao.GetChannelParticipantsDAO(dao.DB_MASTER).Update(inviterId, now, now, this.participants[founded].Id)
	} else {
		channelParticipant := &dataobject.ChannelParticipantsDO{
			ChannelId:       this.channel.Id,
			UserId:          userId,
			ParticipantType: kChannelParticipant,
			InviterUserId:   inviterId,
			InvitedAt:       now,
			JoinedAt:        now,
		}
		channelParticipant.Id = int32(dao.GetChannelParticipantsDAO(dao.DB_MASTER).Insert(channelParticipant))
		this.participants = append(this.participants, *channelParticipant)
	}

	// update chat
	this.channel.ParticipantCount += 1
	this.channel.Version += 1
	this.channel.Date = now
	dao.GetChannelsDAO(dao.DB_MASTER).UpdateParticipantCount(this.channel.ParticipantCount, now, this.channel.Id)

	return nil
}

func (this *channelLogicData) findChatParticipant(selfUserId int32) (int, *dataobject.ChannelParticipantsDO) {
	for i := 0; i < len(this.participants); i++ {
		if this.participants[i].UserId == selfUserId {
			return i, &this.participants[i]
		}
	}
	return -1, nil
}

func (this *channelLogicData) ToChannel(selfUserId int32) *mtproto.Chat {
	// TODO(@benqi): kicked:flags.1?true left:flags.2?true admins_enabled:flags.3?true admin:flags.4?true deactivated:flags.5?true

	var forbidden = false
	for i := 0; i < len(this.participants); i++ {
		if this.participants[i].UserId == selfUserId && this.participants[i].State == 1 {
			forbidden = true
			break
		}
	}

	if forbidden {
		channel := &mtproto.TLChannelForbidden{Data2: &mtproto.Chat_Data{
			Id:    this.channel.Id,
			Title: this.channel.Title,
		}}
		return channel.To_Chat()
	} else {
		channel := &mtproto.TLChannel{Data2: &mtproto.Chat_Data{
			Creator:           this.channel.CreatorUserId == selfUserId,
			Id:                this.channel.Id,
			Title:             this.channel.Title,
			AdminsEnabled:     this.channel.AdminsEnabled == 1,
			// Photo:             mtproto.NewTLChatPhotoEmpty().To_ChatPhoto(),
			ParticipantsCount: this.channel.ParticipantCount,
			Date:              this.channel.Date,
			Version:           this.channel.Version,
		}}

		if this.channel.PhotoId == 0 {
			channel.SetPhoto(mtproto.NewTLChatPhotoEmpty().To_ChatPhoto())
		} else {
			sizeList, _ := nbfs_client.GetPhotoSizeList(this.channel.PhotoId)
			channel.SetPhoto(photo2.MakeChatPhoto(sizeList))
		}
		return channel.To_Chat()
	}
}

func (this *channelLogicData) CheckDeleteChannelUser(operatorId, deleteUserId int32) error {
	// operatorId is creatorUserId，allow delete all user_id
	// other delete me
	if operatorId != this.channel.CreatorUserId && operatorId != deleteUserId {
		return mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_NO_EDIT_CHAT_PERMISSION)
	}

	this.checkOrLoadChannelParticipantList()
	var found = -1
	for i := 0; i < len(this.participants); i++ {
		if deleteUserId == this.participants[i].UserId {
			if this.participants[i].State == 0 {
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

func (this *channelLogicData) DeleteChannelUser(operatorId, deleteUserId int32) error {
	// operatorId is creatorUserId，allow delete all user_id
	// other delete me
	if operatorId != this.channel.CreatorUserId && operatorId != deleteUserId {
		return mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_NO_EDIT_CHAT_PERMISSION)
	}

	this.checkOrLoadChannelParticipantList()
	var found = -1
	for i := 0; i < len(this.participants); i++ {
		if deleteUserId == this.participants[i].UserId {
			if this.participants[i].State == 0 {
				found = i
			}
			break
		}
	}

	if found == -1 {
		return mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_PARTICIPANT_NOT_EXISTS)
	}

	this.participants[found].State = 1
	dao.GetChannelParticipantsDAO(dao.DB_MASTER).DeleteChannelUser(this.participants[found].Id)

	// delete found.
	// this.participants = append(this.participants[:found], this.participants[found+1:]...)

	var now = int32(time.Now().Unix())
	this.channel.ParticipantCount = int32(len(this.participants)-1)
	this.channel.Version += 1
	this.channel.Date = now
	dao.GetChannelsDAO(dao.DB_MASTER).UpdateParticipantCount(this.channel.ParticipantCount, now, this.channel.Id)

	return nil
}

func (this *channelLogicData) EditChannelTitle(editUserId int32, title string) error {
	this.checkOrLoadChannelParticipantList()

	_, participant := this.findChatParticipant(editUserId)

	if participant == nil || participant.State == 1 {
		return mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_PARTICIPANT_NOT_EXISTS)
	}

	// check editUserId is creator or admin
	if this.channel.AdminsEnabled != 0 && participant.ParticipantType == kChannelParticipant {
		return mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_NO_EDIT_CHAT_PERMISSION)
	}

	if this.channel.Title == title {
		return mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_CHAT_NOT_MODIFIED)
	}

	this.channel.Title = title
	this.channel.Date = int32(time.Now().Unix())
	this.channel.Version += 1

	dao.GetChannelsDAO(dao.DB_MASTER).UpdateTitle(title, this.channel.Date, this.channel.Id)
	return nil
}

func (this *channelLogicData) EditChannelPhoto(editUserId int32, photoId int64) error {
	this.checkOrLoadChannelParticipantList()

	_, participant := this.findChatParticipant(editUserId)

	if participant == nil || participant.State == 1 {
		return mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_PARTICIPANT_NOT_EXISTS)
	}

	// check editUserId is creator or admin
	if this.channel.AdminsEnabled != 0 && participant.ParticipantType == kChannelParticipant {
		return mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_NO_EDIT_CHAT_PERMISSION)
	}

	if this.channel.PhotoId == photoId {
		return mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_CHAT_NOT_MODIFIED)
	}

	this.channel.PhotoId = photoId
	this.channel.Date = int32(time.Now().Unix())
	this.channel.Version += 1

	dao.GetChannelsDAO(dao.DB_MASTER).UpdatePhotoId(photoId, this.channel.Date, this.channel.Id)
	return nil
}

func (this *channelLogicData) EditChannelAdmin(operatorId, editChannelAdminId int32, isAdmin bool) error {
	// operatorId is creator
	if operatorId != this.channel.CreatorUserId {
		return mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_NO_EDIT_CHAT_PERMISSION)
	}

	// editChatAdminId not creator
	if editChannelAdminId == operatorId {
		return mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_BAD_REQUEST)
	}

	this.checkOrLoadChannelParticipantList()

	// check exists
	_, participant := this.findChatParticipant(editChannelAdminId)
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
	dao.GetChannelParticipantsDAO(dao.DB_MASTER).UpdateParticipantType(participant.ParticipantType, participant.Id)

	// update version
	this.channel.Date = int32(time.Now().Unix())
	this.channel.Version += 1
	dao.GetChannelsDAO(dao.DB_MASTER).UpdateVersion(this.channel.Date, this.channel.Id)

	return nil
}

func (this *channelLogicData) ToggleChannelAdmins(userId int32, adminsEnabled bool) error {
	// check is creator
	if userId != this.channel.CreatorUserId {
		return mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_NO_EDIT_CHAT_PERMISSION)
	}

	var (
		channelAdminsEnabled = this.channel.AdminsEnabled == 1
	)

	// Check modified
	if channelAdminsEnabled == adminsEnabled {
		return mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_CHAT_NOT_MODIFIED)
	}

	this.channel.AdminsEnabled = base2.BoolToInt8(adminsEnabled)
	this.channel.Date = int32(time.Now().Unix())
	this.channel.Version += 1

	dao.GetChannelsDAO(dao.DB_MASTER).UpdateAdminsEnabled(this.channel.AdminsEnabled, this.channel.Date, this.channel.Id)

	return nil
}
