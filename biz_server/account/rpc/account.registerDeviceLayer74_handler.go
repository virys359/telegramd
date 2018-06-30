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
	"github.com/nebulaim/telegramd/baselib/grpc_util"
	"github.com/nebulaim/telegramd/mtproto"
	"golang.org/x/net/context"
	"github.com/nebulaim/telegramd/biz/core/account"
	"github.com/nebulaim/telegramd/biz/core"
)

// Layer74
// account.registerDevice#1389cc token_type:int token:string app_sandbox:Bool other_uids:Vector<int> = Bool;
func (s *AccountServiceImpl) AccountRegisterDeviceLayer74(ctx context.Context, request *mtproto.TLAccountRegisterDeviceLayer74) (*mtproto.Bool, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("account.registerDevice#637ea878 - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	// Check token format by token_type
	// TODO(@benqi): check token format by token_type
	if request.Token == "" {
		err := mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_BAD_REQUEST)
		glog.Error(err)
		return nil, err
	}

	// TODO(@benqi): check toke_type invalid
	if request.TokenType < core.TOKEN_TYPE_APNS || request.TokenType > core.TOKEN_TYPE_MAXSIZE {
		err := mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_BAD_REQUEST)
		glog.Error(err)
		return nil, err
	}

	registered := account.RegisterDevice(md.AuthId, md.UserId, int8(request.TokenType), request.Token)

	glog.Infof("account.registerDevice#637ea878 - reply: {true}")
	return mtproto.ToBool(registered), nil
}
