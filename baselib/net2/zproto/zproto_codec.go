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

import (
	"net"
	"github.com/nebulaim/telegramd/baselib/net2"
	"io"
	"github.com/golang/glog"
	"fmt"
	"encoding/binary"
)

type ZProto struct {
}

func (this *ZProto) NewCodec(conn *net.TCPConn) (net2.Codec, error) {
	codec := new(ZProtoCodec)
	codec.conn = conn
	codec.headBuf = make([]byte, kFrameHeaderLen)
	return codec, nil
}

type ZProtoCodec struct {
	conn *net.TCPConn
	headBuf []byte
	lastFrameIndex uint16
}

func (c *ZProtoCodec) Send(msg interface{}) error {
	//buf := []byte(msg.(string))
	//
	//if _, err := c.conn.Write(buf); err != nil {
	//	return err
	//}
	return nil
}

func (c *ZProtoCodec) Receive() (interface{}, error) {
	_, err := io.ReadFull(c.conn, c.headBuf)
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	// 1. check magic
	magic := binary.BigEndian.Uint32(c.headBuf[:4])
	if magic != kMagicNumber {
		// glog.Errorf("")
		err = fmt.Errorf("invalid magic number: %d", magic)
		return nil, err
	}

	// 2. check frameIndex
	//bool ZProtoFrameDecoder::CheckPackageIndex(uint16_t frame_index) {
	////
	//return frame_index-last_frame_index_ == 1 ||
	//(last_frame_index_ == std::numeric_limits<uint16_t>::max() &&
	//frame_index == 0);
	//}
	frameIndex := binary.BigEndian.Uint16(c.headBuf[4:6])
	_ = frameIndex

	// 3. check version
	// if c.headBuf[6] == 0

	// 4. check frameType
	switch c.headBuf[7] {
	case PROTO:
	case PING:
	case PONG:
	case DROP:
	case REDIRECT:
	case ACK:
	case HANDSHAKE_REQ:
	case HANDSHAKE_RSP:
	case MARS_SIGNAL:
	}

	bodyLength := binary.BigEndian.Uint32(c.headBuf[8:12])
	// check bodyLength

	payload := make([]byte, bodyLength+4) 	// recv crc32
	_, err = io.ReadFull(c.conn, payload)
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	crc32 := binary.BigEndian.Uint32(payload[bodyLength-4:bodyLength])
	_ = crc32
	// check crc32

	return nil, nil
}

func (c *ZProtoCodec) Close() error {
	return c.conn.Close()
}

const (
	PROTO         = 0x00
	PING          = 0x01
	PONG          = 0x02
	DROP          = 0x03
	REDIRECT      = 0x04
	ACK           = 0x05
	HANDSHAKE_REQ = 0x06
	HANDSHAKE_RSP = 0x07
	MARS_SIGNAL   = 0x08
)

const (
	kFrameHeaderLen = 12
	kMagicNumber = 0x5A4D5450	// "ZMTP"
)

//
type ZProtoFrame struct {
	magicNumber uint32 // "ZMTP"
	frameIndex  uint16 // Index of package starting from zero. If packageIndex is broken connection need to be dropped.
	version     uint8  // version
	frameType   uint8  // frameType
	bodyLength  uint32 // bodyLen: metaDataLength + len(metaData) + len(payload) + len(crc32)
	body        []byte
	crc32       uint32 // CRC32 of body
}

/////////////////////////////////////////////////////////////////
//type Metadata struct {
//}
//
//type ZProtoMessage interface {
//}
//
// frameType = 0x00
type ZProtoRawData struct {
	payload   []byte
	//md      Metadata
	//payload ZProtoMessage
	// randomBytes []byte
}

// frameType = 0x01
// Ping message can be sent from both sides
type Ping struct {
	randomBytes []byte
}

// frameType = 0x02
// Pong message need to be sent immediately after receving Ping message
type Pong struct {
	// Same bytes as in Ping package
	randomBytes []byte
}

// frameType = 0x03
// Notification about connection drop
type Drop struct {
	messageId    int64
	errorCode    int8
	errorMessage string
}

// frameType = 0x04
// Sent by server when we need to temporary redirect to another server
type Redirect struct {
	host string
	port int
	// Redirection timeout
	timeout int
}

// frameType = 0x05
// Proto package is received by destination peer. Used for determening of connection state
type Ack struct {
	receivedPackageIndex int
}

// frameType == 0x06
type HandshakeReq struct {
	// Current MTProto revision
	// For Rev 2 need to eq 1
	protoRevision byte
	// API Major and Minor version
	apiMajorVersion byte
	apiMinorVersion byte
	// Some Random Bytes (suggested size is 32 bytes)
	randomBytes []byte
}

// frameType == 0x07
type HandshakeRes struct {
	// return same versions as request, 0 - version is not supported
	protoRevision   byte
	apiMajorVersion byte
	apiMinorVersion byte
	// SHA256 of randomBytes from request
	sha1 [32]byte
}

// frameType == 0x08
type MarsSignal struct {
}

/////////////////////////////////////////////////////////////////
type Metadata struct {
	dcId int
	serverId int
	clientConnId   uint64
	clientAddr string
	traceId int64
	spanId int64
	recviveTime int64

	from string
	to string

	options map[string]string
}

type MTProtoRawMessage struct {
	// needQuickAck bool
	authKeyId  int64
	sessionId  int64
	salt       int64
	seqNo      int32
	messageId  int64

	rawData    []byte
}


