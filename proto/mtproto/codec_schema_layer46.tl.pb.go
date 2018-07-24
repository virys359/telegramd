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

func NewTLHelpGetInviteTextLayer46() *TLHelpGetInviteTextLayer46 {
	return &TLHelpGetInviteTextLayer46{}
}

func (m *TLHelpGetInviteTextLayer46) Encode() []byte {
	x := NewEncodeBuf(512)
	x.Int(int32(TLConstructor_CRC32_help_getInviteTextLayer46))

	x.String(m.LangCode)

	return x.buf
}

func (m *TLHelpGetInviteTextLayer46) Decode(dbuf *DecodeBuf) error {
	m.LangCode = dbuf.String()

	return dbuf.err
}

func NewTLHelpGetAppUpdateLayer46() *TLHelpGetAppUpdateLayer46 {
	return &TLHelpGetAppUpdateLayer46{}
}

func (m *TLHelpGetAppUpdateLayer46) Encode() []byte {
	x := NewEncodeBuf(512)
	x.Int(int32(TLConstructor_CRC32_help_getAppUpdateLayer46))

	x.String(m.DeviceModel)
	x.String(m.SystemVersion)
	x.String(m.AppVersion)
	x.String(m.LangCode)

	return x.buf
}

func (m *TLHelpGetAppUpdateLayer46) Decode(dbuf *DecodeBuf) error {
	m.DeviceModel = dbuf.String()
	m.SystemVersion = dbuf.String()
	m.AppVersion = dbuf.String()
	m.LangCode = dbuf.String()

	return dbuf.err
}
