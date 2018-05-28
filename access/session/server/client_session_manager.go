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

package server

import (
	"github.com/nebulaim/telegramd/baselib/sync2"
	"github.com/nebulaim/telegramd/mtproto"
	"sync"
	"github.com/golang/glog"
	"fmt"
	"time"
	"github.com/nebulaim/telegramd/baselib/queue2"
	"encoding/hex"
	"github.com/nebulaim/telegramd/baselib/grpc_util"
	"github.com/nebulaim/telegramd/biz/core/user"
)

// client connID
type ClientConnID struct {
	clientConnID uint64			// client -> frontend netlib connID
	frontendConnID  uint64		// frontend -> session netlib connID
}

func makeClientConnID(clientConnID, frontendConnID uint64) ClientConnID {
	return ClientConnID{clientConnID, frontendConnID}
}

const (
	kNetworkMessageStateNone 				= 0		// created
	kNetworkMessageStateReceived 			= 1		// received from client
	kNetworkMessageStateRunning 			= 2		// invoke api
	kNetworkMessageStateWaitReplyTimeout 	= 5		// invoke timeout
	kNetworkMessageStateInvoked 			= 4		// invoke ok, send to client
	kNetworkMessageStateAcked				= 6		// received client ack
	kNetworkMessageStateWaitAckTimeout		= 7		// wait ack timeout
	kNetworkMessageStateError				= 8		// invalid error
	kNetworkMessageStateEnd					= 9		// end state
)

type networkApiMessage struct {
	quickAckId int32 // 0: not use
	rpcRequest *mtproto.TLMessage2
	state      int
	rpcResult  mtproto.TLObject
}

type networkSyncMessage struct {
	update *mtproto.TLMessage2
	state  int
}

//type clientNetworkApiMessage struct {
//	md              *mtproto.ZProtoMetadata
//	sessionId       int64
//	clientSessionId uint64
//	apiMessage      *networkApiMessage
//}

type rpcApiMessage struct {
	connID     ClientConnID
	md         *mtproto.ZProtoMetadata
	sessionId  int64
	rpcMessage *networkApiMessage
}

//type sessionContainerMessage struct {
//	clientSessionId uint64
//	sessionId       int64
//	quickAckId      int32
//	metadata        *mtproto.ZProtoMetadata
//	salt            int64
//	Layer           int32
//	messages        []*mtproto.TLMessage2
//}
//
//type clientConn interface {
//	getConnType() int
//}
//
//func NewClientConn(connType int) clientConn {
//	switch connType {
//	case GENERIC:
//		return &genericClient{}
//	case DOWNLOAD:
//	case UPLOAD:
//	case PUSH:
//	case TEMP:
//	case UNKNOWN:
//	default:
//	}
//	return &unknownClient{}
//}
//

type sessionData struct {
	connID ClientConnID
	md     *mtproto.ZProtoMetadata
	buf    []byte
}

type syncData struct {
	sessionID int64
	md        *mtproto.ZProtoMetadata
	data      *messageData
}

type clientSessionManager struct {
	Layer           int32
	authKeyId       int64
	authKey         []byte
	AuthUserId      int32
	sessions        map[int64]*clientSession
	bizRPCClient    *grpc_util.RPCClient
	nbfsRPCClient   *grpc_util.RPCClient
	closeChan       chan struct{}
	sessionDataChan chan interface{}	// receive from client
	rpcDataChan     chan interface{}	// rpc reply
	rpcQueue        *queue2.SyncQueue
	finish          sync.WaitGroup
	running         sync2.AtomicInt32
}

func newClientSessionManager(authKeyId int64, authKey []byte, userId int32) *clientSessionManager {
	bizRPCClient, _ := getBizRPCClient()
	nbfsRPCClient, _ := getNbfsRPCClient()

	return &clientSessionManager{
		authKeyId:       authKeyId,
		authKey:         authKey,
		AuthUserId:      userId,
		sessions:        make(map[int64]*clientSession),
		bizRPCClient:    bizRPCClient,
		nbfsRPCClient:   nbfsRPCClient,
		closeChan:       make(chan struct{}),
		sessionDataChan: make(chan interface{}),
		rpcDataChan:     make(chan interface{}),
		rpcQueue:        queue2.NewSyncQueue(),
		finish:          sync.WaitGroup{},
	}
}

func (s *clientSessionManager) Start() {
	s.running.Set(1)
	s.finish.Add(1)
	go s.rpcRunLoop()
	go s.runLoop()
}

func (s *clientSessionManager) Stop() {
	s.running.Set(0)
	s.rpcQueue.Close()
	// close(s.closeChan)
}

func (s *clientSessionManager) runLoop() {
	defer func() {
		s.finish.Done()
		close(s.closeChan)
		s.finish.Wait()
	}()

	for s.running.Get() == 1 {
		select {
		case <-s.closeChan:
			// glog.Info("runLoop -> To Close ", this.String())
			return

		case sessionMsg := <-s.sessionDataChan:
			switch sessionMsg.(type) {
			case *sessionData:
				s.onSessionData(sessionMsg.(*sessionData))
			case *syncData:
				s.onSyncData(sessionMsg.(*syncData))
			default:
				panic("receive invalid type msg")
			}
		case rpcMessage := <-s.rpcDataChan:
			result, _ := rpcMessage.(*rpcApiMessage)
			if sess, ok := s.sessions[result.sessionId]; !ok {
				// sess.onRpcRequest()
				//msgs := sess.apiMessages
				//for _, m := range msgs {
				//	if m.rpcRequest.MsgId == result.rpcMessage.rpcRequest.MsgId {
				//		m.state = kNetworkMessageStateInvoked
				//		m.rpcResult = result.rpcMessage.rpcRequest
				//		// TODO(@benqi): send to client
				//	}
 				//}
				// b, _ := sess.encodeMessage(s.authKeyId, s.authKey, false, result.rpcMessage.rpcResult)
				// sendDataByConnID(result.connID.clientConnID, result.connID.frontendConnID, result.md, b)
			} else {
				sess.sendToClient(result.connID, result.md, result.rpcMessage.rpcResult)
			}
		// case timeout:
			// timer
		default:
		}
	}
}

func (s *clientSessionManager) rpcRunLoop() {
	for {
		apiRequest := s.rpcQueue.Pop()
		if apiRequest == nil {
			// request.state
			continue
		} else {
			request, _ := apiRequest.(*rpcApiMessage)
			s.onRpcRequest(request)
		}
	}
}

func (s *clientSessionManager) OnSessionDataArrived(connID ClientConnID, md *mtproto.ZProtoMetadata, buf []byte) error {
	select {
	case s.sessionDataChan <- &sessionData{connID, md, buf} :
		return nil
	}
	return nil
}

func (s *clientSessionManager) OnSyncDataArrived(sessionID int64, md *mtproto.ZProtoMetadata, data *messageData) error {
	select {
	case s.sessionDataChan <- &syncData{sessionID, md, data} :
		return nil
	}
	return nil
}

type messageListWrapper struct {
	messages []*mtproto.TLMessage2
}

func (s *clientSessionManager) onSessionData(sessionMsg *sessionData) {
	glog.Info("sessionDataChan: ", sessionMsg)
	message := mtproto.NewEncryptedMessage2(s.authKeyId)
	err := message.Decode(s.authKeyId, s.authKey, sessionMsg.buf[8:])
	if err != nil {
		// TODO(@benqi): close frontend conn??
		glog.Error(err)
		return
	}

	//=================================================================================================
	// Check Message Identifier (msg_id)
	//
	// https://core.telegram.org/mtproto/description#message-identifier-msg-id
	// Message Identifier (msg_id)
	//
	// A (time-dependent) 64-bit number used uniquely to identify a message within a session.
	// Client message identifiers are divisible by 4,
	// server message identifiers modulo 4 yield 1 if the message is a response to a client message, and 3 otherwise.
	// Client message identifiers must increase monotonically (within a single session),
	// the same as server message identifiers, and must approximately equal unixtime*2^32.
	// This way, a message identifier points to the approximate moment in time the message was created.
	// A message is rejected over 300 seconds after it is created or 30 seconds
	// before it is created (this is needed to protect from replay attacks).
	// In this situation,
	// it must be re-sent with a different identifier (or placed in a container with a higher identifier).
	// The identifier of a message container must be strictly greater than those of its nested messages.
	//
	// Important: to counter replay-attacks the lower 32 bits of msg_id passed
	// by the client must not be empty and must present a fractional
	// part of the time point when the message was created.
	//
	if message.MessageId % 4 != 0 {
		err = fmt.Errorf("client message identifiers are divisible by 4: %d", message.MessageId)
		glog.Error(err)

		// TODO(@benqi): ignore this message or close client conn??
		return
	}

	if message.MessageId & 0xffffffff == 0 {
		err = fmt.Errorf("the lower 32 bits of msg_id passed by the client must not be empty: %d", message.MessageId)
		glog.Error(err)

		// TODO(@benqi): replay-attack, close client conn.
		return
	}

	//timeMessage := int32(message.MessageId / 4294967296.0)
	//date := int32(time.Now().Unix())
	//if date - timeMessage < 30 || timeMessage - date > 300 {
	//	err = fmt.Errorf("message is rejected over 300 seconds after it is created or 30 seconds: %d", message.MessageId)
	//	glog.Error(err)
	//
	//	// TODO(@benqi): ignore this message or close client conn??
	//	return
	//}

	/*
		if sess.lastMsgId > message.MessageId {
			err = fmt.Errorf("client message identifiers must increase monotonically (within a single session): %d", message.MessageId)
			glog.Error(err)

			// TODO(@benqi): ignore this message or close client conn??
			return
		}
		// TODO(@benqi): Notice of Ignored Error Message
		//
		// https://core.telegram.org/mtproto/service_messages_about_messages
		//

		sess.lastMsgId = message.MessageId

		//=============================================================================================
		// Check Server Salt
		if !CheckBySalt(s.authKeyId, message.Salt) {
			salt, _ := GetOrInsertSalt(s.authKeyId)
			sess.salt = salt
			badServerSalt := mtproto.NewTLBadServerSalt()
			badServerSalt.SetBadMsgId(message.MessageId)
			badServerSalt.SetErrorCode(48)
			badServerSalt.SetBadMsgSeqno(message.SeqNo)
			badServerSalt.SetNewServerSalt(salt)
			b, _ := sess.encodeMessage(s.authKeyId, s.authKey, false, badServerSalt)
			sendDataByConnID(sessionMsg.clientSessionId, sessionMsg.sessionID, sessionMsg.md, b)
			// _ = b
			return
		}

		//=============================================================================================
		// TODO(@benqi): Time Synchronization, https://core.telegram.org/mtproto#time-synchronization
		//
		// Time Synchronization
		//
		// If client time diverges widely from server time,
		// a server may start ignoring client messages,
		// or vice versa, because of an invalid message identifier (which is closely related to creation time).
		// Under these circumstances,
		// the server will send the client a special message containing the correct time and
		// a certain 128-bit salt (either explicitly provided by the client in a special RPC synchronization request or
		// equal to the key of the latest message received from the client during the current session).
		// This message could be the first one in a container that includes other messages
		// (if the time discrepancy is significant but does not as yet result in the client’s messages being ignored).
		//
		// Having received such a message or a container holding it,
		// the client first performs a time synchronization (in effect,
		// simply storing the difference between the server’s time
		// and its own to be able to compute the “correct” time in the future)
		// and then verifies that the message identifiers for correctness.
		//
		// Where a correction has been neglected,
		// the client will have to generate a new session to assure the monotonicity of message identifiers.
		//

		//=============================================================================================
		// Check Message Sequence Number (msg_seqno)
		//
		// https://core.telegram.org/mtproto/description#message-sequence-number-msg-seqno
		// Message Sequence Number (msg_seqno)
		//
		// A 32-bit number equal to twice the number of “content-related” messages
		// (those requiring acknowledgment, and in particular those that are not containers)
		// created by the sender prior to this message and subsequently incremented
		// by one if the current message is a content-related message.
		// A container is always generated after its entire contents; therefore,
		// its sequence number is greater than or equal to the sequence numbers of the messages contained in it.
		//

		if message.SeqNo < sess.lastSeqNo {
			err = fmt.Errorf("sequence number is greater than or equal to the sequence numbers of the messages contained in it: %d", message.SeqNo)
			glog.Error(err)

			// TODO(@benqi): ignore this message or close client conn??
			return
		}
		sess.lastSeqNo = message.SeqNo

		sess.onMessageData(sessionMsg.md, message.MessageId, message.SeqNo, message.Object)
	 */

	var messages = &messageListWrapper{[]*mtproto.TLMessage2{}}
	s.onClientMessage(message.MessageId, message.SeqNo, message.Object, messages)

	sess, ok := s.sessions[message.SessionId]
	if !ok {
		sess = NewClientSession(message.SessionId, s)
		s.sessions[message.SessionId] = sess
	}
	sess.clientConnID = sessionMsg.connID

	//=============================================================================================
	// Check Server Salt
	var salt int64
	if !CheckBySalt(s.authKeyId, message.Salt) {
		salt, _ = GetOrInsertSalt(s.authKeyId)
		sess.salt = salt
		badServerSalt := mtproto.NewTLBadServerSalt()
		badServerSalt.SetBadMsgId(message.MessageId)
		badServerSalt.SetErrorCode(48)
		badServerSalt.SetBadMsgSeqno(message.SeqNo)
		badServerSalt.SetNewServerSalt(salt)
		// b, _ := sess.encodeMessage(s.authKeyId, s.authKey, false, badServerSalt)
		// sendDataByConnID(sessionMsg.connID.clientConnID, sessionMsg.connID.frontendConnID, sessionMsg.md, b)
		// return
	} else {
		salt = message.Salt
	}
	sess.onMessageData(sessionMsg.connID, sessionMsg.md, messages.messages)
}

func (s *clientSessionManager) onSyncData(syncMsg *syncData) {
	sess, ok := s.sessions[syncMsg.sessionID]
	if ok {
		sess.sendToClient(sess.clientConnID, syncMsg.md, syncMsg.data.obj)
	}
}

//==================================================================================================
func (s *clientSessionManager) onClientMessage(msgId int64, seqNo int32, object mtproto.TLObject, messages *messageListWrapper) {
	switch object.(type) {
	case *mtproto.TLMsgContainer:
		msgContainer, _ := object.(*mtproto.TLMsgContainer)
		for _, m := range msgContainer.Messages {
			glog.Info("processMsgContainer - request data: ", m)
			if m.Object == nil {
				continue
			}

			// Check msgId
			//
			// A container is always generated after its entire contents; therefore,
			// its sequence number is greater than or equal to the sequence numbers of the messages contained in it.
			//
			// if  m.Seqno > seqNo {
			//	glog.Errorf("sequence number is greater than or equal to the sequence numbers of the messages contained in it: %d", seqNo)
			//	continue
			// }
			s.onClientMessage(m.MsgId, m.Seqno, m.Object, messages)
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
		s.onClientMessage(msgId, seqNo, o, messages)

	case *mtproto.TLMsgCopy:
		// not use in client
		glog.Error("android client not use msg_copy: ", object)

	case *mtproto.TLInvokeAfterMsg:
		invokeAfterMsg := object.(*mtproto.TLInvokeAfterMsg)
		invokeAfterMsgExt := NewInvokeAfterMsgExt(invokeAfterMsg)
		messages.messages = append(messages.messages, &mtproto.TLMessage2{MsgId: msgId, Seqno: seqNo, Object: invokeAfterMsgExt})

	case *mtproto.TLInvokeAfterMsgs:
		invokeAfterMsgs := object.(*mtproto.TLInvokeAfterMsgs)
		invokeAfterMsgsExt := NewInvokeAfterMsgsExt(invokeAfterMsgs)
		messages.messages = append(messages.messages, &mtproto.TLMessage2{MsgId: msgId, Seqno: seqNo, Object: invokeAfterMsgsExt})

	case *mtproto.TLInvokeWithLayer:
		invokeWithLayer := object.(*mtproto.TLInvokeWithLayer)
		if invokeWithLayer.Layer != s.Layer {
			s.Layer = invokeWithLayer.Layer
			// TODO(@benqi):
		}

		if invokeWithLayer.GetQuery() == nil {
			glog.Errorf("invokeWithLayer Query is nil, query: {%v}", invokeWithLayer)
			return
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

			initConnectionExt := NewInitConnectionExt(initConnection)
			messages.messages = append(messages.messages, &mtproto.TLMessage2{MsgId: msgId, Seqno: seqNo, Object: initConnectionExt})
		}

	case *mtproto.TLInvokeWithoutUpdates:
		glog.Error("android client not use invokeWithoutUpdates: ", object)

	default:
		glog.Info("processOthers - request data: ", object)
		messages.messages = append(messages.messages, &mtproto.TLMessage2{MsgId: msgId, Seqno: seqNo, Object: object})
	}
}

func (s *clientSessionManager) PushApiRequest(apiRequest *mtproto.TLMessage2) {
	s.rpcQueue.Push(apiRequest)
}

func (s *clientSessionManager) onRpcRequest(request *rpcApiMessage) {
	glog.Infof("onRpcRequest - request: {%v}", request)

	var (
		rpcMetadata = &grpc_util.RpcMetadata{}
		err       error
		rpcResult mtproto.TLObject
	)

	if s.AuthUserId == 0 {
		if !checkWithoutLogin(request.rpcMessage.rpcRequest.Object) {
			s.AuthUserId = getCacheUserID(s.authKeyId)
			if s.AuthUserId == 0 {
				glog.Error("not found authUserId")
				return
			}
		}
	} else {
		// set online
		setUserOnline(1, request.connID, s.authKeyId, request.sessionId, s.AuthUserId)
	}

	// 初始化metadata
	rpcMetadata.ServerId = 1
	rpcMetadata.NetlibSessionId = int64(request.connID.clientConnID)
	rpcMetadata.AuthId = s.authKeyId
	rpcMetadata.SessionId = request.sessionId
	// rpcMetadata.ClientAddr = request.md.ClientAddr
	// rpcMetadata.TraceId = request.md.TraceId
	rpcMetadata.SpanId = NextId()
	rpcMetadata.ReceiveTime = time.Now().Unix()
	rpcMetadata.UserId = s.AuthUserId
	rpcMetadata.ClientMsgId = request.rpcMessage.rpcRequest.MsgId
	rpcMetadata.Layer = s.Layer

	// TODO(@benqi): change state.
	request.rpcMessage.state = kNetworkMessageStateRunning

	// TODO(@benqi): rpc proxy
	if checkNbfsRpcRequest(request.rpcMessage.rpcRequest.Object) {
		rpcResult, err = s.nbfsRPCClient.Invoke(rpcMetadata, request.rpcMessage.rpcRequest.Object)
	} else {
		rpcResult, err = s.bizRPCClient.Invoke(rpcMetadata, request.rpcMessage.rpcRequest.Object)
	}

	reply := &mtproto.TLRpcResult{
		ReqMsgId: request.rpcMessage.rpcRequest.MsgId,
	}

	if err != nil {
		glog.Error(err)
		rpcErr, _ := err.(*mtproto.TLRpcError)
		if rpcErr.GetErrorCode() == int32(mtproto.TLRpcErrorCodes_NOTRETURN_CLIENT) {
			return
		}
		reply.Result = rpcErr
		// err.(*mtproto.TLRpcError)

	} else {
		glog.Infof("OnMessage - rpc_result: {%v}\n", rpcResult)
		reply.Result = rpcResult
	}

	request.rpcMessage.state = kNetworkMessageStateInvoked
	request.rpcMessage.rpcResult = reply

	// TODO(@benqi): rseult metadata
	s.rpcDataChan <- request
	//{
	//	sessionId: request.sessionId,
	//	reqMsgId:  request.apiMessage.rpcRequest.MsgId,
	//	rpcResult: reply,
	//}
	// reply

	// c.sendMessageList = append(c.sendMessageList, &messageData{true, false, reply})
	//
	//  // TODO(@benqi): 协议底层处理
	//	if _, ok := request.(*mtproto.TLMessagesSendMedia); ok {
	//		if _, ok := rpcResult.(*mtproto.TLRpcError); !ok {
	//			// TODO(@benqi): 由底层处理，通过多种策略（gzip, msg_container等）来打包并发送给客户端
	//			m := &mtproto.MsgDetailedInfoContainer{Message: &mtproto.EncryptedMessage2{
	//				NeedAck: false,
	//				SeqNo:   seqNo,
	//				Object:  reply,
	//			}}
	//			return c.Session.Send(m)
	//		}
	//	}
}

func setUserOnline(serverId int32, connID ClientConnID, authKeyId, sessionId int64, userId int32) {
	defer func() {
		if r := recover(); r != nil {
			glog.Error(r)
		}
	}()

	status := &user.SessionStatus{
		ServerId:        serverId,
		UserId:          userId,
		AuthKeyId:       authKeyId,
		SessionId:       int64(sessionId),
		NetlibSessionId: int64(connID.clientConnID),
		Now:             time.Now().Unix(),
	}

	user.SetOnline(status)
}
