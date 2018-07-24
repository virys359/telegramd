/*
 * WARNING! All changes made in this file will be lost!
 * Created from 'scheme.tl' by 'codegen_encode_decode.py'
 *
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

// ConstructorList
// RequestList

package mtproto

import (
// "encoding/binary"
// "fmt"
// "github.com/golang/protobuf/proto"
)

func NewTLMessagesSearchLayer68() *TLMessagesSearchLayer68 {
	return &TLMessagesSearchLayer68{}
}

func (m *TLMessagesSearchLayer68) Encode() []byte {
	x := NewEncodeBuf(512)
	x.Int(int32(TLConstructor_CRC32_messages_searchLayer68))

	// flags
	var flags uint32 = 0
	if m.FromId != nil {
		flags |= 1 << 0
	}
	x.UInt(flags)

	x.Bytes(m.Peer.Encode())
	x.String(m.Q)
	if m.FromId != nil {
		x.Bytes(m.FromId.Encode())
	}
	x.Bytes(m.Filter.Encode())
	x.Int(m.MinDate)
	x.Int(m.MaxDate)
	x.Int(m.Offset)
	x.Int(m.MaxId)
	x.Int(m.Limit)

	return x.buf
}

func (m *TLMessagesSearchLayer68) Decode(dbuf *DecodeBuf) error {
	flags := dbuf.UInt()
	_ = flags
	m2 := &InputPeer{}
	m2.Decode(dbuf)
	m.Peer = m2
	m.Q = dbuf.String()
	if (flags & (1 << 0)) != 0 {
		m4 := &InputUser{}
		m4.Decode(dbuf)
		m.FromId = m4
	}
	m5 := &MessagesFilter{}
	m5.Decode(dbuf)
	m.Filter = m5
	m.MinDate = dbuf.Int()
	m.MaxDate = dbuf.Int()
	m.Offset = dbuf.Int()
	m.MaxId = dbuf.Int()
	m.Limit = dbuf.Int()

	return dbuf.err
}
