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

// auth.sendCode#86aef0ec flags:# allow_flashcall:flags.0?true phone_number:string current_number:flags.0?Bool api_id:int api_hash:string = auth.SentCode;
func (s *AuthServiceImpl) AuthSendCode(ctx context.Context, request *mtproto.TLAuthSendCode) (*mtproto.Auth_SentCode, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("AuthSendCode - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	// TODO(@benqi): 接入telegram网络首先必须申请api_id和api_hash，验证api_id和api_hash是否合法
	// 1. check api_id and api_hash

	//// 3. check number
	//// 客户端发送的手机号格式为: "+86 111 1111 1111"，归一化
	phoneNumber, err :=  base.CheckAndGetPhoneNumber(request.GetPhoneNumber())
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	// 2. check allow_flashcall and current_number
	// CurrentNumber: 是否为本机电话号码

	// if allow_flashcall is true then current_number is true
	currentNumber := mtproto.FromBool(request.GetCurrentNumber())
	if !currentNumber && request.GetAllowFlashcall() {
		err = mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_BAD_REQUEST), "auth.sendCode#86aef0ec: current_number is true but allow_flashcall is false.")
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

	// TODO(@benqi): MIGRATE datacenter
	// https://core.telegram.org/api/datacenter
	// The auth.sendCode method is the basic entry point when registering a new user or authorizing an existing user.
	//   95% of all redirection cases to a different DC will occure when invoking this method.
	//
	// The client does not yet know which DC it will be associated with; therefore,
	//   it establishes an encrypted connection to a random address and sends its query to that address.
	// Having received a phone_number from a client,
	// 	 we can find out whether or not it is registered in the system.
	//   If it is, then, if necessary, instead of sending a text message,
	//   we request that it establish a connection with a different DC first (PHONE_MIGRATE_X error).
	// If we do not yet have a user with this number, we examine its IP-address.
	//   We can use it to identify the closest DC.
	//   Again, if necessary, we redirect the user to a different DC (NETWORK_MIGRATE_X error).
	//
	//if userDO == nil {
	//	// phone registered
	//	// TODO(@benqi): 由phoneNumber和ip优选
	//} else {
	//	// TODO(@benqi): 由userId优选
	//}

	// 检查phoneNumber是否异常
	// TODO(@benqi): 定义恶意登录规则
	// PhoneNumberFlood
	// FLOOD_WAIT

	// TODO(@benqi): 独立出统一消息推送系统
	// 检查phpne是否存在，若存在是否在线决定是否通过短信发送或通过其他客户端发送
	// 透传AuthId，UserId，终端类型等
	// 检查满足条件的TransactionHash是否存在，可能的条件：
	//  1. is_deleted !=0 and now - created_at < 15 分钟
	//
	//  auth.sentCodeTypeApp#3dbb5986 length:int = auth.SentCodeType;
	//  auth.sentCodeTypeSms#c000bba2 length:int = auth.SentCodeType;
	//  auth.sentCodeTypeCall#5353e5a7 length:int = auth.SentCodeType;
	//  auth.sentCodeTypeFlashCall#ab03c6d9 pattern:string = auth.SentCodeType;
	//

	// TODO(@benqi): 限制同一个authKeyId
	// TODO(@benqi): 使用redis

	//// 15分钟内有效
	code := auth.GetOrCreateCode(md.AuthId, phoneNumber, request.ApiId, request.ApiHash)
	//	if do.Attempts > 3 {
	//		// TODO(@benqi): 输入了太多次错误的phone code
	//		err = mtproto.NewFloodWaitX(15*60, "auth.sendCode#86aef0ec: too many attempts.")
	//		return nil, err
	//	}

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

	glog.Infof("AuthSendCode - reply: %s", logger.JsonDebugData(authSentCode))
	return authSentCode.To_Auth_SentCode(), nil
}
