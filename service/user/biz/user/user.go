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

package user

import (
	"github.com/nebulaim/telegramd/proto/mtproto"
	"github.com/nebulaim/telegramd/biz/dal/dataobject"
	"github.com/nebulaim/telegramd/baselib/base"
	"time"
	"github.com/nebulaim/telegramd/biz/dal/dao/mysql_dao"
	"encoding/hex"
	"github.com/nebulaim/telegramd/baselib/crypto"
	"math/rand"
	"github.com/nebulaim/telegramd/baselib/mysql_client"
	"fmt"
	"github.com/golang/glog"
	photo2 "github.com/nebulaim/telegramd/biz/core/photo"
	"github.com/nebulaim/telegramd/server/nbfs/nbfs_client"
	contact2 "github.com/nebulaim/telegramd/biz/core/contact"
)

type userData struct {
	*mtproto.TLUser
}

func (m *userData) ToUser() *mtproto.User {
	return m.TLUser.To_User()
}

type userModel struct {
	*mysql_dao.UsersDAO
	*mysql_dao.UserPresencesDAO
	*mysql_dao.UserPasswordsDAO
	*mysql_dao.CommonDAO
}

func InitUserModel(dbName string) (*userModel, error) {
	dbClient := mysql_client.GetMysqlClient(dbName)
	if dbClient == nil {
		err := fmt.Errorf("invalid dbName: %s", dbName)
		glog.Error(err)
		return nil, err
	}

	m := &userModel{
		UsersDAO: mysql_dao.NewUsersDAO(dbClient),
		UserPresencesDAO: mysql_dao.NewUserPresencesDAO(dbClient),
		UserPasswordsDAO: mysql_dao.NewUserPasswordsDAO(dbClient),
		CommonDAO: mysql_dao.NewCommonDAO(dbClient),
	}
	return m, nil
}

func makeUserStatusOnline() *mtproto.UserStatus {
	now := time.Now().Unix()
	status := &mtproto.UserStatus{
		Constructor: mtproto.TLConstructor_CRC32_userStatusOnline,
		Data2: &mtproto.UserStatus_Data{
			// WasOnline: int32(now),
			Expires:   int32(now + 60),
		},
	}
	return status
}

func (m *userModel) CheckUserAccessHash(id int32, hash int64) bool {
	return true
}

func (m *userModel) CheckPhoneNumberExist(phoneNumber string) bool {
	return nil != m.UsersDAO.SelectByPhoneNumber(phoneNumber)
}

func (m *userModel) GetUserByID(selfId, id int32) (user *mtproto.User) {
	do := m.SelectById(id)
	if do != nil {
		user = m.makeUserDataByDO(selfId, do).To_User()
	}
	return
}

func (m *userModel) GetUserListByIDList(selfId int32, idList []int32) (users []*mtproto.User) {
	users = make([]*mtproto.User, 0, len(idList))
	if len(idList) > 0 {
		userDOList := m.UsersDAO.SelectUsersByIdList(idList)

		// TODO(@benqi):  需要优化，makeUserDataByDO需要查询用户状态以及获取Mutual和Contact状态信息而导致多次查询
		users = make([]*mtproto.User, 0, len(userDOList))
		for i := 0; i < len(userDOList); i++ {
			user := m.makeUserDataByDO(selfId, &userDOList[i])
			users = append(users, user.To_User())
		}
	}
	return
}

func (m *userModel) GetUserByPhoneNumber(selfId int32, phone string) (user *mtproto.User) {
	do := m.UsersDAO.SelectByPhoneNumber(phone)
	if do != nil {
		do.Phone = phone
		user = m.makeUserDataByDO(selfId, do).To_User()
	}
	return
}

func (m *userModel) GetSelfUserByPhoneNumber(phoneNumber string) (user *mtproto.User) {
	do := m.UsersDAO.SelectByPhoneNumber(phoneNumber)
	if do != nil {
		do.Phone = phoneNumber
		user = m.makeUserDataByDO(do.Id, do).To_User()
	}
	return
}

func (m *userModel) UpdateUserStatus(userId int32, lastSeenAt int64) {
	// presencesDAO := dao.GetUserPresencesDAO(dao.DB_MASTER)
	// now := time.Now().Unix()
	rows := m.UserPresencesDAO.UpdateLastSeen(lastSeenAt, 0, userId)
	if rows == 0 {
		do := &dataobject.UserPresencesDO{
			UserId: userId,
			LastSeenAt: lastSeenAt,
			LastSeenAuthKeyId: 0,
			LastSeenIp: "",
			CreatedAt: base.NowFormatYMDHMS(),
		}
		m.UserPresencesDAO.Insert(do)
	}
}

func (m *userModel) GetUserStatus(userId int32) *mtproto.UserStatus {
	now := time.Now().Unix()
	do := m.UserPresencesDAO.SelectByUserID(userId)
	if do == nil {
		return mtproto.NewTLUserStatusEmpty().To_UserStatus()
	}

	if now <= do.LastSeenAt + 5*60 {
		status := &mtproto.TLUserStatusOnline{Data2: &mtproto.UserStatus_Data{
			Expires: int32(do.LastSeenAt + 5*30),
		}}
		return status.To_UserStatus()
	} else {
		status := &mtproto.TLUserStatusOffline{Data2: &mtproto.UserStatus_Data{
			WasOnline: int32(do.LastSeenAt),
		}}
		return status.To_UserStatus()
	}
}

func (m *userModel) CreateNewUser(phoneNumber, countryCode, firstName, lastName string) *mtproto.TLUser {
	// usersDAO := dao.GetUsersDAO(dao.DB_SLAVE)
	do := &dataobject.UsersDO{
		AccessHash:  rand.Int63(),
		Phone:       phoneNumber,
		FirstName:   firstName,
		LastName:    lastName,
		CountryCode: countryCode,
	}
	do.Id = int32(m.UsersDAO.Insert(do))
	user := &mtproto.TLUser{ Data2: &mtproto.User_Data{
		Id:            do.Id,
		Self:          true,
		Contact:       true,
		MutualContact: true,
		AccessHash:    do.AccessHash,
		FirstName:     do.FirstName,
		LastName:      do.LastName,
		Username:      do.Username,
		Phone:         phoneNumber,
		// TODO(@benqi): Load from db
		Photo:         mtproto.NewTLUserProfilePhotoEmpty().To_UserProfilePhoto(),
		Status:        makeUserStatusOnline(),
	}}
	return user
}

func (m *userModel) CreateNewUserPassword(userId int32) {
	// gen server_nonce
	do := &dataobject.UserPasswordsDO{
		UserId:     userId,
		ServerSalt: hex.EncodeToString(crypto.GenerateNonce(8)),
	}
	m.UserPasswordsDAO.Insert(do)
}

func (m *userModel) CheckAccessHashByUserId(userId int32, accessHash int64) bool {
	params := map[string]interface{}{
		"id":          userId,
		"access_hash": accessHash,
	}
	return m.CommonDAO.CheckExists("users", params)
}

func (m *userModel) GetCountryCodeByUser(userId int32) string {
	do := m.UsersDAO.SelectCountryCode(userId)
	if do == nil {
		return ""
	} else {
		return do.CountryCode
	}
}

func (m *userModel) GetDefaultUserPhotoID(userId int32) int64 {
	do := m.UsersDAO.SelectProfilePhotos(userId)
	if do != nil {
		photoIds := MakeProfilePhotoData(do.Photos)
		return photoIds.Default
	}
	return 0
}

func (m *userModel) GetUserPhotoIDList(userId int32) []int64 {
	do := m.UsersDAO.SelectProfilePhotos(userId)
	if do != nil {
		photoIds := MakeProfilePhotoData(do.Photos)
		return photoIds.IdList
	}
	return []int64{}
}

func (m *userModel) SetUserPhotoID(userId int32, photoId int64) {
	do := m.UsersDAO.SelectProfilePhotos(userId)
	if do != nil {
		photoIds := MakeProfilePhotoData(do.Photos)
		photoIds.AddPhotoId(photoId)
		m.UsersDAO.UpdateProfilePhotos(photoIds.ToJson(), userId)
	}
}

func (m *userModel) DeleteUserPhotoID(userId int32, photoId int64) {
	do := m.UsersDAO.SelectProfilePhotos(userId)
	if do != nil {
		photoIds := MakeProfilePhotoData(do.Photos)
		photoIds.RemovePhotoId(photoId)
		m.UsersDAO.UpdateProfilePhotos(photoIds.ToJson(), userId)
	}
}


// user#2e13f4c3 flags:#
//  self:flags.10?true
// 	contact:flags.11?true
// 	mutual_contact:flags.12?true
// 	deleted:flags.13?true
// 	bot:flags.14?true
// 	bot_chat_history:flags.15?true
// 	bot_nochats:flags.16?true
// 	verified:flags.17?true
// 	restricted:flags.18?true
// 	min:flags.20?true
// 	bot_inline_geo:flags.21?true
// 	id:int
// 	access_hash:flags.0?long
// 	first_name:flags.1?string
// 	last_name:flags.2?string
// 	username:flags.3?string
// 	phone:flags.4?string
// 	photo:flags.5?UserProfilePhoto
// 	status:flags.6?UserStatus
// 	bot_info_version:flags.14?int
// 	restriction_reason:flags.18?string
// 	bot_inline_placeholder:flags.19?string
// 	lang_code:flags.22?string = User;
func (m *userModel)  makeUserDataByDO(selfId int32, do *dataobject.UsersDO) *userData {
	if do == nil {
		return nil
	} else {
		var (
			status *mtproto.UserStatus
			photo *mtproto.UserProfilePhoto
			phone string
			contact, mutualContact bool
			isSelf = selfId == do.Id
		)

		if isSelf {
			status = makeUserStatusOnline()
			contact = true
			mutualContact = true
			phone = do.Phone
		} else {
			status = m.GetUserStatus(do.Id)
			contact, mutualContact = contact2.CheckContactAndMutualByUserId(selfId, do.Id)
			if contact {
				phone = do.Phone
			}
		}

		photoId := m.GetDefaultUserPhotoID(do.Id)
		if photoId == 0 {
			photo =  mtproto.NewTLUserProfilePhotoEmpty().To_UserProfilePhoto()
		} else {
			sizeList, _ := nbfs_client.GetPhotoSizeList(photoId)
			photo = photo2.MakeUserProfilePhoto(photoId, sizeList)
		}

		data := &userData{ TLUser: &mtproto.TLUser{ Data2: &mtproto.User_Data{
			Id:            do.Id,
			Self:          isSelf,
			Contact:       contact,
			MutualContact: mutualContact,
			AccessHash:    do.AccessHash,
			FirstName:     do.FirstName,
			LastName:      do.LastName,
			Username:      do.Username,
			Phone:         phone,
			Photo:         photo,
			Status:        status,
		}}}

		return data
	}
}

