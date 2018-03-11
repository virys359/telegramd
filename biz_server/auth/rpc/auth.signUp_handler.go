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
	"github.com/nebulaim/telegramd/grpc_util"
	"github.com/nebulaim/telegramd/mtproto"
	"golang.org/x/net/context"
	"github.com/nebulaim/telegramd/biz_model/dal/dao"
	"time"
	"github.com/nebulaim/telegramd/biz_model/dal/dataobject"
	"github.com/nebulaim/telegramd/biz_model/base"
	base2 "github.com/nebulaim/telegramd/baselib/base"
)

// auth.signUp#1b067634 phone_number:string phone_code_hash:string phone_code:string first_name:string last_name:string = auth.Authorization;
func (s *AuthServiceImpl) AuthSignUp(ctx context.Context, request *mtproto.TLAuthSignUp) (*mtproto.Auth_Authorization, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("AuthSignUp - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	// 1. check number
	// 客户端发送的手机号格式为: "+86 111 1111 1111"，归一化
	phoneNumber, err := checkAndGetPhoneNumber(request.GetPhoneNumber())
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	if request.GetPhoneCode() == "" {
		err = mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_PHONE_CODE_EMPTY), "auth.signUp#1b067634: phone code empty")
		glog.Error(err)
		return nil, err
	}

	// TODO(@benqi): regist name ruler
	// check first name invalid
	if request.GetFirstName() == "" {
		err = mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_FIRSTNAME_INVALID), "auth.signUp#1b067634: first name invalid")
		glog.Error(err)
		return nil, err
	}

	// check first name invalid
	//if request.GetLastName() == "" {
	//	err = mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_LASTNAME_INVALID), "auth.signUp#1b067634: last name invalid")
	//	glog.Error(err)
	//	return nil, err
	//}

	// <string name="PhoneNumberFlood">Sorry, you have deleted and re-created your account too many times recently.
	//    Please wait for a few days before signing up again.</string>
	//
	userDO := dao.GetUsersDAO(dao.DB_SLAVE).SelectByPhoneNumber(phoneNumber)
	if userDO != nil {
		err = mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_PHONE_NUMBER_OCCUPIED), "auth.signUp#1b067634: phone number occuiped.")
		glog.Error(err)
		return nil, err
	}

	// check
	// PHONE_NUMBER_INVALID
	// PHONE_CODE_EMPTY, PHONE_CODE_INVALID
	// PHONE_CODE_EXPIRED
	// FIRSTNAME_INVALID
	// LASTNAME_INVALID

	// Check code
	authPhoneTransactionsDAO := dao.GetAuthPhoneTransactionsDAO(dao.DB_SLAVE)
	do1 := authPhoneTransactionsDAO.SelectByPhoneCodeHash(request.PhoneCodeHash, phoneNumber)
	if do1 == nil {
		err = mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_PHONE_CODE_HASH_EMPTY), "auth.signUp#1b067634: phone code hash invalid.")
		glog.Error(err)
		return nil, err
	}

	if do1.Code != request.GetPhoneCode() {
		err = mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_PHONE_CODE_INVALID), "auth.signUp#1b067634: phone code invalid.")
		glog.Error(err)
		return nil, err
	}

	if time.Now().Unix() > do1.CreatedTime + 15*60 {
		err = mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_PHONE_CODE_EXPIRED), "auth.signUp#1b067634: phone code expired.")
		glog.Error(err)
		return nil, err
	}

	// usersDAO := dao.GetUsersDAO(dao.DB_SLAVE)
	userDO = &dataobject.UsersDO{
		AccessHash:  base.NextSnowflakeId(),
		Phone:       request.GetPhoneNumber(),
		FirstName:   request.GetFirstName(),
		LastName:    request.GetLastName(),
		CountryCode: "CN",
		CreatedAt:   base2.NowFormatYMDHMS(),
	}
	userId := dao.GetUsersDAO(dao.DB_MASTER).Insert(userDO)

	do3 := &dataobject.AuthUsersDO{
		AuthId: md.AuthId,
		UserId: int32(userId),
	}
	dao.GetAuthUsersDAO(dao.DB_MASTER).Insert(do3)

	// TODO(@benqi): 从数据库加载
	authAuthorization := mtproto.NewTLAuthAuthorization()
	user := mtproto.NewTLUser()

	user.SetSelf(true)
	user.SetId(int32(userId))
	user.SetAccessHash(userDO.AccessHash)
	user.SetFirstName(userDO.FirstName)
	user.SetLastName(userDO.LastName)
	user.SetUsername(userDO.Username)
	user.SetPhone(phoneNumber)

	authAuthorization.SetUser(user.To_User())

	glog.Infof("auth.signUp#1b067634 - reply: %s\n", logger.JsonDebugData(authAuthorization))
	return authAuthorization.To_Auth_Authorization(), nil
}
