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

package zproto

import "github.com/nebulaim/telegramd/baselib/net2"

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
	// PUSH 		   = 0x0006
)

const (
	kFrameHeaderLen = 32
	kMagicNumber    = 0x5A4D5450 // "ZMTP"
)

type newZProtoMessage func() net2.MessageBase

var zprotoFactories = map[uint16]newZProtoMessage{
	PROTO:              func() net2.MessageBase { return &RawPayload{} },
	PING:               func() net2.MessageBase { return &Ping{} },
	PONG:               func() net2.MessageBase { return &Pong{} },
	DROP:               func() net2.MessageBase { return &Drop{} },
	REDIRECT:           func() net2.MessageBase { return &Redirect{} },
	ACK:                func() net2.MessageBase { return &Ack{} },
	HANDSHAKE_REQ:      func() net2.MessageBase { return &HandshakeReq{} },
	HANDSHAKE_RSP:      func() net2.MessageBase { return &HandshakeRes{} },
	MARS_SIGNAL:        func() net2.MessageBase { return &MarsSignal{} },
	MESSAGE_ACK:        func() net2.MessageBase { return &MessageAck{} },
	RPC_REQUEST:        func() net2.MessageBase { return &RpcRequest{} },
	RPC_OK:             func() net2.MessageBase { return &RpcOk{} },
	RPC_ERROR:          func() net2.MessageBase { return &RpcError{} },
	RPC_FLOOD_WAIT:     func() net2.MessageBase { return &RpcFloodWait{} },
	RPC_INTERNAL_ERROR: func() net2.MessageBase { return &RpcInternalError{} },
	// PUSH: func() net2.MessageBase { return &{} },
}

func CheckPackageType(packageType uint16) (r bool) {
	_, r = zprotoFactories[packageType]
	return
}

func NewZProtoMessage(packageType uint16) net2.MessageBase {
	m, ok := zprotoFactories[packageType]
	if !ok {
		return nil
	}
	return m()
}

type PackageData struct {
	packageLength uint32 // 整个数据包长度 packageLen + bodyLen: metaDataLength + len(metaData) + len(payload) + len(crc32)
	magicNumber   uint32 // "ZMTP"
	packageIndex  uint32 // Index of package starting from zero. If packageIndex is broken connection need to be dropped.
	sessionId     uint64
	seqNum        uint64
	version       uint16 // version
	packageType   uint16 // frameType
	metadata      []byte
	body          []byte
	crc32         uint32 // CRC32 of body
}

type ZProtoMessage struct {
	sessionId uint64
	seqNum    uint64
	metadata  *Metadata
	message   net2.MessageBase
}

func (m *ZProtoMessage) Encode() ([]byte, error) {
	return nil, nil
}

func (m *ZProtoMessage) Decode(b []byte) error {
	return nil
}

/////////////////////////////////////////////////////////////////
type Metadata struct {
	serverId     int
	clientConnId uint64
	clientAddr   string
	traceId      int64
	spanId       int64
	receiveTime  int64
	from         string
	to           string
	options      map[string]string
}

func (m *Metadata) Encode() ([]byte, error) {
	return nil, nil
}

func (m *Metadata) Decode(b []byte) error {
	return nil
}

type RawPayload struct {
	payload []byte
}

func (m *RawPayload) Encode() ([]byte, error) {
	return nil, nil
}

func (m *RawPayload) Decode(b []byte) error {
	return nil
}

type Ping struct {
	pingId int64
}

func (m *Ping) Encode() ([]byte, error) {
	return nil, nil
}

func (m *Ping) Decode(b []byte) error {
	return nil
}

type Pong struct {
	messageId int64
	pingId int64
}

func (m *Pong) Encode() ([]byte, error) {
	return nil, nil
}

func (m *Pong) Decode(b []byte) error {
	return nil
}

type Drop struct {
	messageId    int64
	errorCode    int8
	errorMessage string
}

func (m *Drop) Encode() ([]byte, error) {
	return nil, nil
}

func (m *Drop) Decode(b []byte) error {
	return nil
}

type Redirect struct {
	host string
	port int
	timeout int
}

func (m *Redirect) Encode() ([]byte, error) {
	return nil, nil
}

func (m *Redirect) Decode(b []byte) error {
	return nil
}

type Ack struct {
	receivedPackageIndex int
}

func (m *Ack) Encode() ([]byte, error) {
	return nil, nil
}

func (m *Ack) Decode(b []byte) error {
	return nil
}

type HandshakeReq struct {
	protoRevision byte
	randomBytes []byte
}

func (m *HandshakeReq) Encode() ([]byte, error) {
	return nil, nil
}

func (m *HandshakeReq) Decode(b []byte) error {
	return nil
}

type HandshakeRes struct {
	protoRevision byte
	sha1 [32]byte
}

func (m *HandshakeRes) Encode() ([]byte, error) {
	return nil, nil
}

func (m *HandshakeRes) Decode(b []byte) error {
	return nil
}

type MarsSignal struct {
}

func (m *MarsSignal) Encode() ([]byte, error) {
	return nil, nil
}

func (m *MarsSignal) Decode(b []byte) error {
	return nil
}

type MessageAck struct {
	messageIds []uint64
}

func (m *MessageAck) Encode() ([]byte, error) {
	return nil, nil
}

func (m *MessageAck) Decode(b []byte) error {
	return nil
}

type RpcRequest struct {
	methodId string
	body []byte
}

func (m *RpcRequest) Encode() ([]byte, error) {
	return nil, nil
}

func (m *RpcRequest) Decode(b []byte) error {
	return nil
}

type RpcOk struct {
	requestMessageId int64
	methodResponseId string
	body []byte
}

func (m *RpcOk) Encode() ([]byte, error) {
	return nil, nil
}

func (m *RpcOk) Decode(b []byte) error {
	return nil
}

type RpcError struct {
	requestMessageId int64
	errorCode int
	errorTag string
	userMessage string
	canTryAgain bool
	errorData []byte
}

func (m *RpcError) Encode() ([]byte, error) {
	return nil, nil
}

func (m *RpcError) Decode(b []byte) error {
	return nil
}

type RpcFloodWait struct {
	requestMessageId int64
	delay int
}

func (m *RpcFloodWait) Encode() ([]byte, error) {
	return nil, nil
}

func (m *RpcFloodWait) Decode(b []byte) error {
	return nil
}

type RpcInternalError struct {
	requestMessageId int64
	canTryAgain      bool
	tryAgainDelay    int
}

func (m *RpcInternalError) Encode() ([]byte, error) {
	return nil, nil
}

func (m *RpcInternalError) Decode(b []byte) error {
	return nil
}

//type MTProtoRawMessage struct {
//	// needQuickAck bool
//	authKeyId  int64
//	sessionId  int64
//	salt       int64
//	seqNo      int32
//	messageId  int64
//	rawData    []byte
//}
