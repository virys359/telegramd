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
	"github.com/nebulaim/telegramd/base/logger"
	"github.com/nebulaim/telegramd/grpc_util"
	"github.com/nebulaim/telegramd/mtproto"
	"golang.org/x/net/context"
	"time"
	"github.com/nebulaim/telegramd/biz_model/model"
	"github.com/nebulaim/telegramd/base/base"
	"github.com/nebulaim/telegramd/biz_server/delivery"
)

/*
	body: { phone_confirmCall
	  peer: { inputPhoneCall
		id: 1926738269646495698 [LONG],
		access_hash: 12281641976525210193 [LONG],
	  },
	  g_a: 6D E4 02 D2 13 3E 55 D1 4E 4F D1 73 FC CA B1 3F... [256 BYTES],
	  key_fingerprint: 12060240716975480264 [LONG],
	  protocol: { phoneCallProtocol
		flags: 3 [INT],
		udp_p2p: YES [ BY BIT 0 IN FIELD flags ],
		udp_reflector: YES [ BY BIT 1 IN FIELD flags ],
		min_layer: 65 [INT],
		max_layer: 65 [INT],
	  },
	},

 respone:
      phone_call: { phoneCall
        id: 1926738269646495698 [LONG],
        access_hash: 12281641976525210193 [LONG],
        date: 1514943491 [INT],
        admin_id: 448603711 [INT],
        participant_id: 264696845 [INT],
        g_a_or_b: 7B 68 CA 23 C9 72 DA 94 88 15 82 48 ED EB A2 18... [256 BYTES],
        key_fingerprint: 12060240716975480264 [LONG],
        protocol: { phoneCallProtocol
          flags: 3 [INT],
          udp_p2p: YES [ BY BIT 0 IN FIELD flags ],
          udp_reflector: YES [ BY BIT 1 IN FIELD flags ],
          min_layer: 65 [INT],
          max_layer: 65 [INT],
        },
        connection: { phoneConnection
          id: 50003 [LONG],
          ip: "91.108.16.3" [STRING],
          ipv6: "2a00:ab00:300:3bc::20" [STRING],
          port: 532 [INT],
          peer_tag: "24ffcbeb7980d28b" [STRING],
        },
 */

// phone.confirmCall#2efe1722 peer:InputPhoneCall g_a:bytes key_fingerprint:long protocol:PhoneCallProtocol = phone.PhoneCall;
func (s *PhoneServiceImpl) PhoneConfirmCall(ctx context.Context, request *mtproto.TLPhoneConfirmCall) (*mtproto.Phone_PhoneCall, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("PhoneConfirmCall - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	peer := request.GetPeer().To_InputPhoneCall()
	now := int32(time.Now().Unix())

	// TODO(@benqi): check peer invalid
	callSession, _ := phoneCallSessionManager[peer.GetId()]
	callSession.g_a = request.GetGA()

	phoneCall := mtproto.NewTLPhoneCall()
	phoneCall.SetId(callSession.id)
	phoneCall.SetAccessHash(callSession.participantAccessHash)
	phoneCall.SetDate(now)
	phoneCall.SetAdminId(callSession.adminId)
	phoneCall.SetParticipantId(callSession.participantId)
	phoneCall.SetGAOrB(callSession.g_a)
	phoneCall.SetKeyFingerprint(int64(fingerprint))
	// TODO(@benqi): protocol:PhoneCallProtocol or callSession.protocol??
	phoneCall.SetProtocol(callSession.protocol.To_PhoneCallProtocol())
	// TODO(@benqi): connection config!
	connection := mtproto.NewTLPhoneConnection()
	connection.SetId(50003)
	connection.SetIp("127.0.0.1")
	connection.SetIpv6("")
	connection.SetPort(532)
	connection.SetPeerTag([]byte("24ffcbeb7980d28b"))
	phoneCall.SetConnection(connection.To_PhoneConnection())
	phoneCall.SetStartDate(0)

	// alternative_connections := make([]*mtproto.PhoneConnection, 0)
	participantIds := []int32{callSession.adminId, callSession.participantId}
	users := model.GetUserModel().GetUserList(participantIds)

	// delivey
	updates := mtproto.NewTLUpdates()

	updatePhoneCall := mtproto.NewTLUpdatePhoneCall()
	updatePhoneCall.SetPhoneCall(phoneCall.To_PhoneCall())
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
	seq := int32(model.GetSequenceModel().NextSeqId(base.Int32ToString(callSession.adminId)))
	updates.SetSeq(seq)
	updates.SetDate(callSession.date)

	delivery.GetDeliveryInstance().DeliveryUpdatesNotMe(
		md.AuthId,
		md.SessionId,
		md.NetlibSessionId,
		[]int32{callSession.adminId},
		updates.To_Updates().Encode())

	// reply  phone.PhoneCall
	phone_PoneCall := mtproto.NewTLPhonePhoneCall()
	phoneCall.SetAccessHash(callSession.adminAccessHash)
	phoneCall.SetGAOrB(callSession.g_b)
	phone_PoneCall.SetPhoneCall(phoneCall.To_PhoneCall())

	for _, u := range users {
		if u.GetId() == md.UserId {
			u.SetSelf(true)
		} else {
			u.SetSelf(false)
		}
		phone_PoneCall.Data2.Users = append(phone_PoneCall.Data2.Users, u.To_User())
	}

	glog.Infof("PhoneConfirmCall - reply: {%v}", phoneCall)
	return phone_PoneCall.To_Phone_PhoneCall(), nil
}
