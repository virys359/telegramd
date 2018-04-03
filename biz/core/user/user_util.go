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

func GetUserByPhoneNumber(self bool, phoneNumber string) *userData {
	do := dao.GetUsersDAO(dao.DB_SLAVE).SelectByPhoneNumber(phoneNumber)
	if do == nil {
		return nil
	} else {
		var (
			contact = false
			mutalContact = false
			status *mtproto.UserStatus
		)

		if self {
			contact = true
			mutalContact = true
			status = makeUserStatusOnline()
		} else {
			// TODO(@benqi): check contacts, getUserStatus
		}

		// Status: &mtproto.T
		data := &userData{ TLUser: &mtproto.TLUser{ Data2: &mtproto.User_Data{
			Id:            do.Id,
			Self:          self,
			Contact:       contact,
			MutualContact: mutalContact,
			AccessHash:    do.AccessHash,
			FirstName:     do.FirstName,
			LastName:      do.LastName,
			Username:      do.Username,
			Phone:         phoneNumber,
			// TODO(@benqi): Load from db
			Photo:         mtproto.NewTLUserProfilePhotoEmpty().To_UserProfilePhoto(),
			Status:        status,
		}}}
		return data
	}
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
