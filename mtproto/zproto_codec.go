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
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/golang/glog"
	"github.com/nebulaim/telegramd/baselib/net2"
	"io"
	"net"
)

func init() {
	net2.RegisterProtocol("zproto", &ZProto{})
}

type ZProto struct {
}

func (this *ZProto) NewCodec(rw io.ReadWriter) (net2.Codec, error) {
	codec := new(ZProtoCodec)
	codec.conn = rw.(*net.TCPConn)
	codec.headBuf = make([]byte, kFrameHeaderLen)
	codec.recvLastPakageIndex = 0
	codec.sendLastPakageIndex = 1
	return codec, nil
}

type ZProtoCodec struct {
	conn                *net.TCPConn
	headBuf             []byte
	recvLastPakageIndex uint32
	sendLastPakageIndex uint32
}

func (c *ZProtoCodec) Send(msg interface{}) error {
	switch msg.(type) {
	case *ZProtoMessage:
		message, _ := msg.(*ZProtoMessage)
		x := NewEncodeBuf(512)
		// 16
		x.UInt(0)            // packageLength
		x.UInt(kMagicNumber) // magicNumber
		c.sendLastPakageIndex++
		x.UInt(c.sendLastPakageIndex) // packageIndex

		x.UInt16(0)        // reserved
		x.UInt16(kVersion) // version

		//
		x.Long(int64(message.SessionId)) // sessionId
		x.Long(int64(message.SeqNum))    // seqNum
		// metadata
		if message.Metadata == nil {
			x.UInt(0)
		} else {
			b2 := message.Metadata.Encode()
			x.UInt(uint32(len(b2)))
			x.Bytes(b2)
		}

		x.Bytes(message.Message.Encode()) // packageType + body
		x.UInt(0)                         // crc32

		xbuf := x.GetBuf()
		binary.LittleEndian.PutUint32(xbuf, uint32(len(xbuf)))
		_, err := c.conn.Write(xbuf)
		return err
	default:
		return fmt.Errorf("Invalid zproto message")
	}
}

func (c *ZProtoCodec) Receive() (interface{}, error) {
	for {
		_, err := io.ReadFull(c.conn, c.headBuf)
		if err != nil {
			glog.Error(err)
			return nil, err
		}

		// 1. check packageLength
		// TODO(@benqi): check packageLength
		packageLength := binary.LittleEndian.Uint32(c.headBuf[:4])
		// glog.Infof("packageLength: %d", packageLength)

		// 1. check magic
		magic := binary.LittleEndian.Uint32(c.headBuf[4:8])
		if magic != kMagicNumber {
			err = fmt.Errorf("invalid magic number: %d", magic)
			glog.Error(err)
			return nil, err
		}

		// 2. check packageIndex
		// TODO(@benqi): check packageIndex
		packageIndex := binary.LittleEndian.Uint32(c.headBuf[8:12])
		c.recvLastPakageIndex = packageIndex

		// 3. check version
		version := binary.LittleEndian.Uint16(c.headBuf[14:16])
		if version != kVersion {
			err = fmt.Errorf("invalid magic number: %d", magic)
			glog.Error(err)
			return nil, err
		}

		// 4. read reserved
		// skip reserved c.headBuf[14:16]

		payload := make([]byte, packageLength-kFrameHeaderLen) // recv crc32
		_, err = io.ReadFull(c.conn, payload)
		if err != nil {
			glog.Error(err)
			return nil, err
		}

		sessionId := binary.LittleEndian.Uint64(payload[:8])
		seqNo := binary.LittleEndian.Uint64(payload[8:16])

		// TODO(@benqi): check mdLen
		mdLen := binary.LittleEndian.Uint32(payload[16:20])
		md := &ZProtoMetadata{}
		if mdLen != 0 {
			err = md.Decode(payload[20 : mdLen+20])
			if err != nil {
				glog.Error(err)
				return nil, err
			}
		}

		// glog.Infof("mdLen: %d, md: {%v}, payload: %s%s", mdLen, md, hex.EncodeToString(c.headBuf), hex.EncodeToString(payload))
		packageType := binary.LittleEndian.Uint32(payload[mdLen+20 : mdLen+24])
		m2 := NewZProtoMessage(packageType)
		if m2 == nil {
			err = fmt.Errorf("unregister packageType: %d, payload: %s", packageType, hex.EncodeToString(payload[mdLen+24:]))
			glog.Error(err)
			return nil, err
		}

		err = m2.Decode(payload[mdLen+24 : len(payload)-4])
		if err != nil {
			glog.Error(err)
			return nil, err
		}

		// TODO(@benqi): check crc32
		crc32 := binary.LittleEndian.Uint32(payload[len(payload)-4:])
		_ = crc32

		message := &ZProtoMessage{
			SessionId: sessionId,
			SeqNum:    seqNo,
			Metadata:  md,
			Message:   m2,
		}
		return message, nil
	}
}

func (c *ZProtoCodec) Close() error {
	return c.conn.Close()
}
