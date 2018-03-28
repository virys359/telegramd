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

// 路由
package server

import (
	"github.com/nebulaim/telegramd/mtproto"
	"github.com/golang/glog"
	"encoding/hex"
	// "container/list"
	// "container/list"
	"github.com/nebulaim/telegramd/baselib/net2"
	"fmt"
)

// PUSU ==> ConnectionTypePush
// ConnectionTypePush和其它类型不太一样，session一旦创建以后不会改变
const (
	GENERIC = 0
	DOWNLOAD = 1
	UPLOAD = 3

	// Android
	PUSH = 7

	// 暂时不考虑
	TEMP = 8

	UNKNOWN = 256

	INVALID_TYPE = -1 // math.MaxInt32
)

const (
	DOWNLOAD_CONNECTIONS_COUNT = 2 	// Download conn count
	UPLOAD_CONNECTIONS_COUNT = 4	//
	MAX_CONNECTIONS_COUNT = 9		// 最大连接数为9
)

const (
	kSessionStateUnknown = iota
	kSessionStateCreated
	kSessionStateOnline
	kSessionStateOffline
)

type sessionDataList struct {
	clientSessionId   uint64
	quickAckId  int32
	metadata    *mtproto.ZProtoMetadata
	salt		int64
	sessionId   int64
	messages    []*mtproto.TLMessage2
}

func newSessionDataList(sessionId uint64, md *mtproto.ZProtoMetadata, message *mtproto.EncryptedMessage2) *sessionDataList {
	sessDatas := &sessionDataList{
		clientSessionId: uint64(sessionId),
		metadata:        md,
		salt:            message.Salt,
		sessionId:       message.SessionId,
		// messageId:       message.MessageId,
		// seqNo:           message.SeqNo,
	}
	sessDatas.onMessage(message.MessageId, message.SeqNo, message.Object)
	glog.Info("newSessionDataList - sessDatas: ", sessDatas)
	return sessDatas
}

// TODO(@benqi): handle error
func (this *sessionDataList) onMessage(msgId int64, seqNo int32, object mtproto.TLObject) {
	switch object.(type) {
	case *mtproto.TLMsgContainer:
		msgContainer, _ := object.(*mtproto.TLMsgContainer)
		glog.Info("processMsgContainer - request data: ", msgContainer)
		for _, m := range msgContainer.Messages {
			if m.Object == nil {
				continue
			}
			this.onMessage(m.MsgId, m.Seqno, m.Object)
		}
	case *mtproto.TLGzipPacked:
		gzipPacked, _ := object.(*mtproto.TLGzipPacked)
		glog.Info("processGzipPacked - request data: ", gzipPacked)

		dbuf := mtproto.NewDecodeBuf(gzipPacked.PackedData)
		o := dbuf.Object()
		if o == nil {
			glog.Errorf("Decode query error: %s", hex.EncodeToString(gzipPacked.PackedData))
			return
		}
		// return s.onGzipPacked(sessionId, msgId, seqNo, request)
		this.onMessage(msgId, seqNo, o)
	case *mtproto.TLInvokeAfterMsg:
		invokeAfterMsg, _ := object.(*mtproto.TLInvokeAfterMsg)
		glog.Info("processInvokeAfterMsg - request data: ", object)
		if invokeAfterMsg.GetQuery() == nil {
			glog.Errorf("invokeAfterMsg Query is nil, query: {%v}", invokeAfterMsg)
			return
		}

		dbuf := mtproto.NewDecodeBuf(invokeAfterMsg.Query)
		query := dbuf.Object()
		if query == nil {
			glog.Errorf("Decode query error: %s", hex.EncodeToString(invokeAfterMsg.Query))
			return
		}

		// TODO(@benqi): process invokeAfterMsg.MsgId
		//

		this.onMessage(msgId, seqNo, query)

	case *mtproto.TLInvokeAfterMsgs:
		glog.Info("TLInvokeAfterMsgs - request data: ", object)

		// @benqi: android client not use InvokeAfterMsgs

	case *mtproto.TLInvokeWithLayer:
		invokeWithLayer, _ := object.(*mtproto.TLInvokeWithLayer)
		glog.Info("processInvokeWithLayer - request data: ", object)

		// TODO(@benqi): Check api layer
		// if invokeWithLayer.Layer > API_LAYER {
		// 	return fmt.Errorf("Not suppoer api layer: %d", invokeWithLayer.layer)
		// }

		if invokeWithLayer.GetQuery() == nil {
			glog.Errorf("invokeWithLayer Query is nil, query: {%v}", invokeWithLayer)
		} else {
			dbuf := mtproto.NewDecodeBuf(invokeWithLayer.Query)
			classID := dbuf.Int()
			if classID != int32(mtproto.TLConstructor_CRC32_initConnection) {
				glog.Errorf("Not initConnection classID: %d", classID)
				return
			}

			initConnection := &mtproto.TLInitConnection{}
			err := initConnection.Decode(dbuf)
			if err != nil {
				glog.Error("Decode initConnection error: ", err)
				return
			}
			this.messages = append(this.messages, &mtproto.TLMessage2{MsgId: msgId, Seqno: seqNo, Object: initConnection})
			dbuf = mtproto.NewDecodeBuf(initConnection.Query)
			query := dbuf.Object()
			if query == nil {
				glog.Errorf("Decode query error: %s", hex.EncodeToString(invokeWithLayer.Query))
				return
			}
			this.onMessage(msgId, seqNo, query)
		}

	case *mtproto.TLInvokeWithoutUpdates:
		glog.Info("processInvokeWithoutUpdates - request data: ", object)
		// @benqi: android client not use InvokeAfterMsgs

	// case *mtproto.TLMsgsStateReq:

	default:
		glog.Info("processOthers - request data: ", object)

		this.messages = append(this.messages, &mtproto.TLMessage2{MsgId: msgId, Seqno: seqNo, Object: object})
	}
}

type sessionClientList struct {
	authKeyId  int64
	authKey    []byte
	authUserId int32
	sessions   map[int64]*sessionClient
}

func newSessionClientList(authKeyId int64, authKey []byte) *sessionClientList {
	sessionList := &sessionClientList{
		authKeyId:  authKeyId,
		authKey:    authKey,
		sessions:   make(map[int64]*sessionClient),
	}
	return sessionList
}

//////////////////////////////////////////////////////////////////////////////////////////////
func (s *sessionClientList) onNewClient(sessionID int64) {
}

func (s *sessionClientList) onSessionClientData(conn *net2.TcpConnection, sessionID uint64, md *mtproto.ZProtoMetadata, buf []byte) error {
	message := &mtproto.EncryptedMessage2{}
	err := message.Decode(s.authKeyId, s.authKey, buf[8:])
	_ = err

	sess, ok := s.sessions[message.SessionId]
	if !ok {
		bizRPCClient, _ := getBizRPCClient()
		sess = &sessionClient{
			authKeyId:   s.authKeyId,
			sessionType: UNKNOWN,
			// clientSessionId:      true,
			sessionId:    message.SessionId,
			state:        kSessionStateCreated,
			authUserId:   s.authUserId,
			callback:     s,
			bizRPCClient: bizRPCClient,
		}
		s.sessions[message.SessionId] = sess
		sess.onSessionClientConnected(conn, sessionID)
		// sess.onNewSessionClient(message.SessionId, message.MessageId, message.SeqNo)
	} else {
		sess.onSessionClientConnected(conn, sessionID)
		// if sess.clientSession == nil {
		// 	sess.onSessionClientConnected(conn, sessionID)
		// }
	}

	// check salt
	if !checkSalt(message.Salt) {
		badServerSalt := mtproto.NewTLBadServerSalt()
		badServerSalt.SetBadMsgId(message.MessageId)
		badServerSalt.SetErrorCode(48)
		badServerSalt.SetBadMsgSeqno(message.SeqNo)
		badServerSalt.SetNewServerSalt(getSalt())
		b, _ := sess.encodeMessage(s.authKeyId, s.authKey, false, badServerSalt)
		return sendDataByConnection(conn, sessionID, md, b)
	}

	sessDatas := newSessionDataList(sessionID, md, message)
	if len(sessDatas.messages) == 0 {
		return nil
	}
	if s.authUserId == 0 {
		var hasLoginedMessage = false
		for _, m := range sessDatas.messages {
			if !checkWithoutLogin(m.Object) {
				hasLoginedMessage = true
				break
			}
		}
		if hasLoginedMessage {
			s.authUserId = getUserIDByAuthKeyID(s.authKeyId)
			if s.authUserId == 0 {
				err = fmt.Errorf("recv without login message: %v", sessDatas)
				glog.Errorf("onSessionClientData - authKeyId: %d, error: %v", s.authKeyId, err)
				// TODO(@benqi): close client
				// return err

			} else {
				// 设置所有的客户端
				for _, c := range s.sessions {
					c.authUserId = s.authUserId
				}
			}
			glog.Info("authUserId: ", s.authUserId, ", sessDatas: ", sessDatas)
		}
	} else {
		if sess.authUserId == 0 {
			sess.authUserId = s.authUserId
		}
		glog.Info("authUserId: ", s.authUserId)
	}

	if s.authUserId !=0 {
		// TODO(@benqi): Set user online for a period of time (timeout)
		sess.onUserOnline(1)
	}

	glog.Info("authUserId: ", s.authUserId, ", sessDatas: ", sessDatas)

	sess.onSessionClientData(sessDatas)
	return nil
}

func (s *sessionClientList) onClientClose(sessionID int64) {
}

func (s* sessionClientList) SendToClientData(client *sessionClient, quickAckId int32, md *mtproto.ZProtoMetadata, messages []*messageData) error {
	if client.clientSession == nil || len(messages) == 0 {
		return fmt.Errorf("client offline or messages is nil.")
	}

	if len(messages) == 0 {
	} else if len(messages) == 1 {
		b, err := client.encodeMessage(s.authKeyId, s.authKey, messages[0].confirmFlag, messages[0].obj)
		if err != nil {
			glog.Error(err)
			return err
		}
		return sendDataByConnection(client.clientSession.conn, client.clientSession.clientSessionId, md, b)
	} else {
		// TODO(@benqi): pack msg_container
		for _, m := range messages {
			b, err := client.encodeMessage(s.authKeyId, s.authKey, m.confirmFlag, m.obj)
			if err == nil {
				sendDataByConnection(client.clientSession.conn, client.clientSession.clientSessionId, md, b)
			}
		}
		return nil
	}
	return nil
}
