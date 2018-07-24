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

func NewTLMessagesReadHistoryLayer2() *TLMessagesReadHistoryLayer2 {
	return &TLMessagesReadHistoryLayer2{}
}

func (m *TLMessagesReadHistoryLayer2) Encode() []byte {
	x := NewEncodeBuf(512)
	x.Int(int32(TLConstructor_CRC32_messages_readHistory))

	x.Bytes(m.Peer.Encode())
	x.Int(m.MaxId)
	x.Int(m.Offset)

	return x.buf
}

func (m *TLMessagesReadHistoryLayer2) Decode(dbuf *DecodeBuf) error {
	m1 := &InputPeer{}
	m1.Decode(dbuf)
	m.Peer = m1
	m.MaxId = dbuf.Int()
	m.Offset = dbuf.Int()

	return dbuf.err
}
