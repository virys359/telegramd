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
	"github.com/nebulaim/telegramd/baselib/base"
	"github.com/nebulaim/telegramd/biz/core/user"
	update2 "github.com/nebulaim/telegramd/biz/core/update"
)

/*
 phone_call: { phoneCallDiscarded
  flags: 11 [INT],
  need_rating: [ SKIPPED BY BIT 2 IN FIELD flags ],
  need_debug: YES [ BY BIT 3 IN FIELD flags ],
  id: 1926738269646495698 [LONG],
  reason: { phoneCallDiscardReasonDisconnect },
  duration: 3 [INT],
},
 */

// phone.discardCall#78d413a6 peer:InputPhoneCall duration:int reason:PhoneCallDiscardReason connection_id:long = Updates;
func (s *PhoneServiceImpl) PhoneDiscardCall(ctx context.Context, request *mtproto.TLPhoneDiscardCall) (*mtproto.Updates, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("PhoneDiscardCall - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	// TODO(@benqi): Impl PhoneDiscardCall logic
	peer := request.GetPeer().To_InputPhoneCall()
	callSession, _ := phoneCallSessionManager[peer.GetId()]

	phoneCallDiscarded := mtproto.NewTLPhoneCallDiscarded()
	phoneCallDiscarded.SetId(callSession.id)
	phoneCallDiscarded.SetNeedDebug(true)
	phoneCallDiscarded.SetReason(request.GetReason())
	phoneCallDiscarded.SetDuration(request.GetDuration())

	// TODO(@benqi): notify relay_server's connection_id

	participantIds := []int32{callSession.adminId, callSession.participantId}
	users := user.GetUserList(participantIds)

	// delivey
	updates := mtproto.NewTLUpdates()

	updatePhoneCall := mtproto.NewTLUpdatePhoneCall()
	updatePhoneCall.SetPhoneCall(phoneCallDiscarded.To_PhoneCall())
	updates.Data2.Updates = append(updates.Data2.Updates, updatePhoneCall.To_Update())

	for _, u := range users {
		if u.GetId() == md.UserId {
			u.SetSelf(false)
		} else {
			u.SetSelf(true)
		}
		updates.Data2.Users = append(updates.Data2.Users, u.To_User())
	}

	// TODO(@benqi): seq
	seq := int32(update2.NextSeqId(base.Int32ToString(callSession.adminId)))
	updates.SetSeq(seq)
	updates.SetDate(callSession.date)

	//var toId int32 = md.UserId
	//if md.UserId == callSession.adminId {
	//	toId = callSession.participantId
	//}

	//delivery.GetDeliveryInstance().DeliveryUpdatesNotMe(
	//	md.AuthId,
	//	md.SessionId,
	//	md.NetlibSessionId,
	//	[]int32{toId},
	//	updates.To_Updates().Encode())

	for _, u := range users {
		if u.GetId() == md.UserId {
			u.SetSelf(true)
		} else {
			u.SetSelf(false)
		}
	}

	glog.Infof("PhoneDiscardCall - reply {%v}", updates)
	return updates.To_Updates(), nil
}
