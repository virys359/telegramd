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
	"net"
	"io"
	"github.com/golang/glog"
	"encoding/hex"
	"github.com/nebulaim/telegramd/baselib/crypto"
	"encoding/binary"
	"fmt"
)

type MTProtoFullCodec struct {
	// conn *net.TCPConn
	stream *AesCTR128Stream
}

func NewMTProtoFullCodec(conn *net.TCPConn, d *crypto.AesCTR128Encrypt, e *crypto.AesCTR128Encrypt) *MTProtoFullCodec {
	return &MTProtoFullCodec{
		// conn:   conn,
		stream: NewAesCTR128Stream(conn, d, e),
	}
}

func (c *MTProtoFullCodec) Receive() (interface{}, error) {
	var size int
	var n int
	var err error

	b := make([]byte, 1)
	n, err = io.ReadFull(c.stream, b)
	if err != nil {
		return nil, err
	}

	// glog.Info("first_byte: ", hex.EncodeToString(b[:1]))
	needAck := bool(b[0] >> 7 == 1)
	_ = needAck

	b[0] = b[0] & 0x7f
	// glog.Info("first_byte2: ", hex.EncodeToString(b[:1]))

	if b[0] < 0x7f {
		size = int(b[0]) << 2
		// glog.Info("size1: ", size)
	} else {
		glog.Info("first_byte2: ", hex.EncodeToString(b[:1]))
		b := make([]byte, 3)
		n, err = io.ReadFull(c.stream, b)
		if err != nil {
			return nil, err
		}
		size = (int(b[0]) | int(b[1])<<8 | int(b[2])<<16) << 2
		// glog.Info("size2: ", size)
	}

	left := size
	buf := make([]byte, size)
	for left > 0 {
		n, err = io.ReadFull(c.stream, buf[size-left:])
		if err != nil {
			return nil, err
		}
		// glog.Info("ReadFull2: ", hex.EncodeToString(buf))
		left -= n
	}

	// TODO(@benqi): process report ack and quickack
	// 截断QuickAck消息，客户端有问题
	if size == 4 {
		glog.Errorf("Server response error: ", int32(binary.LittleEndian.Uint32(buf)))
		return nil, fmt.Errorf("Recv QuickAckMessage, ignore!!!!") //  connId: ", c.stream, ", by client ", m.RemoteAddr())
	}

	authKeyId := int64(binary.LittleEndian.Uint64(buf))
	message := NewMTPRawMessage(authKeyId, 0)
	message.Decode(buf)
	return message, nil

	//var message MessageBase
	//if authKeyId == 0 {
	//	message = NewUnencryptedRawMessage()
	//	// message.Decode(buf[8:])
	//} else {
	//	message = NewEncryptedRawMessage(authKeyId)
	//}
	//
	//err = message.Decode(buf[8:])
	//if err != nil {
	//	glog.Errorf("decode message error: {%v}", err)
	//	return nil, err
	//}
	//
	// return message, nil
}

func (c *MTProtoFullCodec) Send(msg interface{}) error {
	message, ok := msg.(*MTPRawMessage)
	if !ok {
		err := fmt.Errorf("msg type error, only MTPRawMessage, msg: {%v}", msg)
		glog.Error(err)
		return err
	}

	b := message.Encode()

	sb := make([]byte, 4)
	// minus padding
	size := len(b)/4

	if size < 127 {
		sb = []byte{byte(size)}
	} else {
		binary.LittleEndian.PutUint32(sb, uint32(size<<8|127))
	}

	b = append(sb, b...)
	_, err := c.stream.Write(b)

	if err != nil {
		glog.Errorf("Send msg error: %s", err)
	}

	return err
}

func (c *MTProtoFullCodec) Close() error {
	return c.stream.conn.Close()
}

type AesCTR128Stream struct {
	conn      *net.TCPConn
	encrypt *crypto.AesCTR128Encrypt
	decrypt *crypto.AesCTR128Encrypt
}

func NewAesCTR128Stream(conn *net.TCPConn, d *crypto.AesCTR128Encrypt, e *crypto.AesCTR128Encrypt) *AesCTR128Stream {
	return &AesCTR128Stream{
		conn:    conn,
		decrypt: d,
		encrypt: e,
	}
}

func (this *AesCTR128Stream) Read(p []byte) (int, error) {
	n, err := this.conn.Read(p)
	if err == nil {
		this.decrypt.Encrypt(p[:])
		return n, err
	}
	return n, err
}

func (this *AesCTR128Stream) Write(p []byte) (int, error) {
	this.encrypt.Encrypt(p[:])
	return this.conn.Write(p)
}
