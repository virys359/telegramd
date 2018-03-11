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
	"github.com/nebulaim/telegramd/biz_model/model"
	"github.com/nebulaim/telegramd/biz_server/delivery"
	"github.com/nebulaim/telegramd/baselib/base"
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

	// TODO(@benqi): check peer
	peer := request.GetPeer().To_InputPhoneCall()
	callSession, ok := phoneCallSessionManager[peer.GetId()]
	if !ok {
		glog.Infof("PhoneReceivedCall - invalid peer: {%v}", peer)
		return mtproto.ToBool(false), nil
	}
	if peer.GetAccessHash() != callSession.participantAccessHash {
		glog.Infof("PhoneReceivedCall - invalid peer: {%v}", peer)
		return mtproto.ToBool(false), nil
	}

	// Delivery to adminId
	updates := mtproto.NewTLUpdates()

	phoneCall := mtproto.NewTLPhonePhoneCall()
	phoneCallWaiting := mtproto.NewTLPhoneCallWaiting()
	phoneCallWaiting.SetDate(callSession.date)
	phoneCallWaiting.SetId(callSession.id)
	phoneCallWaiting.SetAccessHash(callSession.adminAccessHash)
	phoneCallWaiting.SetAdminId(callSession.adminId)
	phoneCallWaiting.SetParticipantId(callSession.participantId)
	phoneCallWaiting.SetProtocol(callSession.protocol.To_PhoneCallProtocol())
	phoneCallWaiting.SetReceiveDate(int32(time.Now().Unix()))
	// SetPhoneCall
	phoneCall.SetPhoneCall(phoneCallWaiting.To_PhoneCall())

	participantIds := []int32{callSession.adminId, callSession.participantId}
	users := model.GetUserModel().GetUserList(participantIds)
	for _, u := range users {
		if u.GetId() != md.UserId {
			u.SetSelf(true)
		} else {
			u.SetSelf(false)
		}
		// Add users
		updates.Data2.Users = append(updates.Data2.Users, u.To_User())
	}

	for _, u := range users {
		if u.GetId() == md.UserId {
			u.SetSelf(true)
		} else {
			u.SetSelf(false)
		}
		// Add users
		phoneCall.Data2.Users = append(phoneCall.Data2.Users, u.To_User())
	}

	// TODO(@benqi): seq
	seq := int32(model.GetSequenceModel().NextSeqId(base.Int32ToString(callSession.adminId)))
	updates.SetSeq(seq)
	updates.SetDate(callSession.date)

	delivery.GetDeliveryInstance().DeliveryUpdatesNotMe(
		md.AuthId,
		md.SessionId,
		md.NetlibSessionId,
		[]int32{callSession.adminId},
		updates.To_Updates().Encode())
	// 返回给客户端

	glog.Infof("PhoneReceivedCall - reply {true}")
	return mtproto.ToBool(true), nil
}
