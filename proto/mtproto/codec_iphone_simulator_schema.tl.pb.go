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
	"github.com/golang/glog"
)

///////////////////////////////////////////////////////////////////////////////
// Scheme <--
//  + TL_SchemeNotModified
//  + TL_Scheme
//

func (m *Scheme) Encode() []byte {
	switch m.GetConstructor() {
	case TLConstructor_CRC32_schemeNotModified:
		t := m.To_SchemeNotModified()
		return t.Encode()
	case TLConstructor_CRC32_scheme:
		t := m.To_Scheme()
		return t.Encode()

	default:
		glog.Error("Constructor error: ", m.GetConstructor())
		return nil
	}
}

func (m *Scheme) Decode(dbuf *DecodeBuf) error {
	m.Constructor = TLConstructor(dbuf.Int())
	switch m.Constructor {
	case TLConstructor_CRC32_schemeNotModified:
		m2 := &TLSchemeNotModified{&Scheme_Data{}}
		m2.Decode(dbuf)
		m.Data2 = m2.Data2
	case TLConstructor_CRC32_scheme:
		m2 := &TLScheme{&Scheme_Data{}}
		m2.Decode(dbuf)
		m.Data2 = m2.Data2

	default:
		return fmt.Errorf("Invalid constructorId: %d", int32(m.Constructor))
	}
	return dbuf.err
}

// schemeNotModified#263c9c58 = Scheme;
func (m *Scheme) To_SchemeNotModified() *TLSchemeNotModified {
	return &TLSchemeNotModified{
		Data2: m.Data2,
	}
}

// scheme#4e6ef65e scheme_raw:string types:Vector<SchemeType> methods:Vector<SchemeMethod> version:int = Scheme;
func (m *Scheme) To_Scheme() *TLScheme {
	return &TLScheme{
		Data2: m.Data2,
	}
}

// schemeNotModified#263c9c58 = Scheme;
func (m *TLSchemeNotModified) To_Scheme() *Scheme {
	return &Scheme{
		Constructor: TLConstructor_CRC32_schemeNotModified,
		Data2:       m.Data2,
	}
}

func NewTLSchemeNotModified() *TLSchemeNotModified {
	return &TLSchemeNotModified{Data2: &Scheme_Data{}}
}

func (m *TLSchemeNotModified) Encode() []byte {
	x := NewEncodeBuf(512)
	x.Int(int32(TLConstructor_CRC32_schemeNotModified))

	return x.buf
}

func (m *TLSchemeNotModified) Decode(dbuf *DecodeBuf) error {

	return dbuf.err
}

// scheme#4e6ef65e scheme_raw:string types:Vector<SchemeType> methods:Vector<SchemeMethod> version:int = Scheme;
func (m *TLScheme) To_Scheme() *Scheme {
	return &Scheme{
		Constructor: TLConstructor_CRC32_scheme,
		Data2:       m.Data2,
	}
}

func (m *TLScheme) SetSchemeRaw(v string) { m.Data2.SchemeRaw = v }
func (m *TLScheme) GetSchemeRaw() string  { return m.Data2.SchemeRaw }

func (m *TLScheme) SetTypes(v []*SchemeType) { m.Data2.Types = v }
func (m *TLScheme) GetTypes() []*SchemeType  { return m.Data2.Types }

func (m *TLScheme) SetMethods(v []*SchemeMethod) { m.Data2.Methods = v }
func (m *TLScheme) GetMethods() []*SchemeMethod  { return m.Data2.Methods }

func (m *TLScheme) SetVersion(v int32) { m.Data2.Version = v }
func (m *TLScheme) GetVersion() int32  { return m.Data2.Version }

func NewTLScheme() *TLScheme {
	return &TLScheme{Data2: &Scheme_Data{}}
}

func (m *TLScheme) Encode() []byte {
	x := NewEncodeBuf(512)
	x.Int(int32(TLConstructor_CRC32_scheme))

	x.String(m.GetSchemeRaw())
	x.Int(int32(TLConstructor_CRC32_vector))
	x.Int(int32(len(m.GetTypes())))
	for _, v := range m.GetTypes() {
		x.buf = append(x.buf, (*v).Encode()...)
	}
	x.Int(int32(TLConstructor_CRC32_vector))
	x.Int(int32(len(m.GetMethods())))
	for _, v := range m.GetMethods() {
		x.buf = append(x.buf, (*v).Encode()...)
	}
	x.Int(m.GetVersion())

	return x.buf
}

func (m *TLScheme) Decode(dbuf *DecodeBuf) error {
	m.SetSchemeRaw(dbuf.String())
	c2 := dbuf.Int()
	if c2 != int32(TLConstructor_CRC32_vector) {
		dbuf.err = fmt.Errorf("Invalid CRC32_vector, c%d: %d", 2, c2)
		return dbuf.err
	}
	l2 := dbuf.Int()
	v2 := make([]*SchemeType, l2)
	for i := int32(0); i < l2; i++ {
		v2[i] = &SchemeType{}
		v2[i].Decode(dbuf)
	}
	m.SetTypes(v2)

	c3 := dbuf.Int()
	if c3 != int32(TLConstructor_CRC32_vector) {
		dbuf.err = fmt.Errorf("Invalid CRC32_vector, c%d: %d", 3, c3)
		return dbuf.err
	}
	l3 := dbuf.Int()
	v3 := make([]*SchemeMethod, l3)
	for i := int32(0); i < l3; i++ {
		v3[i] = &SchemeMethod{}
		v3[i].Decode(dbuf)
	}
	m.SetMethods(v3)

	m.SetVersion(dbuf.Int())

	return dbuf.err
}

///////////////////////////////////////////////////////////////////////////////
// SchemeParam <--
//  + TL_SchemeParam
//

func (m *SchemeParam) Encode() []byte {
	switch m.GetConstructor() {
	case TLConstructor_CRC32_schemeParam:
		t := m.To_SchemeParam()
		return t.Encode()

	default:
		glog.Error("Constructor error: ", m.GetConstructor())
		return nil
	}
}

func (m *SchemeParam) Decode(dbuf *DecodeBuf) error {
	m.Constructor = TLConstructor(dbuf.Int())
	switch m.Constructor {
	case TLConstructor_CRC32_schemeParam:
		m2 := &TLSchemeParam{&SchemeParam_Data{}}
		m2.Decode(dbuf)
		m.Data2 = m2.Data2

	default:
		return fmt.Errorf("Invalid constructorId: %d", int32(m.Constructor))
	}
	return dbuf.err
}

// schemeParam#21b59bef name:string type:string = SchemeParam;
func (m *SchemeParam) To_SchemeParam() *TLSchemeParam {
	return &TLSchemeParam{
		Data2: m.Data2,
	}
}

// schemeParam#21b59bef name:string type:string = SchemeParam;
func (m *TLSchemeParam) To_SchemeParam() *SchemeParam {
	return &SchemeParam{
		Constructor: TLConstructor_CRC32_schemeParam,
		Data2:       m.Data2,
	}
}

func (m *TLSchemeParam) SetName(v string) { m.Data2.Name = v }
func (m *TLSchemeParam) GetName() string  { return m.Data2.Name }

func (m *TLSchemeParam) SetType(v string) { m.Data2.Type = v }
func (m *TLSchemeParam) GetType() string  { return m.Data2.Type }

func NewTLSchemeParam() *TLSchemeParam {
	return &TLSchemeParam{Data2: &SchemeParam_Data{}}
}

func (m *TLSchemeParam) Encode() []byte {
	x := NewEncodeBuf(512)
	x.Int(int32(TLConstructor_CRC32_schemeParam))

	x.String(m.GetName())
	x.String(m.GetType())

	return x.buf
}

func (m *TLSchemeParam) Decode(dbuf *DecodeBuf) error {
	m.SetName(dbuf.String())
	m.SetType(dbuf.String())

	return dbuf.err
}

///////////////////////////////////////////////////////////////////////////////
// SchemeMethod <--
//  + TL_SchemeMethod
//

func (m *SchemeMethod) Encode() []byte {
	switch m.GetConstructor() {
	case TLConstructor_CRC32_schemeMethod:
		t := m.To_SchemeMethod()
		return t.Encode()

	default:
		glog.Error("Constructor error: ", m.GetConstructor())
		return nil
	}
}

func (m *SchemeMethod) Decode(dbuf *DecodeBuf) error {
	m.Constructor = TLConstructor(dbuf.Int())
	switch m.Constructor {
	case TLConstructor_CRC32_schemeMethod:
		m2 := &TLSchemeMethod{&SchemeMethod_Data{}}
		m2.Decode(dbuf)
		m.Data2 = m2.Data2

	default:
		return fmt.Errorf("Invalid constructorId: %d", int32(m.Constructor))
	}
	return dbuf.err
}

// schemeMethod#479357c0 id:int method:string params:Vector<SchemeParam> type:string = SchemeMethod;
func (m *SchemeMethod) To_SchemeMethod() *TLSchemeMethod {
	return &TLSchemeMethod{
		Data2: m.Data2,
	}
}

// schemeMethod#479357c0 id:int method:string params:Vector<SchemeParam> type:string = SchemeMethod;
func (m *TLSchemeMethod) To_SchemeMethod() *SchemeMethod {
	return &SchemeMethod{
		Constructor: TLConstructor_CRC32_schemeMethod,
		Data2:       m.Data2,
	}
}

func (m *TLSchemeMethod) SetId(v int32) { m.Data2.Id = v }
func (m *TLSchemeMethod) GetId() int32  { return m.Data2.Id }

func (m *TLSchemeMethod) SetMethod(v string) { m.Data2.Method = v }
func (m *TLSchemeMethod) GetMethod() string  { return m.Data2.Method }

func (m *TLSchemeMethod) SetParams(v []*SchemeParam) { m.Data2.Params = v }
func (m *TLSchemeMethod) GetParams() []*SchemeParam  { return m.Data2.Params }

func (m *TLSchemeMethod) SetType(v string) { m.Data2.Type = v }
func (m *TLSchemeMethod) GetType() string  { return m.Data2.Type }

func NewTLSchemeMethod() *TLSchemeMethod {
	return &TLSchemeMethod{Data2: &SchemeMethod_Data{}}
}

func (m *TLSchemeMethod) Encode() []byte {
	x := NewEncodeBuf(512)
	x.Int(int32(TLConstructor_CRC32_schemeMethod))

	x.Int(m.GetId())
	x.String(m.GetMethod())
	x.Int(int32(TLConstructor_CRC32_vector))
	x.Int(int32(len(m.GetParams())))
	for _, v := range m.GetParams() {
		x.buf = append(x.buf, (*v).Encode()...)
	}
	x.String(m.GetType())

	return x.buf
}

func (m *TLSchemeMethod) Decode(dbuf *DecodeBuf) error {
	m.SetId(dbuf.Int())
	m.SetMethod(dbuf.String())
	c3 := dbuf.Int()
	if c3 != int32(TLConstructor_CRC32_vector) {
		dbuf.err = fmt.Errorf("Invalid CRC32_vector, c%d: %d", 3, c3)
		return dbuf.err
	}
	l3 := dbuf.Int()
	v3 := make([]*SchemeParam, l3)
	for i := int32(0); i < l3; i++ {
		v3[i] = &SchemeParam{}
		v3[i].Decode(dbuf)
	}
	m.SetParams(v3)

	m.SetType(dbuf.String())

	return dbuf.err
}

///////////////////////////////////////////////////////////////////////////////
// SchemeType <--
//  + TL_SchemeType
//

func (m *SchemeType) Encode() []byte {
	switch m.GetConstructor() {
	case TLConstructor_CRC32_schemeType:
		t := m.To_SchemeType()
		return t.Encode()

	default:
		glog.Error("Constructor error: ", m.GetConstructor())
		return nil
	}
}

func (m *SchemeType) Decode(dbuf *DecodeBuf) error {
	m.Constructor = TLConstructor(dbuf.Int())
	switch m.Constructor {
	case TLConstructor_CRC32_schemeType:
		m2 := &TLSchemeType{&SchemeType_Data{}}
		m2.Decode(dbuf)
		m.Data2 = m2.Data2

	default:
		return fmt.Errorf("Invalid constructorId: %d", int32(m.Constructor))
	}
	return dbuf.err
}

// schemeType#a8e1e989 id:int predicate:string params:Vector<SchemeParam> type:string = SchemeType;
func (m *SchemeType) To_SchemeType() *TLSchemeType {
	return &TLSchemeType{
		Data2: m.Data2,
	}
}

// schemeType#a8e1e989 id:int predicate:string params:Vector<SchemeParam> type:string = SchemeType;
func (m *TLSchemeType) To_SchemeType() *SchemeType {
	return &SchemeType{
		Constructor: TLConstructor_CRC32_schemeType,
		Data2:       m.Data2,
	}
}

func (m *TLSchemeType) SetId(v int32) { m.Data2.Id = v }
func (m *TLSchemeType) GetId() int32  { return m.Data2.Id }

func (m *TLSchemeType) SetPredicate(v string) { m.Data2.Predicate = v }
func (m *TLSchemeType) GetPredicate() string  { return m.Data2.Predicate }

func (m *TLSchemeType) SetParams(v []*SchemeParam) { m.Data2.Params = v }
func (m *TLSchemeType) GetParams() []*SchemeParam  { return m.Data2.Params }

func (m *TLSchemeType) SetType(v string) { m.Data2.Type = v }
func (m *TLSchemeType) GetType() string  { return m.Data2.Type }

func NewTLSchemeType() *TLSchemeType {
	return &TLSchemeType{Data2: &SchemeType_Data{}}
}

func (m *TLSchemeType) Encode() []byte {
	x := NewEncodeBuf(512)
	x.Int(int32(TLConstructor_CRC32_schemeType))

	x.Int(m.GetId())
	x.String(m.GetPredicate())
	x.Int(int32(TLConstructor_CRC32_vector))
	x.Int(int32(len(m.GetParams())))
	for _, v := range m.GetParams() {
		x.buf = append(x.buf, (*v).Encode()...)
	}
	x.String(m.GetType())

	return x.buf
}

func (m *TLSchemeType) Decode(dbuf *DecodeBuf) error {
	m.SetId(dbuf.Int())
	m.SetPredicate(dbuf.String())
	c3 := dbuf.Int()
	if c3 != int32(TLConstructor_CRC32_vector) {
		dbuf.err = fmt.Errorf("Invalid CRC32_vector, c%d: %d", 3, c3)
		return dbuf.err
	}
	l3 := dbuf.Int()
	v3 := make([]*SchemeParam, l3)
	for i := int32(0); i < l3; i++ {
		v3[i] = &SchemeParam{}
		v3[i].Decode(dbuf)
	}
	m.SetParams(v3)

	m.SetType(dbuf.String())

	return dbuf.err
}

func NewTLHelpGetScheme() *TLHelpGetScheme {
	return &TLHelpGetScheme{}
}

func (m *TLHelpGetScheme) Encode() []byte {
	x := NewEncodeBuf(512)
	x.Int(int32(TLConstructor_CRC32_help_getScheme))

	x.Int(m.Version)

	return x.buf
}

func (m *TLHelpGetScheme) Decode(dbuf *DecodeBuf) error {
	m.Version = dbuf.Int()

	return dbuf.err
}
