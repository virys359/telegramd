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
	"github.com/nebulaim/telegramd/biz/base"
	"github.com/nebulaim/telegramd/biz/core/user"
	"github.com/nebulaim/telegramd/biz/core/auth"
)

// auth.resendCode#3ef1a9bf phone_number:string phone_code_hash:string = auth.SentCode;
func (s *AuthServiceImpl) AuthResendCode(ctx context.Context, request *mtproto.TLAuthResendCode) (*mtproto.Auth_SentCode, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("AuthResendCode - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	// 1. check number
	// 客户端发送的手机号格式为: "+86 111 1111 1111"，归一化
	phoneNumber, err := base.CheckAndGetPhoneNumber(request.GetPhoneNumber())
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	// <string name="PhoneNumberFlood">Sorry, you have deleted and re-created your account too many times recently.
	//    Please wait for a few days before signing up again.</string>
	//
	banned := user.CheckBannedByPhoneNumber(phoneNumber)
	if !banned {
		err = mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_PHONE_NUMBER_BANNED), "auth.sendCode#86aef0ec: phone number banned.")
		return nil, err
	}

	code, err := auth.GetCodeByCodeHash(request.GetPhoneCodeHash(), phoneNumber)
	if err != nil {
		return nil, err
	}

	// 如果手机号已经注册，检查是否有其他设备在线，有则使用sentCodeTypeApp
	// 否则使用sentCodeTypeSms
	// TODO(@benqi): 有则使用sentCodeTypeFlashCall和entCodeTypeCall？？

	// 使用app类型，code统一为123456
	phoneRegistered := user.CheckPhoneNumberExist(phoneNumber)

	authSentCode := &mtproto.TLAuthSentCode{Data2: &mtproto.Auth_SentCode_Data{
		PhoneRegistered: phoneRegistered,
		PhoneCodeHash:   code.CodeHash,
		Timeout:         60,	// TODO(@benqi): 默认60s
	}}

	if phoneRegistered {
		// TODO(@benqi): check other session online
		authSentCodeType := &mtproto.TLAuthSentCodeTypeApp{Data2: &mtproto.Auth_SentCodeType_Data{
			Length: code.GetPhoneCodeLength(),
		}}
		authSentCode.SetType(authSentCodeType.To_Auth_SentCodeType())
	} else {
		// TODO(@benqi): sentCodeTypeFlashCall and sentCodeTypeCall, nextType
		authSentCodeType := &mtproto.TLAuthSentCodeTypeSms{Data2: &mtproto.Auth_SentCodeType_Data{
			Length: code.GetPhoneCodeLength(),
		}}
		authSentCode.SetType(authSentCodeType.To_Auth_SentCodeType())

		// TODO(@benqi): nextType
		// authSentCode.SetNextType()
	}

	glog.Infof("auth.resendCode#3ef1a9bf - reply: %s", logger.JsonDebugData(authSentCode))
	return authSentCode.To_Auth_SentCode(), nil
}
