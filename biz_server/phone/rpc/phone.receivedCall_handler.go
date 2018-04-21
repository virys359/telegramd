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
	"time"
	update2 "github.com/nebulaim/telegramd/biz/core/update"
	"github.com/nebulaim/telegramd/biz/core/user"
	"github.com/nebulaim/telegramd/biz/core/phone_call"
	"fmt"
	"github.com/nebulaim/telegramd/biz_server/sync_client"
)

/*
	body: { phone_receivedCall
	  peer: { inputPhoneCall
		id: 1926738269689442709 [LONG],
		access_hash: 7643618436880411068 [LONG],
	  },
	},
 */
// phone.receivedCall#17d54f61 peer:InputPhoneCall = Bool;
func (s *PhoneServiceImpl) PhoneReceivedCall(ctx context.Context, request *mtproto.TLPhoneReceivedCall) (*mtproto.Bool, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("PhoneReceivedCall - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	//// TODO(@benqi): check peer
	peer := request.GetPeer().To_InputPhoneCall()

	callSession, err := phone_call.MakePhoneCallLogcByLoad(peer.GetId())
	if err != nil {
		glog.Errorf("invalid peer: {%v}, err: %v", peer, err)
		return nil, err
	}
	if peer.GetAccessHash() != callSession.ParticipantAccessHash {
		err = fmt.Errorf("invalid peer: {%v}", peer)
		glog.Errorf("invalid peer: {%v}", peer)
		return nil, err
	}

	/////////////////////////////////////////////////////////////////////////////////
	updatesData := update2.NewUpdatesLogic(md.UserId)

	// 1. add updateUserStatus
	//var status *mtproto.UserStatus
	statusOnline := &mtproto.TLUserStatusOnline{Data2: &mtproto.UserStatus_Data{
		Expires: int32(time.Now().Unix() + 5*30),
	}}
	// status = statusOnline.To_UserStatus()
	updateUserStatus := &mtproto.TLUpdateUserStatus{Data2: &mtproto.Update_Data{
		UserId: md.UserId,
		Status: statusOnline.To_UserStatus(),
	}}
	updatesData.AddUpdate(updateUserStatus.To_Update())

	// 2. add phoneCallRequested
	updatePhoneCall := &mtproto.TLUpdatePhoneCall{Data2: &mtproto.Update_Data{
		PhoneCall: callSession.ToPhoneCallWaiting(callSession.AdminId, int32(time.Now().Unix())).To_PhoneCall(),
	}}
	updatesData.AddUpdate(updatePhoneCall.To_Update())

	// 3. add users
	updatesData.AddUsers(user.GetUsersBySelfAndIDList(callSession.AdminId, []int32{md.UserId, callSession.AdminId}))

	sync_client.GetSyncClient().PushToUserUpdatesData(callSession.AdminId, updatesData.ToUpdates())

	/////////////////////////////////////////////////////////////////////////////////
	glog.Infof("PhoneReceivedCall - reply {true}")
	return mtproto.ToBool(true), nil
}
