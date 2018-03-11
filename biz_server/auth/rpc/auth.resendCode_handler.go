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
	"fmt"
	"github.com/golang/glog"
	"github.com/nebulaim/telegramd/base/logger"
	"github.com/nebulaim/telegramd/grpc_util"
	"github.com/nebulaim/telegramd/mtproto"
	"golang.org/x/net/context"
	"github.com/nebulaim/telegramd/biz_model/dal/dao"
	"time"
)

// auth.resendCode#3ef1a9bf phone_number:string phone_code_hash:string = auth.SentCode;
func (s *AuthServiceImpl) AuthResendCode(ctx context.Context, request *mtproto.TLAuthResendCode) (*mtproto.Auth_SentCode, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("AuthResendCode - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	// 1. check number
	// 客户端发送的手机号格式为: "+86 111 1111 1111"，归一化
	phoneNumber, err := checkAndGetPhoneNumber(request.GetPhoneNumber())
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	// <string name="PhoneNumberFlood">Sorry, you have deleted and re-created your account too many times recently.
	//    Please wait for a few days before signing up again.</string>
	//
	userDO := dao.GetUsersDAO(dao.DB_SLAVE).SelectByPhoneNumber(phoneNumber)
	if userDO != nil && userDO.Banned != 0 {
		err = mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_PHONE_NUMBER_BANNED), "auth.sendCode#86aef0ec: phone number banned.")
		return nil, err
	}

	// 2. sentCode
	// lastCreatedAt := time.Unix(time.Now().Unix()-15*60, 0).Format("2006-01-02 15:04:05")
	do := dao.GetAuthPhoneTransactionsDAO(dao.DB_SLAVE).SelectByPhoneCodeHash(request.GetPhoneCodeHash(), phoneNumber)
	if do == nil {
		err = mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_PHONE_CODE_HASH_EMPTY), "invalid phone number")
		glog.Error(err)
		return nil, err
	}

	if time.Now().Unix() > do.CreatedTime + 15*60 {
		err = mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_PHONE_CODE_EXPIRED), "code expired")
		glog.Error(err)
		return nil, err
	}

	if do.Attempts > 3 {
		// TODO(@benqi): 输入了太多次错误的phone code
		err = mtproto.NewFloodWaitX(15*60, "auth.sendCode#86aef0ec: too many attempts.")
		return nil, err
	}

	// 如果手机号已经注册，检查是否有其他设备在线，有则使用sentCodeTypeApp
	// 否则使用sentCodeTypeSms
	// TODO(@benqi): 有则使用sentCodeTypeFlashCall和entCodeTypeCall？？

	// 使用app类型，code统一为123456
	authSentCode := mtproto.NewTLAuthSentCode()

	phoneRegistered := userDO != nil
	authSentCode.SetPhoneRegistered(phoneRegistered)

	if phoneRegistered {
		// TODO(@benqi): check other session online
		authSentCodeType := mtproto.NewTLAuthSentCodeTypeApp()
		authSentCodeType.SetLength(6)
		authSentCode.SetType(authSentCodeType.To_Auth_SentCodeType())
	} else {
		// TODO(@benqi): sentCodeTypeFlashCall and sentCodeTypeCall, nextType
		authSentCodeType := mtproto.NewTLAuthSentCodeTypeSms()
		authSentCodeType.SetLength(6)
		authSentCode.SetType(authSentCodeType.To_Auth_SentCodeType())

		// TODO(@benqi): nextType
		// authSentCode.SetNextType()
	}

	authSentCode.SetPhoneCodeHash(do.TransactionHash)

	// TODO(@benqi): 默认60s
	authSentCode.SetTimeout(60)

	glog.Infof("auth.resendCode#3ef1a9bf - reply: %s", logger.JsonDebugData(authSentCode))
	return authSentCode.To_Auth_SentCode(), nil
}
