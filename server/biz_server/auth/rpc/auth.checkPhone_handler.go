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
	"github.com/nebulaim/telegramd/baselib/grpc_util"
	"github.com/nebulaim/telegramd/baselib/logger"
	"github.com/nebulaim/telegramd/biz/base"
	"github.com/nebulaim/telegramd/proto/mtproto"
	"golang.org/x/net/context"
)

// tdesktop客户端会调用，android客户端未使用

// auth.checkPhone#6fe51dfb phone_number:string = auth.CheckedPhone;
func (s *AuthServiceImpl) AuthCheckPhone(ctx context.Context, request *mtproto.TLAuthCheckPhone) (*mtproto.Auth_CheckedPhone, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("auth.checkPhone#6fe51dfb - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	phoneNumber, err := base.CheckAndGetPhoneNumber(request.GetPhoneNumber())
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	registered := s.UserModel.CheckPhoneNumberExist(phoneNumber)
	checkedPhone := mtproto.TLAuthCheckedPhone{Data2: &mtproto.Auth_CheckedPhone_Data{
		PhoneRegistered: mtproto.ToBool(registered),
	}}

	glog.Infof("uth.checkPhone#6fe51dfb - reply: %s\n", checkedPhone)
	return checkedPhone.To_Auth_CheckedPhone(), nil
}
