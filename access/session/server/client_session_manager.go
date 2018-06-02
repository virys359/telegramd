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
	kNetworkMessageStateAck  				= 6		// received client ack
	kNetworkMessageStateWaitAckTimeout		= 7		// wait ack timeout
	kNetworkMessageStateError				= 8		// invalid error
	kNetworkMessageStateEnd					= 9		// end state
)

type networkApiMessage struct {
	date       int64
	quickAckId int32 // 0: not use
	rpcRequest *mtproto.TLMessage2
	state      int	// TODO(@benqi): sync.AtomicInt32
	rpcMsgId   int64
	rpcResult  mtproto.TLObject
}

type networkSyncMessage struct {
	date       int64
	update *mtproto.TLMessage2
	state  int
}

type rpcApiMessage struct {
	connID     ClientConnID
	md         *mtproto.ZProtoMetadata
	sessionId  int64
	rpcMessage *networkApiMessage
}

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
				// result.rpcMessage.state = kNetworkMessageStateAcked
				result.rpcMessage.rpcMsgId = mtproto.GenerateMessageId()
				sess.sendToClient(result.connID, result.md, result.rpcMessage.rpcMsgId, true, result.rpcMessage.rpcResult)
			}
		case <-time.After(time.Second):
			var delList = []int64{}
			for k, v := range s.sessions {
				if !v.onTimer() {
					delList = append(delList, k)
				}
			}

			for _, id := range delList {
				delete(s.sessions, id)
			}

			if len(s.sessions) == 0 {
				deleteClientSessionManager(s.authKeyId)
			}
		}
	}

	glog.Info("quit runLoop...")
}

func (s *clientSessionManager) rpcRunLoop() {
	for {
		apiRequest := s.rpcQueue.Pop()
		if apiRequest == nil {
			glog.Info("quit rpcRunLoop...")
			return
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


	if message.MessageId & 0xffffffff == 0 {
		err = fmt.Errorf("the lower 32 bits of msg_id passed by the client must not be empty: %d", message.MessageId)
		glog.Error(err)

		// TODO(@benqi): replay-attack, close client conn.
		return
	}

	/*
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
	sess, ok := s.sessions[message.SessionId]
	if !ok {
		sess = NewClientSession(message.SessionId, message.Salt, message.MessageId, s)
		if !sess.CheckBadServerSalt(sessionMsg.connID, sessionMsg.md, message.MessageId, message.SeqNo, message.Salt) {
			// glog.Error("salt invalid..")
			return
		}
		// sess.salt = salt
		s.sessions[message.SessionId] = sess
		sess.onNewSessionCreated(sessionMsg.connID, sessionMsg.md, message.MessageId)
	} else {
		if !sess.CheckBadServerSalt(sessionMsg.connID, sessionMsg.md, message.MessageId, message.SeqNo, message.Salt) {
			// glog.Error("salt invalid..")
			return
		}

		// New Session Creation Notification
		//
		// The server notifies the client that a new session (from the server’s standpoint)
		// had to be created to handle a client message.
		// If, after this, the server receives a message with an even smaller msg_id within the same session,
		// a similar notification will be generated for this msg_id as well.
		// No such notifications are generated for high msg_id values.
		//
		if message.MessageId < sess.firstMsgId {
			sess.firstMsgId = message.MessageId
			sess.onNewSessionCreated(sessionMsg.connID, sessionMsg.md, message.MessageId)
		}
	}

	_, isContainer := message.Object.(*mtproto.TLMsgContainer)
	if !sess.CheckBadMsgNotification(sessionMsg.connID, sessionMsg.md, message.MessageId, message.SeqNo, isContainer) {
		glog.Error("bad msg invalid..")
		return
	}

	sess.clientConnID = sessionMsg.connID
	sess.onClientMessage(message.MessageId, message.SeqNo, message.Object, messages)
	sess.onMessageData(sessionMsg.connID, sessionMsg.md, messages.messages)
}

func (s *clientSessionManager) onSyncData(syncMsg *syncData) {
	sess, ok := s.sessions[syncMsg.sessionID]
	if ok {
		sess.onSyncData(sess.clientConnID, syncMsg.md, syncMsg.data.obj)
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
