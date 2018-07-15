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
	"github.com/nebulaim/telegramd/proto/mtproto"
	"golang.org/x/net/context"
	"github.com/nebulaim/telegramd/biz/core/account"
	user2 "github.com/nebulaim/telegramd/biz/core/user"
	"github.com/nebulaim/telegramd/server/sync/sync_client"
	updates2 "github.com/nebulaim/telegramd/biz/core/update"
)

// account.setPrivacy#c9f81ce8 key:InputPrivacyKey rules:Vector<InputPrivacyRule> = account.PrivacyRules;
func (s *AccountServiceImpl) AccountSetPrivacy(ctx context.Context, request *mtproto.TLAccountSetPrivacy) (*mtproto.Account_PrivacyRules, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("account.setPrivacy#c9f81ce8 - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	// TODO(@benqi): Check request valid.

	key := account.FromInputPrivacyKey(request.GetKey())

	privacyLogic := account.MakePrivacyLogic(md.UserId)
	rulesData := privacyLogic.SetPrivacy(key, request.GetRules())

	var rules *mtproto.TLAccountPrivacyRules
	idList := rulesData.PickAllUserIdList()
	ruleList := rulesData.ToPrivacyRuleList()

	/////////////////////////////////////////////////////////////////////////////////
	// Sync unblocked: updateUserBlocked
	updatePrivacy := &mtproto.TLUpdatePrivacy{Data2: &mtproto.Update_Data{
		Key: key.ToPrivacyKey(),
		Rules: ruleList,
	}}

	unBlockedUpdates := updates2.NewUpdatesLogic(md.UserId)
	unBlockedUpdates.AddUpdate(updatePrivacy.To_Update())

	if len(idList) == 0 {
		rules = &mtproto.TLAccountPrivacyRules{ Data2: &mtproto.Account_PrivacyRules_Data{
			Rules: ruleList,
		}}
	} else {
		users := user2.GetUsersBySelfAndIDList(md.UserId, idList)
		rules = &mtproto.TLAccountPrivacyRules{ Data2: &mtproto.Account_PrivacyRules_Data{
			Rules: ruleList,
			Users: users,
		}}
		unBlockedUpdates.AddUsers(users)
	}

	// TODO(@benqi): handle seq
	sync_client.GetSyncClient().SyncUpdatesData(md.AuthId, md.SessionId, md.UserId, unBlockedUpdates.ToUpdates())

	glog.Infof("account.setPrivacy#c9f81ce8 - reply: %s", logger.JsonDebugData(rules))
	return rules.To_Account_PrivacyRules(), nil
}
