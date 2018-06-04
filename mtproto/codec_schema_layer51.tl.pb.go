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

func NewTLAuthSendCodeLayer51() * TLAuthSendCodeLayer51 {
	return &TLAuthSendCodeLayer51{}
}

func (m* TLAuthSendCodeLayer51) Encode() []byte {
	x := NewEncodeBuf(512)
	x.Int(int32(TLConstructor_CRC32_auth_sendCodeLayer51))

	// flags
	var flags uint32 = 0
	if m.AllowFlashcall == true { flags |= 1 << 0 }
	if m.CurrentNumber != nil { flags |= 1 << 0 }
	x.UInt(flags)


	x.String(m.PhoneNumber)
	if m.CurrentNumber != nil {
		x.Bytes(m.CurrentNumber.Encode())
	}
	x.Int(m.ApiId)
	x.String(m.ApiHash)
	x.String(m.LangCode)

	return x.buf
}

func (m* TLAuthSendCodeLayer51) Decode(dbuf *DecodeBuf) error {
	flags := dbuf.UInt()
	_ = flags
	if (flags & (1 << 0)) != 0 { m.AllowFlashcall = true }
	m.PhoneNumber = dbuf.String()
	if (flags & (1 << 0)) != 0 {
		m4 := &Bool{}
		m4.Decode(dbuf)
		m.CurrentNumber = m4
	}
	m.ApiId = dbuf.Int()
	m.ApiHash = dbuf.String()
	m.LangCode = dbuf.String()

	return dbuf.err
}
