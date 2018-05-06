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
	"github.com/golang/glog"
	"github.com/nebulaim/telegramd/baselib/net2"
)

const (
	PROTO              = 0xFF00
	PING               = 0x0100
	PONG               = 0x0200
	DROP               = 0x0300
	REDIRECT           = 0x0400
	ACK                = 0x0500
	HANDSHAKE_REQ      = 0x0600
	HANDSHAKE_RSP      = 0x0700
	MARS_SIGNAL        = 0x0800
	MESSAGE_ACK        = 0x0001
	RPC_REQUEST        = 0x000F
	RPC_OK             = 0x0010
	RPC_ERROR          = 0x0011
	RPC_FLOOD_WAIT     = 0x0012
	RPC_INTERNAL_ERROR = 0x0013
	PUSH 		   	   = 0x0014
)

const (
	kFrameHeaderLen = 16
	kMagicNumber    = 0x5A4D5450 // "ZMTP"
	kVersion 		= 1
)

type newZProtoMessage func() net2.MessageBase

var zprotoFactories = map[uint32]newZProtoMessage{
	PROTO:              func() net2.MessageBase { return &ZProtoRawPayload{} },
	PING:               func() net2.MessageBase { return &ZProtoPing{} },
	PONG:               func() net2.MessageBase { return &ZProtoPong{} },
	DROP:               func() net2.MessageBase { return &ZProtoDrop{} },
	REDIRECT:           func() net2.MessageBase { return &ZProtoRedirect{} },
	ACK:                func() net2.MessageBase { return &ZProtoAck{} },
	HANDSHAKE_REQ:      func() net2.MessageBase { return &ZProtoHandshakeReq{} },
	HANDSHAKE_RSP:      func() net2.MessageBase { return &ZProtoHandshakeRes{} },
	MARS_SIGNAL:        func() net2.MessageBase { return &ZProtoMarsSignal{} },
	MESSAGE_ACK:        func() net2.MessageBase { return &ZProtoMessageAck{} },
	RPC_REQUEST:        func() net2.MessageBase { return &ZProtoRpcRequest{} },
	RPC_OK:             func() net2.MessageBase { return &ZProtoRpcOk{} },
	RPC_ERROR:          func() net2.MessageBase { return &ZProtoRpcError{} },
	RPC_FLOOD_WAIT:     func() net2.MessageBase { return &ZProtoRpcFloodWait{} },
	RPC_INTERNAL_ERROR: func() net2.MessageBase { return &ZProtoRpcInternalError{} },
	// PUSH: func() net2.MessageBase { return &{} },
}

func CheckPackageType(packageType uint32) (r bool) {
	_, r = zprotoFactories[packageType]
	return
}

func NewZProtoMessage(packageType uint32) net2.MessageBase {
	m, ok := zprotoFactories[packageType]
	if !ok {
		glog.Errorf("Invalid packageType: %d", packageType)
		return nil
	}
	return m()
}

type ZProtoPackageData struct {
	packageLength uint32 // 整个数据包长度 packageLen + bodyLen: metaDataLength + len(metaData) + len(payload) + len(crc32)
	magicNumber   uint32 // "ZMTP"
	packageIndex  uint32 // Index of package starting from zero. If packageIndex is broken connection need to be dropped.
	version       uint16 // version
	reserved      uint16 // reserved

	sessionId   uint64
	seqNum      uint64
	metadata    []byte
	packageType uint32 // frameType
	body        []byte
	crc32       uint32 // CRC32 of body
}

type ZProtoMessage struct {
	SessionId  uint64
	SeqNum     uint64
	Metadata   *ZProtoMetadata
	Message    net2.MessageBase
}

func (m *ZProtoMessage) Encode() []byte {
	return nil
}

func (m *ZProtoMessage) Decode(b []byte) error {
	return nil
}

/*
type ZProtoMessageData struct {
	SessionId uint64
	SeqNum    uint64
	Metadata  *ZProtoMetadata
	Message   net2.MessageBase
}

func (m *ZProtoMessageData) Encode() ([]byte) {
	return nil
}

func (m *ZProtoMessageData) Decode(b []byte) error {
	return nil
}
*/

/////////////////////////////////////////////////////////////////
type ZProtoMetadata struct {
	ServerId     int
	ClientConnId uint64
	ClientAddr   string
	TraceId      int64
	SpanId       int64
	ReceiveTime  int64
	From         string
	// To           string
	Options map[string]string
	extend  []byte
}

func (m *ZProtoMetadata) Encode() []byte {
	x := NewEncodeBuf(512)

	x.Int(int32(m.ServerId))
	x.Long(int64(m.ClientConnId))
	x.String(m.ClientAddr)
	x.Long(m.TraceId)
	x.Long(m.SpanId)
	x.Long(m.ReceiveTime)
	x.String(m.From)

	// x.String(m.To)
	x.Int(int32(len(m.Options)))
	for k, v := range m.Options {
		x.String(k)
		x.String(v)
	}
	return x.GetBuf()
}

func (m *ZProtoMetadata) Decode(b []byte) error {
	dbuf := NewDecodeBuf(b)

	m.ServerId = int(dbuf.Int())
	m.ClientConnId = uint64(dbuf.Long())
	m.ClientAddr = dbuf.String()
	m.TraceId = dbuf.Long()
	m.SpanId = dbuf.Long()
	m.ReceiveTime = dbuf.Long()
	m.From = dbuf.String()

	// m.To = dbuf.String()
	len := int(dbuf.Int())
	for i := 0; i < len; i++ {
		k := dbuf.String()
		v := dbuf.String()
		m.Options[k] = v
	}
	return dbuf.GetError()
}

type ZProtoRawPayload struct {
	Payload []byte
}

func (m *ZProtoRawPayload) Encode() []byte {
	x := NewEncodeBuf(8 + len(m.Payload))
	x.UInt(PROTO)
	x.Bytes(m.Payload)
	return x.GetBuf()
}

func (m *ZProtoRawPayload) Decode(b []byte) error {
	m.Payload = b
	return nil
}

type ZProtoPing struct {
	PingId int64
}

func (m *ZProtoPing) Encode() []byte {
	x := NewEncodeBuf(12)
	x.UInt(PING)
	x.Long(m.PingId)
	return x.GetBuf()
}

func (m *ZProtoPing) Decode(b []byte) error {
	dbuf := NewDecodeBuf(b)
	m.PingId = dbuf.Long()
	return dbuf.GetError()
}

type ZProtoPong struct {
	MessageId int64
	PingId    int64
}

func (m *ZProtoPong) Encode() []byte {
	x := NewEncodeBuf(20)
	x.UInt(PONG)
	x.Long(m.MessageId)
	x.Long(m.PingId)
	return x.GetBuf()
}

func (m *ZProtoPong) Decode(b []byte) error {
	dbuf := NewDecodeBuf(b)
	m.MessageId = dbuf.Long()
	m.PingId = dbuf.Long()
	return dbuf.GetError()
}

type ZProtoDrop struct {
	MessageId    int64
	ErrorCode    int32
	ErrorMessage string
}

func (m *ZProtoDrop) Encode() []byte {
	x := NewEncodeBuf(64)
	x.UInt(DROP)
	x.Long(m.MessageId)
	x.Int(m.ErrorCode)
	x.String(m.ErrorMessage)
	return x.GetBuf()
}

func (m *ZProtoDrop) Decode(b []byte) error {
	dbuf := NewDecodeBuf(b)
	m.MessageId = dbuf.Long()
	m.ErrorCode = dbuf.Int()
	m.ErrorMessage = dbuf.String()
	return dbuf.GetError()
}

// RPC很少使用这条消息
type ZProtoRedirect struct {
	Host    string
	Port    int
	Timeout int
}

func (m *ZProtoRedirect) Encode() []byte {
	x := NewEncodeBuf(64)
	x.UInt(REDIRECT)
	x.String(m.Host)
	x.Int(int32(m.Port))
	x.Int(int32(m.Timeout))
	return x.GetBuf()
}

func (m *ZProtoRedirect) Decode(b []byte) error {
	dbuf := NewDecodeBuf(b)
	m.Host = dbuf.String()
	m.Port = int(dbuf.Int())
	m.Timeout = int(dbuf.Int())
	return nil
}

type ZProtoAck struct {
	ReceivedPackageIndex int
}

func (m *ZProtoAck) Encode() []byte {
	x := NewEncodeBuf(8)
	x.UInt(ACK)
	x.Int(int32(m.ReceivedPackageIndex))
	return x.GetBuf()
}

func (m *ZProtoAck) Decode(b []byte) error {
	dbuf := NewDecodeBuf(b)
	m.ReceivedPackageIndex = int(dbuf.Int())
	return dbuf.GetError()
}

type ZProtoHandshakeReq struct {
	ProtoRevision int
	RandomBytes   [32]byte
}

func (m *ZProtoHandshakeReq) Encode() []byte {
	x := NewEncodeBuf(64)
	x.UInt(HANDSHAKE_REQ)
	x.Int(int32(m.ProtoRevision))
	x.Bytes(m.RandomBytes[:])
	return x.GetBuf()
}

func (m *ZProtoHandshakeReq) Decode(b []byte) error {
	dbuf := NewDecodeBuf(b)
	m.ProtoRevision = int(dbuf.Int())
	randomBytes := dbuf.Bytes(32)
	copy(m.RandomBytes[:], randomBytes)
	return dbuf.GetError()
}

type ZProtoHandshakeRes struct {
	ProtoRevision int
	Sha1          [32]byte
}

func (m *ZProtoHandshakeRes) Encode() []byte {
	x := NewEncodeBuf(64)
	x.UInt(HANDSHAKE_RSP)
	x.Int(int32(m.ProtoRevision))
	x.Bytes(m.Sha1[:])
	return x.GetBuf()
}

func (m *ZProtoHandshakeRes) Decode(b []byte) error {
	dbuf := NewDecodeBuf(b)
	m.ProtoRevision = int(dbuf.Int())
	sha1 := dbuf.Bytes(32)
	copy(m.Sha1[:], sha1)
	return dbuf.GetError()
}

type ZProtoMarsSignal struct {
}

func (m *ZProtoMarsSignal) Encode() []byte {
	x := NewEncodeBuf(4)
	x.UInt(MARS_SIGNAL)
	return x.GetBuf()
}

func (m *ZProtoMarsSignal) Decode(b []byte) error {
	return nil
}

type ZProtoMessageAck struct {
	MessageIds []uint64
}

func (m *ZProtoMessageAck) Encode() []byte {
	x := NewEncodeBuf(512)
	x.UInt(MESSAGE_ACK)
	x.Int(int32(len(m.MessageIds)))
	for _, id := range m.MessageIds {
		x.Long(int64(id))
	}
	return x.GetBuf()
}

func (m *ZProtoMessageAck) Decode(b []byte) error {
	dbuf := NewDecodeBuf(b)
	len := int(dbuf.Int())
	for i := 0; i < len; i++ {
		m.MessageIds = append(m.MessageIds, uint64(dbuf.Long()))
	}
	return dbuf.GetError()
}

type ZProtoRpcRequest struct {
	MethodId string
	Body     []byte
}

func (m *ZProtoRpcRequest) Encode() []byte {
	x := NewEncodeBuf(512)
	x.UInt(RPC_REQUEST)
	x.String(m.MethodId)
	x.Bytes(m.Body)
	return x.GetBuf()
}

func (m *ZProtoRpcRequest) Decode(b []byte) error {
	dbuf := NewDecodeBuf(b)
	m.MethodId = dbuf.String()
	m.Body = dbuf.StringBytes()
	return dbuf.GetError()
}

type ZProtoRpcOk struct {
	RequestMessageId int64
	MethodResponseId string
	Body             []byte
}

func (m *ZProtoRpcOk) Encode() []byte {
	x := NewEncodeBuf(512)
	x.UInt(RPC_OK)
	x.Long(m.RequestMessageId)
	x.String(m.MethodResponseId)
	x.StringBytes(m.Body)
	return x.GetBuf()
}

func (m *ZProtoRpcOk) Decode(b []byte) error {
	dbuf := NewDecodeBuf(b)
	m.RequestMessageId = dbuf.Long()
	m.MethodResponseId = dbuf.String()
	m.Body = dbuf.StringBytes()
	return dbuf.GetError()
}

type ZProtoRpcError struct {
	RequestMessageId int64
	ErrorCode        int
	ErrorTag         string
	UserMessage      string
	CanTryAgain      bool
	ErrorData        []byte
}

func (m *ZProtoRpcError) Encode() []byte {
	x := NewEncodeBuf(512)
	x.UInt(RPC_ERROR)
	x.Long(m.RequestMessageId)
	x.Int(int32(m.ErrorCode))
	x.String(m.ErrorTag)
	x.String(m.UserMessage)
	if m.CanTryAgain {
		x.Int(1)
	} else {
		x.Int(0)
	}
	x.StringBytes(m.ErrorData)
	return x.GetBuf()
}

func (m *ZProtoRpcError) Decode(b []byte) error {
	dbuf := NewDecodeBuf(b)
	m.RequestMessageId = dbuf.Long()
	m.ErrorCode = int(dbuf.Int())
	m.ErrorTag = dbuf.String()
	m.UserMessage = dbuf.String()
	canTryAgain := dbuf.Int()
	m.CanTryAgain = canTryAgain == 1
	m.ErrorData = dbuf.StringBytes()
	return dbuf.GetError()
}

type ZProtoRpcFloodWait struct {
	RequestMessageId int64
	Delay            int
}

func (m *ZProtoRpcFloodWait) Encode() []byte {
	x := NewEncodeBuf(512)
	x.UInt(RPC_FLOOD_WAIT)
	x.Long(m.RequestMessageId)
	x.Int(int32(m.Delay))
	return x.GetBuf()
}

func (m *ZProtoRpcFloodWait) Decode(b []byte) error {
	dbuf := NewDecodeBuf(b)
	m.RequestMessageId = dbuf.Long()
	m.Delay = int(dbuf.Int())
	return dbuf.GetError()
}

type ZProtoRpcInternalError struct {
	RequestMessageId int64
	CanTryAgain      bool
	TryAgainDelay    int
}

func (m *ZProtoRpcInternalError) Encode() []byte {
	x := NewEncodeBuf(512)
	x.UInt(RPC_INTERNAL_ERROR)
	x.Long(m.RequestMessageId)
	if m.CanTryAgain {
		x.Int(1)
	} else {
		x.Int(0)
	}
	x.Int(int32(m.TryAgainDelay))
	return x.GetBuf()
}

func (m *ZProtoRpcInternalError) Decode(b []byte) error {
	dbuf := NewDecodeBuf(b)
	m.RequestMessageId = dbuf.Long()
	canTryAgain := dbuf.Int()
	m.CanTryAgain = canTryAgain == 1
	m.TryAgainDelay = int(dbuf.Int())
	return dbuf.GetError()
}
