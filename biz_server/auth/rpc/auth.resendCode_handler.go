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

	// 2. sentCode
	lastCreatedAt := time.Unix(time.Now().Unix()-15*60, 0).Format("2006-01-02 15:04:05")
	do := dao.GetAuthPhoneTransactionsDAO(dao.DB_SLAVE).SelectByPhoneCodeHash(request.GetPhoneCodeHash(), phoneNumber, lastCreatedAt)
	if do == nil {
		err = mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_PHONE_CODE_EXPIRED), "invalid phone number")
	}

	return nil, fmt.Errorf("Not impl AuthResendCode")
}
