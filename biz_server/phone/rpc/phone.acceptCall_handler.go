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
	"github.com/nebulaim/telegramd/baselib/base"
	"github.com/nebulaim/telegramd/biz/core/user"
	update2 "github.com/nebulaim/telegramd/biz/core/update"
)

/*
body: { phone_acceptCall
  peer: { inputPhoneCall
	id: 1926738269689442709 [LONG],
	access_hash: 7643618436880411068 [LONG],
  },
  g_b: 49 7E A2 C4 CE 6C 28 2E 19 13 C9 95 0F 71 34 FA... [256 BYTES],
  protocol: { phoneCallProtocol
	flags: 3 [INT],
	udp_p2p: YES [ BY BIT 0 IN FIELD flags ],
	udp_reflector: YES [ BY BIT 1 IN FIELD flags ],
	min_layer: 65 [INT],
	max_layer: 65 [INT],
  },
},

 */
// https://core.telegram.org/api/end-to-end/voice-calls
//
// B accepts the call on one of their devices,
// stores the received value of g_a_hash for this instance of the voice call creation protocol,
// chooses a random value of b, 1 < b < p-1, computes g_b:=power(g,b) mod p,
// performs all the required security checks, and invokes the phone.acceptCall method,
// which has a g_b:bytes field (among others), to be filled with the value of g_b itself (not its hash).
//
// The Server S sends an updatePhoneCall with the phoneCallDiscarded constructor to all other devices B has authorized,
// to prevent accepting the same call on any of the other devices. From this point on,
// the server S works only with that of B's devices which has invoked phone.acceptCall first.
//
// phone.acceptCall#3bd2b4a0 peer:InputPhoneCall g_b:bytes protocol:PhoneCallProtocol = phone.PhoneCall;
func (s *PhoneServiceImpl) PhoneAcceptCall(ctx context.Context, request *mtproto.TLPhoneAcceptCall) (*mtproto.Phone_PhoneCall, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("PhoneAcceptCall - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	peer := request.GetPeer().To_InputPhoneCall()
	// TODO(@benqi): check peer
	callSession, _ := phoneCallSessionManager[peer.GetId()]
	callSession.g_b = request.GetGB()

	//if !ok {
	//	glog.Infof("PhoneReceivedCall - invalid peer: {%v}", peer)
	//	// return mtproto.ToBool(false), nil
	//}
	//if peer.GetAccessHash() != callSession.participantAccessHash {
	//	glog.Infof("PhoneReceivedCall - invalid peer: {%v}", peer)
	//	// return mtproto.ToBool(false), nil
	//}

	// now := int32(time.Now().Unix())
	// participantId := request.GetPeer().GetData2().GetId()

	updates := mtproto.NewTLUpdates()

	updateUserStatus := mtproto.NewTLUpdateUserStatus()
	updateUserStatus.SetUserId(callSession.participantId)
	status := mtproto.NewTLUserStatusOnline()
	status.SetExpires(callSession.date+300)
	updateUserStatus.SetStatus(status.To_UserStatus())
	updates.Data2.Updates = append(updates.Data2.Updates, updateUserStatus.To_Update())

	updatePhoneCall := mtproto.NewTLUpdatePhoneCall()
	// push to participant_id
	phoneCallAccepted := mtproto.NewTLPhoneCallAccepted()
	phoneCallAccepted.SetId(callSession.id)
	phoneCallAccepted.SetAccessHash(callSession.participantAccessHash)
	phoneCallAccepted.SetDate(callSession.date)
	phoneCallAccepted.SetAdminId(callSession.adminId)
	phoneCallAccepted.SetParticipantId(callSession.participantId)
	phoneCallAccepted.SetGB(request.GetGB())
	phoneCallAccepted.SetProtocol(callSession.protocol.To_PhoneCallProtocol())
	updatePhoneCall.SetPhoneCall(phoneCallAccepted.To_PhoneCall())

	updates.Data2.Updates = append(updates.Data2.Updates, updatePhoneCall.To_Update())

	participantIds := []int32{callSession.adminId, callSession.participantId}
	users := user.GetUserList(participantIds)
	for _, u := range users {
		if u.GetId() == md.UserId {
			u.SetSelf(false)
		} else {
			u.SetSelf(true)
		}
		// Add users
		updates.Data2.Users = append(updates.Data2.Users, u.To_User())
	}

	seq := int32(update2.NextSeqId(base.Int32ToString(callSession.adminId)))
	updates.SetSeq(seq)
	updates.SetDate(callSession.date)

	//delivery.GetDeliveryInstance().DeliveryUpdatesNotMe(
	//	md.AuthId,
	//	md.SessionId,
	//	md.NetlibSessionId,
	//	[]int32{callSession.adminId},
	//	updates.To_Updates().Encode())

	// TODO(@benqi): delivery to other phoneCallDiscarded
	// 返回给客户端

	// 2. reply
	phoneCall := mtproto.NewTLPhonePhoneCall()

	phoneCallWaiting := mtproto.NewTLPhoneCallWaiting()
	phoneCallWaiting.SetDate(callSession.date)
	phoneCallWaiting.SetId(callSession.id)
	phoneCallWaiting.SetAccessHash(callSession.participantAccessHash)
	phoneCallWaiting.SetAdminId(callSession.adminId)
	phoneCallWaiting.SetParticipantId(callSession.participantId)
	phoneCallWaiting.SetReceiveDate(int32(time.Now().Unix()))
	// TODO(@benqi): use request.GetProtocol() or callSession.protocol??
	// phoneCallWaiting.SetProtocol(callSession.protocol.To_PhoneCallProtocol())
	phoneCallWaiting.SetProtocol(request.GetProtocol())
	// SetPhoneCall
	phoneCall.SetPhoneCall(phoneCallWaiting.To_PhoneCall())

	for _, u := range users {
		if u.GetId() == md.UserId {
			u.SetSelf(true)
		} else {
			u.SetSelf(false)
		}
		// Add users
		phoneCall.Data2.Users = append(phoneCall.Data2.Users, u.To_User())
	}

	glog.Infof("PhoneAcceptCall - reply: {%v}", phoneCall)
	return phoneCall.To_Phone_PhoneCall(), nil
}
