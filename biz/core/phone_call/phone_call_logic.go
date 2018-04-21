/*
 *  Copyright (c) 2018, https://github.com/nebulaim
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

package phone_call

import (
	"github.com/nebulaim/telegramd/mtproto"
	"github.com/nebulaim/telegramd/biz/base"
	"math/rand"
	"time"
	"github.com/nebulaim/telegramd/biz/dal/dataobject"
	"github.com/nebulaim/telegramd/biz/dal/dao"
	base2 "github.com/nebulaim/telegramd/baselib/base"
	"encoding/hex"
	"fmt"
)

// TODO(@benqi): Using redis storage phone_call_sessions

type phoneCallLogic PhoneCallSession

func NewPhoneCallLogic(adminId, participantId int32, ga []byte, protocol *mtproto.TLPhoneCallProtocol) *phoneCallLogic {
	session := &phoneCallLogic{
		Id:                    base.NextSnowflakeId(),
		AdminId:               adminId,
		AdminAccessHash:       rand.Int63(),
		ParticipantId:         participantId,
		ParticipantAccessHash: rand.Int63(),
		UdpP2P:                protocol.GetUdpP2P(),
		UdpReflector:          protocol.GetUdpReflector(),
		MinLayer:              protocol.GetMinLayer(),
		MaxLayer:              protocol.GetMaxLayer(),
		GA:                    ga,
		State:                 0,
		Date:                  time.Now().Unix(),
	}

	do := &dataobject.PhoneCallSessionsDO{
		CallSessionId: session.Id,
		AdminId: session.AdminId,
		AdminAccessHash: session.AdminAccessHash,
		ParticipantId: session.ParticipantId,
		ParticipantAccessHash: session.ParticipantAccessHash,
		UdpP2p: base2.BoolToInt8(session.UdpP2P),
		UdpReflector: base2.BoolToInt8(session.UdpReflector),
		MinLayer: session.MinLayer,
		MaxLayer: session.MaxLayer,
		GA: hex.EncodeToString(session.GA),
		Date: int32(session.Date),
	}
	dao.GetPhoneCallSessionsDAO(dao.DB_MASTER).Insert(do)
	return session
}

func MakePhoneCallLogcByLoad(id int64) (*phoneCallLogic, error) {
	do := dao.GetPhoneCallSessionsDAO(dao.DB_SLAVE).Select(id)
	if do == nil {
		err := fmt.Errorf("not found call session: %d", id)
		return nil, err
	}

	session := &phoneCallLogic{
		Id:                    do.CallSessionId,
		AdminId:               do.AdminId,
		AdminAccessHash:       do.AdminAccessHash,
		ParticipantId:         do.ParticipantId,
		ParticipantAccessHash: do.ParticipantAccessHash,
		UdpP2P:                do.UdpP2p == 1,
		UdpReflector:          do.UdpReflector == 1,
		MinLayer:              do.MinLayer,
		MaxLayer:              do.MaxLayer,
		// GA:                    do.GA,
		State:                 0,
		Date:                  int64(do.Date),
	}

	session.GA, _ = hex.DecodeString(do.GA)
	return session, nil
}

func (p *phoneCallLogic) toPhoneCallProtocol() *mtproto.PhoneCallProtocol {
	return &mtproto.PhoneCallProtocol{
		Constructor: mtproto.TLConstructor_CRC32_phoneCallProtocol,
		Data2: &mtproto.PhoneCallProtocol_Data{
			UdpP2P:       p.UdpP2P,
			UdpReflector: p.UdpReflector,
			MinLayer:     p.MinLayer,
			MaxLayer:     p.MaxLayer,
		},
	}
}

// phoneCallRequested#83761ce4 id:long access_hash:long date:int admin_id:int participant_id:int g_a_hash:bytes protocol:PhoneCallProtocol = PhoneCall;
func (p *phoneCallLogic) ToPhoneCallRequested(selfId int32) *mtproto.TLPhoneCallRequested {
	var (
		accessHash int64
	)

	if selfId == p.AdminId {
		accessHash = p.AdminAccessHash
	} else {
		accessHash = p.ParticipantAccessHash
	}

	return &mtproto.TLPhoneCallRequested{Data2: &mtproto.PhoneCall_Data{
		Id:            p.Id,
		AccessHash:    accessHash,
		Date:          int32(p.Date),
		AdminId:       p.AdminId,
		ParticipantId: p.ParticipantId,
		GAHash:        p.GA,
		Protocol:      p.toPhoneCallProtocol(),
	}}
}

// phoneCallWaiting#1b8f4ad1 flags:# id:long access_hash:long date:int admin_id:int participant_id:int protocol:PhoneCallProtocol receive_date:flags.0?int = PhoneCall;
func (p *phoneCallLogic) ToPhoneCallWaiting(selfId int32, receiveDate int32) *mtproto.TLPhoneCallWaiting {
	var (
		accessHash int64
	)

	if selfId == p.AdminId {
		accessHash = p.AdminAccessHash
	} else {
		accessHash = p.ParticipantAccessHash
	}

	return &mtproto.TLPhoneCallWaiting{Data2: &mtproto.PhoneCall_Data{
		Id:            p.Id,
		AccessHash:    accessHash,
		Date:          int32(p.Date),
		AdminId:       p.AdminId,
		ParticipantId: p.ParticipantId,
		GAHash:        p.GA,
		Protocol:      p.toPhoneCallProtocol(),
		ReceiveDate:   receiveDate,
	}}
}
