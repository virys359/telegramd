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

///////////////////////////////////////////////////////////////////////////////
// InputFileLocation <--
//  + TL_InputDocumentFileLocation
//

// inputDocumentFileLocation#4e45abe9 id:long access_hash:long = InputFileLocation;
func (m *TLInputDocumentFileLocationLayer11) To_InputFileLocation() *InputFileLocation {
	return &InputFileLocation{
		Constructor: TLConstructor_CRC32_inputDocumentFileLocationLayer11,
		Data2:       m.Data2,
	}
}

func (m *TLInputDocumentFileLocationLayer11) SetId(v int64) { m.Data2.Id = v }
func (m *TLInputDocumentFileLocationLayer11) GetId() int64  { return m.Data2.Id }

func (m *TLInputDocumentFileLocationLayer11) SetAccessHash(v int64) { m.Data2.AccessHash = v }
func (m *TLInputDocumentFileLocationLayer11) GetAccessHash() int64  { return m.Data2.AccessHash }

func NewTLInputDocumentFileLocationLayer11() *TLInputDocumentFileLocationLayer11 {
	return &TLInputDocumentFileLocationLayer11{Data2: &InputFileLocation_Data{}}
}

func (m *TLInputDocumentFileLocationLayer11) Encode() []byte {
	x := NewEncodeBuf(512)
	x.Int(int32(TLConstructor_CRC32_inputDocumentFileLocationLayer11))

	x.Long(m.GetId())
	x.Long(m.GetAccessHash())

	return x.buf
}

func (m *TLInputDocumentFileLocationLayer11) Decode(dbuf *DecodeBuf) error {
	m.SetId(dbuf.Long())
	m.SetAccessHash(dbuf.Long())

	return dbuf.err
}
