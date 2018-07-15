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
	"fmt"
	// "github.com/golang/protobuf/proto"
)

///////////////////////////////////////////////////////////////////////////////
// Contacts_Found <--
//  + TL_ContactsFound
//

// contacts.found#b3134d9d my_results:Vector<Peer> results:Vector<Peer> chats:Vector<Chat> users:Vector<User> = contacts.Found;
func (m *TLContactsFoundLayer74) To_Contacts_Found() *Contacts_Found {
	return &Contacts_Found{
		Constructor: TLConstructor_CRC32_contacts_foundLayer74,
		Data2:       m.Data2,
	}
}

func (m *TLContactsFoundLayer74) SetMyResults(v []*Peer) { m.Data2.MyResults = v }
func (m *TLContactsFoundLayer74) GetMyResults() []*Peer  { return m.Data2.MyResults }

func (m *TLContactsFoundLayer74) SetResults(v []*Peer) { m.Data2.Results = v }
func (m *TLContactsFoundLayer74) GetResults() []*Peer  { return m.Data2.Results }

func (m *TLContactsFoundLayer74) SetChats(v []*Chat) { m.Data2.Chats = v }
func (m *TLContactsFoundLayer74) GetChats() []*Chat  { return m.Data2.Chats }

func (m *TLContactsFoundLayer74) SetUsers(v []*User) { m.Data2.Users = v }
func (m *TLContactsFoundLayer74) GetUsers() []*User  { return m.Data2.Users }

func NewTLContactsFoundLayer74() *TLContactsFoundLayer74 {
	return &TLContactsFoundLayer74{Data2: &Contacts_Found_Data{}}
}

func (m *TLContactsFoundLayer74) Encode() []byte {
	x := NewEncodeBuf(512)
	x.Int(int32(TLConstructor_CRC32_contacts_foundLayer74))

	x.Int(int32(TLConstructor_CRC32_vector))
	x.Int(int32(len(m.GetMyResults())))
	for _, v := range m.GetMyResults() {
		x.buf = append(x.buf, (*v).Encode()...)
	}
	x.Int(int32(TLConstructor_CRC32_vector))
	x.Int(int32(len(m.GetResults())))
	for _, v := range m.GetResults() {
		x.buf = append(x.buf, (*v).Encode()...)
	}
	x.Int(int32(TLConstructor_CRC32_vector))
	x.Int(int32(len(m.GetChats())))
	for _, v := range m.GetChats() {
		x.buf = append(x.buf, (*v).Encode()...)
	}
	x.Int(int32(TLConstructor_CRC32_vector))
	x.Int(int32(len(m.GetUsers())))
	for _, v := range m.GetUsers() {
		x.buf = append(x.buf, (*v).Encode()...)
	}

	return x.buf
}

func (m *TLContactsFoundLayer74) Decode(dbuf *DecodeBuf) error {
	c1 := dbuf.Int()
	if c1 != int32(TLConstructor_CRC32_vector) {
		dbuf.err = fmt.Errorf("Invalid CRC32_vector, c%d: %d", 1, c1)
		return dbuf.err
	}
	l1 := dbuf.Int()
	v1 := make([]*Peer, l1)
	for i := int32(0); i < l1; i++ {
		v1[i] = &Peer{}
		v1[i].Decode(dbuf)
	}
	m.SetMyResults(v1)

	c2 := dbuf.Int()
	if c2 != int32(TLConstructor_CRC32_vector) {
		dbuf.err = fmt.Errorf("Invalid CRC32_vector, c%d: %d", 2, c2)
		return dbuf.err
	}
	l2 := dbuf.Int()
	v2 := make([]*Peer, l2)
	for i := int32(0); i < l2; i++ {
		v2[i] = &Peer{}
		v2[i].Decode(dbuf)
	}
	m.SetResults(v2)

	c3 := dbuf.Int()
	if c3 != int32(TLConstructor_CRC32_vector) {
		dbuf.err = fmt.Errorf("Invalid CRC32_vector, c%d: %d", 3, c3)
		return dbuf.err
	}
	l3 := dbuf.Int()
	v3 := make([]*Chat, l3)
	for i := int32(0); i < l3; i++ {
		v3[i] = &Chat{}
		v3[i].Decode(dbuf)
	}
	m.SetChats(v3)

	c4 := dbuf.Int()
	if c4 != int32(TLConstructor_CRC32_vector) {
		dbuf.err = fmt.Errorf("Invalid CRC32_vector, c%d: %d", 4, c4)
		return dbuf.err
	}
	l4 := dbuf.Int()
	v4 := make([]*User, l4)
	for i := int32(0); i < l4; i++ {
		v4[i] = &User{}
		v4[i].Decode(dbuf)
	}
	m.SetUsers(v4)

	return dbuf.err
}

///////////////////////////////////////////////////////////////////////////////
// ExportedMessageLink <--
//  + TL_ExportedMessageLink
//

// exportedMessageLink#5dab1af4 link:string html:string = ExportedMessageLink;
func (m *TLExportedMessageLinkLayer74) To_ExportedMessageLink() *ExportedMessageLink {
	return &ExportedMessageLink{
		Constructor: TLConstructor_CRC32_exportedMessageLinkLayer74,
		Data2:       m.Data2,
	}
}

func (m *TLExportedMessageLinkLayer74) SetLink(v string) { m.Data2.Link = v }
func (m *TLExportedMessageLinkLayer74) GetLink() string  { return m.Data2.Link }

func (m *TLExportedMessageLinkLayer74) SetHtml(v string) { m.Data2.Html = v }
func (m *TLExportedMessageLinkLayer74) GetHtml() string  { return m.Data2.Html }

func NewTLExportedMessageLinkLayer74() *TLExportedMessageLinkLayer74 {
	return &TLExportedMessageLinkLayer74{Data2: &ExportedMessageLink_Data{}}
}

func (m *TLExportedMessageLinkLayer74) Encode() []byte {
	x := NewEncodeBuf(512)
	x.Int(int32(TLConstructor_CRC32_exportedMessageLinkLayer74))

	x.String(m.GetLink())
	x.String(m.GetHtml())

	return x.buf
}

func (m *TLExportedMessageLinkLayer74) Decode(dbuf *DecodeBuf) error {
	m.SetLink(dbuf.String())
	m.SetHtml(dbuf.String())

	return dbuf.err
}

///////////////////////////////////////////////////////////////////////////////
// InputPaymentCredentials <--
//  + TL_InputPaymentCredentialsAndroidPay
//

// inputPaymentCredentialsAndroidPay#ca05d50e payment_token:DataJSON google_transaction_id:string = InputPaymentCredentials;
func (m *TLInputPaymentCredentialsAndroidPayLayer74) To_InputPaymentCredentials() *InputPaymentCredentials {
	return &InputPaymentCredentials{
		Constructor: TLConstructor_CRC32_inputPaymentCredentialsAndroidPayLayer74,
		Data2:       m.Data2,
	}
}

func (m *TLInputPaymentCredentialsAndroidPayLayer74) SetPaymentToken(v *DataJSON) {
	m.Data2.PaymentToken = v
}
func (m *TLInputPaymentCredentialsAndroidPayLayer74) GetPaymentToken() *DataJSON {
	return m.Data2.PaymentToken
}

func (m *TLInputPaymentCredentialsAndroidPayLayer74) SetGoogleTransactionId(v string) {
	m.Data2.GoogleTransactionId = v
}
func (m *TLInputPaymentCredentialsAndroidPayLayer74) GetGoogleTransactionId() string {
	return m.Data2.GoogleTransactionId
}

func NewTLInputPaymentCredentialsAndroidPayLayer74() *TLInputPaymentCredentialsAndroidPayLayer74 {
	return &TLInputPaymentCredentialsAndroidPayLayer74{Data2: &InputPaymentCredentials_Data{}}
}

func (m *TLInputPaymentCredentialsAndroidPayLayer74) Encode() []byte {
	x := NewEncodeBuf(512)
	x.Int(int32(TLConstructor_CRC32_inputPaymentCredentialsAndroidPayLayer74))

	x.Bytes(m.GetPaymentToken().Encode())
	x.String(m.GetGoogleTransactionId())

	return x.buf
}

func (m *TLInputPaymentCredentialsAndroidPayLayer74) Decode(dbuf *DecodeBuf) error {
	m1 := &DataJSON{}
	m1.Decode(dbuf)
	m.SetPaymentToken(m1)
	m.SetGoogleTransactionId(dbuf.String())

	return dbuf.err
}

////////////////////////////////////////////////////////////////////////////
func NewTLAccountRegisterDeviceLayer74() *TLAccountRegisterDeviceLayer74 {
	return &TLAccountRegisterDeviceLayer74{}
}

func (m *TLAccountRegisterDeviceLayer74) Encode() []byte {
	x := NewEncodeBuf(512)
	x.Int(int32(TLConstructor_CRC32_account_registerDeviceLayer74))

	x.Int(m.TokenType)
	x.String(m.Token)
	x.Bytes(m.AppSandbox.Encode())
	x.VectorInt(m.OtherUids)

	return x.buf
}

func (m *TLAccountRegisterDeviceLayer74) Decode(dbuf *DecodeBuf) error {
	m.TokenType = dbuf.Int()
	m.Token = dbuf.String()
	m3 := &Bool{}
	m3.Decode(dbuf)
	m.AppSandbox = m3
	m.OtherUids = dbuf.VectorInt()

	return dbuf.err
}

func NewTLAccountUnregisterDeviceLayer74() *TLAccountUnregisterDeviceLayer74 {
	return &TLAccountUnregisterDeviceLayer74{}
}

func (m *TLAccountUnregisterDeviceLayer74) Encode() []byte {
	x := NewEncodeBuf(512)
	x.Int(int32(TLConstructor_CRC32_account_unregisterDeviceLayer74))

	x.Int(m.TokenType)
	x.String(m.Token)
	x.VectorInt(m.OtherUids)

	return x.buf
}

func (m *TLAccountUnregisterDeviceLayer74) Decode(dbuf *DecodeBuf) error {
	m.TokenType = dbuf.Int()
	m.Token = dbuf.String()
	m.OtherUids = dbuf.VectorInt()

	return dbuf.err
}

func NewTLChannelsExportMessageLinkLayer74() *TLChannelsExportMessageLinkLayer74 {
	return &TLChannelsExportMessageLinkLayer74{}
}

func (m *TLChannelsExportMessageLinkLayer74) Encode() []byte {
	x := NewEncodeBuf(512)
	x.Int(int32(TLConstructor_CRC32_channels_exportMessageLinkLayer74))

	x.Bytes(m.Channel.Encode())
	x.Int(m.Id)
	x.Bytes(m.Grouped.Encode())

	return x.buf
}

func (m *TLChannelsExportMessageLinkLayer74) Decode(dbuf *DecodeBuf) error {
	m1 := &InputChannel{}
	m1.Decode(dbuf)
	m.Channel = m1
	m.Id = dbuf.Int()
	m3 := &Bool{}
	m3.Decode(dbuf)
	m.Grouped = m3

	return dbuf.err
}
