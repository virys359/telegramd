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
	"github.com/nebulaim/telegramd/mtproto"
	"github.com/nebulaim/telegramd/biz/dal/dao"
	"time"
	"github.com/nebulaim/telegramd/biz/dal/dataobject"
	"github.com/nebulaim/telegramd/biz/base"
	"github.com/nebulaim/telegramd/baselib/crypto"
	"encoding/hex"
	contact2 "github.com/nebulaim/telegramd/biz/core/contact"
	photo2 "github.com/nebulaim/telegramd/biz/core/photo"
)

func CheckUserAccessHash(id int32, hash int64) bool {
	return true
}

func CheckPhoneNumberExist(phoneNumber string) bool {
	return nil != dao.GetUsersDAO(dao.DB_SLAVE).SelectByPhoneNumber(phoneNumber)
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

func makeUserDataByDO(selfId int32, do *dataobject.UsersDO) *userData {
	if do == nil {
		return nil
	} else {
		var (
			status *mtproto.UserStatus
			photo *mtproto.UserProfilePhoto
		)

		if selfId == do.Id {
			status = makeUserStatusOnline()
		} else {
			status = GetUserStatus(do.Id)
		}

		photoId := GetDefaultUserPhotoID(do.Id)
		if photoId == 0 {
			photo =  mtproto.NewTLUserProfilePhotoEmpty().To_UserProfilePhoto()
		} else {
			sizeList := photo2.GetPhotoSizeList(photoId)
			photo = photo2.MakeUserProfilePhoto(photoId, sizeList)
		}
		// GetPhotoSizeList(photoId)
		contact, mutalContact := contact2.CheckContactAndMutualByUserId(selfId, do.Id)
		data := &userData{ TLUser: &mtproto.TLUser{ Data2: &mtproto.User_Data{
			Id:            do.Id,
			Self:          selfId == do.Id,
			Contact:       contact,
			MutualContact: mutalContact,
			AccessHash:    do.AccessHash,
			FirstName:     do.FirstName,
			LastName:      do.LastName,
			Username:      do.Username,
			Phone:         do.Phone,
			// TODO(@benqi): Load from db
			Photo:         photo,
			// mtproto.NewTLUserProfilePhotoEmpty().To_UserProfilePhoto(),
			Status:        status,
		}}}
		return data
	}
}

func GetUserByPhoneNumber(selfId int32, phoneNumber string) *userData {
	do := dao.GetUsersDAO(dao.DB_SLAVE).SelectByPhoneNumber(phoneNumber)
	return makeUserDataByDO(selfId, do)
}

func GetMyUserByPhoneNumber(phoneNumber string) *userData {
	do := dao.GetUsersDAO(dao.DB_SLAVE).SelectByPhoneNumber(phoneNumber)
	return makeUserDataByDO(do.Id, do)
}

func GetUserById(selfId int32, userId int32) *userData {
	do := dao.GetUsersDAO(dao.DB_SLAVE).SelectById(userId)
	return makeUserDataByDO(selfId, do)
}

func CreateNewUser(phoneNumber, firstName, lastName string) *mtproto.TLUser {
	// usersDAO := dao.GetUsersDAO(dao.DB_SLAVE)
	do := &dataobject.UsersDO{
		AccessHash:  base.NextSnowflakeId(),
		Phone:       phoneNumber,
		FirstName:   firstName,
		LastName:    lastName,
		CountryCode: "CN",
	}
	do.Id = int32(dao.GetUsersDAO(dao.DB_MASTER).Insert(do))
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

func CreateNewUserPassword(userId int32) {
	// gen server_nonce
	do := &dataobject.UserPasswordsDO{
		UserId:     userId,
		ServerSalt: hex.EncodeToString(crypto.GenerateNonce(8)),
	}
	dao.GetUserPasswordsDAO(dao.DB_MASTER).Insert(do)
}

func CheckAccessHashByUserId(userId int32, accessHash int64) bool {
	params := map[string]interface{}{
		"id":          userId,
		"access_hash": accessHash,
	}
	return dao.GetCommonDAO(dao.DB_SLAVE).CheckExists("users", params)
}
