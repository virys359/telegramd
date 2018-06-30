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

package mtproto

import (
	"encoding/hex"
	"fmt"
)

// import "github.com/golang/glog"

const (
	STATE_ERROR = 0x0000

	STATE_CONNECTED2 = 0x0100
	STATE_HANDSHAKE  = 0x0200

	STATE_pq     = 0x0201
	STATE_pq_res = 0x0202
	STATE_pq_ack = 0x0203

	STATE_DH_params     = 0x0204
	STATE_DH_params_res = 0x0205
	STATE_DH_params_ack = 0x0206

	STATE_dh_gen     = 0x0207
	STATE_dh_gen_res = 0x0208
	STATE_dh_gen_ack = 0x0209

	STATE_AUTH_KEY = 0x0300
)

const (
	RES_STATE_NONE  = 0x00
	RES_STATE_OK    = 0x01
	RES_STATE_ERROR = 0x02
)

const (
	SESSION_HANDSHAKE    			= 0x01
	SESSION_SESSION_DATA 			= 0x02
	SYNC_DATA            			= 0x03
	SESSION_SESSION_CLIENT_NEW 		= 0x04
	SESSION_SESSION_CLIENT_CLOSED 	= 0x05
)

//func isHandshake(state int) bool {
//	return state >= STATE_CONNECTED2 && state <= STATE_dh_gen_ack
//}

type HandshakeState struct {
	State    int    // 状态
	ResState int    // 后端握手返回的结果
	Ctx      []byte // 握手上下文数据，透传给后端
}

func (s *HandshakeState) String() string {
	return fmt.Sprintf("{state: %d, res_state: %d, ctx: %s}", s.State, s.ResState, hex.EncodeToString(s.Ctx))
}

type ZProtoHandshakeMessage struct {
	State      *HandshakeState
	MTPMessage *MTPRawMessage
}

func (m *ZProtoHandshakeMessage) Encode() []byte {
	x := NewEncodeBuf(512)
	x.UInt(SESSION_HANDSHAKE)
	x.Int(int32(m.State.State))
	x.Int(int32(m.State.ResState))
	x.StringBytes(m.State.Ctx)
	x.Bytes(m.MTPMessage.Encode())
	return x.GetBuf()
}

func (m *ZProtoHandshakeMessage) Decode(b []byte) error {
	dbuf := NewDecodeBuf(b)
	m.State.State = int(dbuf.Int())
	m.State.ResState = int(dbuf.Int())
	m.State.Ctx = dbuf.StringBytes()

	authKeyId := dbuf.Long()
	err := dbuf.GetError()
	if err == nil {
		m.MTPMessage = &MTPRawMessage{
			AuthKeyId: authKeyId,
		}
		err = m.MTPMessage.Decode(b[dbuf.off-8:])
	}
	return err
}

type ZProtoSessionData struct {
	// MTPPayload    []byte
	MTPMessage *MTPRawMessage
}

func (m *ZProtoSessionData) Encode() []byte {
	x := NewEncodeBuf(512)
	x.UInt(SESSION_SESSION_DATA)
	x.Bytes(m.MTPMessage.Encode())
	return x.GetBuf()
}

func (m *ZProtoSessionData) Decode(b []byte) error {
	m.MTPMessage = &MTPRawMessage{}
	return m.MTPMessage.Decode(b)
}

///////////////////////////////////////////////////////////////////////////////////////////
type ZProtoSessionClientNew struct {
	proto int32
	// MTPPayload    []byte
	// MTPMessage *MTPRawMessage
}

func (m *ZProtoSessionClientNew) Encode() []byte {
	x := NewEncodeBuf(512)
	x.UInt(SESSION_SESSION_CLIENT_NEW)
	x.Int(m.proto)
	return x.GetBuf()
}

func (m *ZProtoSessionClientNew) Decode(b []byte) error {
	dbuf := NewDecodeBuf(b)
	m.proto = dbuf.Int()
	return dbuf.GetError()
}

///////////////////////////////////////////////////////////////////////////////////////////
type ZProtoSessionClientClosed struct {
	// MTPPayload    []byte
	// MTPMessage *MTPRawMessage
}

func (m *ZProtoSessionClientClosed) Encode() []byte {
	x := NewEncodeBuf(512)
	x.UInt(SESSION_SESSION_CLIENT_CLOSED)
	return x.GetBuf()
}

func (m *ZProtoSessionClientClosed) Decode(b []byte) error {
	return nil
}

///////////////////////////////////////////////////////////////////////////////////////////
type SessionHandshakeMessage struct {
	State      *HandshakeState
	MTPMessage *UnencryptedMessage
}

func (m *SessionHandshakeMessage) Encode() []byte {
	x := NewEncodeBuf(512)
	x.UInt(SESSION_HANDSHAKE)
	x.Int(int32(m.State.State))
	x.Int(int32(m.State.ResState))
	x.StringBytes(m.State.Ctx)
	x.Bytes(m.MTPMessage.Encode())
	return x.GetBuf()
}

func (m *SessionHandshakeMessage) Decode(b []byte) error {
	// glog.Info(b)
	dbuf := NewDecodeBuf(b)
	m.State.State = int(dbuf.Int())
	m.State.ResState = int(dbuf.Int())
	m.State.Ctx = dbuf.StringBytes()
	m.MTPMessage = &UnencryptedMessage{}
	err := dbuf.GetError()
	if err == nil {
		dbuf.Long()
		err = m.MTPMessage.Decode(b[dbuf.off:])
	}
	return err
}

type SessionDataMessage struct {
	MTPMessage *EncryptedMessage2
}

func (m *SessionDataMessage) Encode() []byte {
	x := NewEncodeBuf(512)
	x.UInt(SESSION_SESSION_DATA)
	// x.Bytes(m.MTPMessage.Encode())
	return x.GetBuf()
}

func (m *SessionDataMessage) Decode(b []byte) error {
	//m.MTPMessage = &EncryptedMessage2{}
	//return m.MTPMessage.Decode(b)
	return nil
}
