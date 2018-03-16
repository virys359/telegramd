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
	"github.com/nebulaim/telegramd/baselib/base"
	base2 "github.com/nebulaim/telegramd/biz_model/base"
	"math/rand"
)

/*
Request: { phone_requestCall
  user_id: { inputUser
	user_id: 264696845 [INT],
	access_hash: 13587416540126710881 [LONG],
  },
  random_id: 566731286 [INT],
  g_a_hash: 18 35 70 75 22 14 80 BB 70 42 86 78 FE 43 EB E5 43 40 6C D0 EC 39 2A B5 25 03 F6 0B 31 DF EC 1E [32 BYTES],
  protocol: { phoneCallProtocol
	flags: 3 [INT],
	udp_p2p: YES [ BY BIT 0 IN FIELD flags ],
	udp_reflector: YES [ BY BIT 1 IN FIELD flags ],
	min_layer: 65 [INT],
	max_layer: 65 [INT],
  },
},

Reply: { phone_phoneCall
  phone_call: { phoneCallWaiting
	flags: 0 [INT],
	id: 1926738269646495698 [LONG],
	access_hash: 12281641976525210193 [LONG],
	date: 1514943491 [INT],
	admin_id: 448603711 [INT],
	participant_id: 264696845 [INT],
	protocol: { phoneCallProtocol
	  flags: 3 [INT],
	  udp_p2p: YES [ BY BIT 0 IN FIELD flags ],
	  udp_reflector: YES [ BY BIT 1 IN FIELD flags ],
	  min_layer: 65 [INT],
	  max_layer: 65 [INT],
	},
	receive_date: [ SKIPPED BY BIT 0 IN FIELD flags ],
  },
}
 */

// phone.requestCall#5b95b3d4 user_id:InputUser random_id:int g_a_hash:bytes protocol:PhoneCallProtocol = phone.PhoneCall;
func (s *PhoneServiceImpl) PhoneRequestCall(ctx context.Context, request *mtproto.TLPhoneRequestCall) (*mtproto.Phone_PhoneCall, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("PhoneRequestCall - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	callSession := &phoneCallSession{
		id:                    base2.NextSnowflakeId(),
		adminId:               md.UserId,
		adminAccessHash:       rand.Int63(),
		participantId:         request.GetUserId().GetData2().GetUserId(),
		participantAccessHash: rand.Int63(),
		date:                  int32(time.Now().Unix()),
		state:                 0,
		protocol:              request.GetProtocol().To_PhoneCallProtocol(),
	}

	phoneCallSessionManager[callSession.id] = callSession

	//now := int32(time.Now().Unix())
	//participantId := request.GetUserId().GetData2().GetUserId()
	//phoneCallId := id.NextId()

	// push participantId
	/*
		updates: [ vector<0x0>
		  { updatePhoneCall
			phone_call: { phoneCallRequested
			  id: 1926738269689442709 [LONG],
			  access_hash: 7643618436880411068 [LONG],
			  date: 1514943513 [INT],
			  admin_id: 448603711 [INT],
			  participant_id: 264696845 [INT],
			  g_a_hash: 07 27 6B 44 B5 13 63 82 65 DC 8F C7 95 8C 3A 7C 5A F9 75 B6 E4 E2 90 8F C6 DA 90 DA B6 D9 7D 55 [32 BYTES],
			  protocol: { phoneCallProtocol
				flags: 3 [INT],
				udp_p2p: YES [ BY BIT 0 IN FIELD flags ],
				udp_reflector: YES [ BY BIT 1 IN FIELD flags ],
				min_layer: 65 [INT],
				max_layer: 65 [INT],
			  },
			},
		  },
		  { updateUserStatus
			user_id: 448603711 [INT],
			status: { userStatusOnline
			  expires: 1514943813 [INT],
			},
		  },
		],
	 */
	updates := mtproto.NewTLUpdates()

	updatePhoneCall := mtproto.NewTLUpdatePhoneCall()
	// push to participant_id
	phoneCallRequested := mtproto.NewTLPhoneCallRequested()
	phoneCallRequested.SetId(callSession.id)
	phoneCallRequested.SetAccessHash(callSession.participantAccessHash)
	phoneCallRequested.SetDate(callSession.date)
	phoneCallRequested.SetAdminId(callSession.adminId)
	phoneCallRequested.SetParticipantId(callSession.participantId)
	phoneCallRequested.SetGAHash(request.GetGAHash())
	phoneCallRequested.SetProtocol(callSession.protocol.To_PhoneCallProtocol())
	updatePhoneCall.SetPhoneCall(phoneCallRequested.To_PhoneCall())

	updates.Data2.Updates = append(updates.Data2.Updates, updatePhoneCall.To_Update())

	updateUserStatus := mtproto.NewTLUpdateUserStatus()
	updateUserStatus.SetUserId(md.UserId)
	status := mtproto.NewTLUserStatusOnline()
	status.SetExpires(callSession.date+300)
	updateUserStatus.SetStatus(status.To_UserStatus())
	updates.Data2.Updates = append(updates.Data2.Updates, updateUserStatus.To_Update())

	participantIds := []int32{callSession.adminId, callSession.participantId}
	users := model.GetUserModel().GetUserList(participantIds)
	for _, u := range users {
		if u.GetId() == md.UserId {
			u.SetSelf(false)
		} else {
			u.SetSelf(true)
		}
		// Add users
		updates.Data2.Users = append(updates.Data2.Users, u.To_User())
	}

	seq := int32(model.GetSequenceModel().NextSeqId(base.Int32ToString(callSession.participantId)))
	updates.SetSeq(seq)
	updates.SetDate(callSession.date)

	//delivery.GetDeliveryInstance().DeliveryUpdatesNotMe(
	//	md.AuthId,
	//	md.SessionId,
	//	md.NetlibSessionId,
	//	[]int32{callSession.participantId},
	//	updates.To_Updates().Encode())
	//// 返回给客户端

	// 2. reply
	phoneCall := mtproto.NewTLPhonePhoneCall()

	phoneCallWaiting := mtproto.NewTLPhoneCallWaiting()
	phoneCallWaiting.SetDate(callSession.date)
	phoneCallWaiting.SetId(callSession.id)
	phoneCallWaiting.SetAccessHash(callSession.adminAccessHash)
	//
	phoneCallWaiting.SetAdminId(callSession.adminId)
	phoneCallWaiting.SetParticipantId(callSession.participantId)
	phoneCallWaiting.SetProtocol(callSession.protocol.To_PhoneCallProtocol())

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

	glog.Infof("PhoneRequestCall - reply: {%v}", phoneCall)
	return phoneCall.To_Phone_PhoneCall(), nil
}
