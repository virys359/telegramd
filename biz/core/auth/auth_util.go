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

package auth

import (
	"github.com/nebulaim/telegramd/biz/dal/dao"
	"github.com/nebulaim/telegramd/biz/dal/dataobject"
)

func CheckBannedByPhoneNumber(phoneNumber string) bool {
	params := map[string]interface{}{
		"phone": phoneNumber,
	}
	return dao.GetCommonDAO(dao.DB_SLAVE).CheckExists("banned", params)
}

func CheckPhoneNumberExist(phoneNumber string) bool {
	params := map[string]interface{}{
		"phone": phoneNumber,
	}
	return dao.GetCommonDAO(dao.DB_SLAVE).CheckExists("users", params)
}

func BindAuthKeyAndUser(authKeyId int64, userId int32) {
	do3 := dao.GetAuthUsersDAO(dao.DB_MASTER).SelectByAuthId(authKeyId)
	if do3 == nil {
	    do3 := &dataobject.AuthUsersDO{
			AuthId: authKeyId,
			UserId: userId,
		}
		dao.GetAuthUsersDAO(dao.DB_MASTER).Insert(do3)
	}
}

/*
  auth.checkedPhone#811ea28e phone_registered:Bool = auth.CheckedPhone;
  auth.sentCode#5e002502 flags:# phone_registered:flags.0?true type:auth.SentCodeType phone_code_hash:string next_type:flags.1?auth.CodeType timeout:flags.2?int = auth.SentCode;
  auth.authorization#cd050916 flags:# tmp_sessions:flags.0?int user:User = auth.Authorization;
  auth.exportedAuthorization#df969c2d id:int bytes:bytes = auth.ExportedAuthorization;
*/
