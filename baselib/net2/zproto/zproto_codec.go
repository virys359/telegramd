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
	"encoding/binary"
	"fmt"
	"github.com/golang/glog"
	"github.com/nebulaim/telegramd/baselib/net2"
	"io"
	"net"
	// "golang.org/x/text/message"
)

func init() {
	net2.RegisterPtotocol("zproto", &ZProto{})
}

type ZProto struct {
	recvLastPakageIndex uint32
	sendLastPakageIndex uint32
}

func (this *ZProto) NewCodec(rw io.ReadWriter) (net2.Codec, error) {
	codec := new(ZProtoCodec)
	codec.conn = rw.(*net.TCPConn)
	codec.headBuf = make([]byte, kFrameHeaderLen)
	return codec, nil
}

type ZProtoCodec struct {
	conn           *net.TCPConn
	headBuf        []byte
	lastFrameIndex uint16
}

func (c *ZProtoCodec) Send(msg interface{}) error {
	switch msg.(type) {
	case *ZProtoMessage:
	default:
		return fmt.Errorf("Invalid zprotomessage")
	}

	return nil
}

func (c *ZProtoCodec) Receive() (interface{}, error) {
	_, err := io.ReadFull(c.conn, c.headBuf)
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	// 1. check packageLength
	packageLength := binary.BigEndian.Uint32(c.headBuf[:4])

	// 1. check magic
	magic := binary.BigEndian.Uint32(c.headBuf[4:8])
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
	packageIndex := binary.BigEndian.Uint32(c.headBuf[8:12])
	_ = packageIndex

	// 3. check version
	// if c.headBuf[6] == 0
	//crc32       	uint32 // CRC32 of body

	packageType := binary.BigEndian.Uint16(c.headBuf[14:16])
	// 4. check frameType

	//switch packageType {
	//case PROTO:
	//case PING:
	//case PONG:
	//case DROP:
	//case REDIRECT:
	//case ACK:
	//case HANDSHAKE_REQ:
	//case HANDSHAKE_RSP:
	//case MARS_SIGNAL:
	//}

	sessionId := binary.BigEndian.Uint64(c.headBuf[16:24])
	seqNo := binary.BigEndian.Uint64(c.headBuf[24:32])

	// bodyLength := binary.BigEndian.Uint32(c.headBuf[8:12])
	// check bodyLength

	payload := make([]byte, packageLength) // recv crc32
	_, err = io.ReadFull(c.conn, payload)
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	mdLen := binary.BigEndian.Uint32(payload[:4])
	// check mdLen

	md := &Metadata{}
	md.Decode(payload[4 : mdLen+4])

	// messageLen := binary.BigEndian.Uint32(payload[mdLen:mdLen+4])
	// check messageLen

	//switch packageType {
	//case PROTO:
	//}
	//

	m2 := NewZProtoMessage(packageType)
	m2.Decode(payload[mdLen : packageLength-4-32])

	crc32 := binary.BigEndian.Uint32(payload[packageLength-4 : packageLength])
	_ = crc32
	// check crc32

	message := &ZProtoMessage{
		sessionId: sessionId,
		seqNum:    seqNo,
		message:   m2,
	}
	// message.sessionId = sessionId
	// message.message = m2

	return message, nil
}

func (c *ZProtoCodec) Close() error {
	return c.conn.Close()
}
