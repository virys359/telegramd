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
	"github.com/golang/glog"
	"time"
	"container/list"
)

// PUSH ==> ConnectionTypePush
// ConnectionTypePush和其它类型不太一样，session一旦创建以后不会改变
const (
	GENERIC  = 0
	DOWNLOAD = 1
	UPLOAD   = 3

	// Android
	PUSH = 7

	// 暂时不考虑
	TEMP = 8

	UNKNOWN = 256

	INVALID_TYPE = -1 // math.MaxInt32
)

const (
	kDefaultPingTimeout = 30
	kPingAddTimeout = 15
)

type messageData struct {
	confirmFlag  bool
	compressFlag bool
	obj          mtproto.TLObject
}

type clientSession struct {
	closeDate       int64
	connType        int
	clientConnID    ClientConnID
	salt            int64
	nextSeqNo       uint32
	sessionId       int64
	manager			*clientSessionManager
	apiMessages     *list.List // []*networkApiMessage
	syncMessages    *list.List // []*networkSyncMessage
	refcount        sync2.AtomicInt32
}

func NewClientSession(sessionId int64, m *clientSessionManager) *clientSession{
	return &clientSession{
		closeDate:       time.Now().Unix() + kDefaultPingTimeout + kPingAddTimeout,
		connType:        UNKNOWN,
		// clientConnID: connID,
		sessionId:       sessionId,
		manager:         m,
		apiMessages:     list.New(), // []*networkApiMessage{},
		syncMessages:    list.New(), // []*networkSyncMessage{},
	}
}

//============================================================================================
func getConnectionType2(messages []*mtproto.TLMessage2) int {
	for _, m := range messages {
		if m.Object != nil {
			//connType := getConnectionType(m.Object)
			//if connType != UNKNOWN {
			//	c.connType = connType
			//	break
			//}
		}
	}

	return 0
}

//============================================================================================
// return false, will delete this clientSession
func (c *clientSession) onTimer() bool {
	date := time.Now().Unix()

	for e := c.apiMessages.Front(); e != nil; e = e.Next() {
		if date - e.Value.(*networkApiMessage).date > 300 {
			c.apiMessages.Remove(e)
		}
	}

	for e := c.syncMessages.Front(); e != nil; e = e.Next() {
		if date - e.Value.(*networkSyncMessage).date > 300 {
			c.apiMessages.Remove(e)
		}
	}

	if date >= c.closeDate {
		return false
	} else {
		return true
	}
}

//func (c *clientSession) AddRef() {
//	c.refcount.Add(1)
//}
//
//func (c *clientSession) Release() int32 {
//	return c.refcount.Add(-1)
//}
//
//func (c *clientSession) TimerCallback() {
//	// TODO(@benqi): disconnect client conn, notify status server offline
//}

//============================================================================================
func (c *clientSession) encodeMessage(authKeyId int64, authKey []byte, messageId int64, confirm bool, tl mtproto.TLObject) ([]byte, error) {
	message := &mtproto.EncryptedMessage2{
		Salt:      c.salt,
		SeqNo:     c.generateMessageSeqNo(confirm),
		MessageId: messageId,
		// mtproto.GenerateMessageId(),
		SessionId: c.sessionId,
		Object:    tl,
	}
	return message.Encode(authKeyId, authKey)
}

func (c *clientSession) generateMessageSeqNo(increment bool) int32 {
	value := c.nextSeqNo
	if increment {
		c.nextSeqNo++
		return int32(value*2 + 1)
	} else {
		return int32(value * 2)
	}
}

func (c *clientSession) sendToClient(connID ClientConnID, md *mtproto.ZProtoMetadata, messageId int64, confirm bool, obj mtproto.TLObject) error {
	// glog.Infof("sendToClient - manager: %v", c.manager)
	b, err := c.encodeMessage(c.manager.authKeyId, c.manager.authKey, messageId, confirm, obj)
	if err != nil {
		glog.Error(err)
		return err
	}
	return sendDataByConnID(connID.clientConnID, connID.frontendConnID, md, b)

}

func (c *clientSession) onSyncData(connID ClientConnID, md *mtproto.ZProtoMetadata, obj mtproto.TLObject) {

}

func (c *clientSession) onNewSessionCreated(connID ClientConnID, md *mtproto.ZProtoMetadata, msgId int64) {
	glog.Info("onNewSessionCreated - request data: ", msgId)
	serverSalt, _ := GetOrInsertSalt(c.manager.authKeyId)
	c.salt = serverSalt
	newSessionCreated := &mtproto.TLNewSessionCreated{Data2: &mtproto.NewSession_Data{
		FirstMsgId: msgId,
		// TODO(@benqi): gen new_session_created.unique_id
		UniqueId:   int64(connID.frontendConnID),
		ServerSalt: serverSalt,
	}}
	c.sendToClient(connID, md, 0, true, newSessionCreated)
}

func (c *clientSession) onMessageData(connID ClientConnID, md *mtproto.ZProtoMetadata, messages []*mtproto.TLMessage2) {
	// glog.Info("onMessageData - ", messages)
	if c.connType == UNKNOWN {
		connType := getConnectionType2(messages)
		if connType != UNKNOWN {
			c.connType = connType
		}
	}

	//if c.connType == GENERIC || c.connType == PUSH {
	//	if c.manager.AuthUserId != 0 {
	//		for _, m := range messages {
	//			if !checkWithoutLogin(m.Object) {
	//				c.manager.AuthUserId = getCacheUserID(c.manager.authKeyId)
	//			}
	//		}
	//	}
	//}
	//

	for _, message := range messages{
		glog.Info("onMessageData - ", message)

		if message.Object == nil {
			continue
		}

		switch message.Object.(type) {
		case *mtproto.TLRpcDropAnswer:	// 所有链接都有可能
			rpcDropAnswer, _ :=  message.Object.(*mtproto.TLRpcDropAnswer)
			c.onRpcDropAnswer(connID, md, message.MsgId, message.Seqno, rpcDropAnswer)
		case *mtproto.TLGetFutureSalts:	// GENERIC
			getFutureSalts, _ := message.Object.(*mtproto.TLGetFutureSalts)
			c.onGetFutureSalts(connID, md, message.MsgId, message.Seqno, getFutureSalts)
		case *mtproto.TLHttpWait:		// 未知
			c.onHttpWait(connID, md, message.MsgId, message.Seqno, message.Object)
		case *mtproto.TLPing:			// android未用
			ping, _ := message.Object.(*mtproto.TLPing)
			c.onPing(connID, md, message.MsgId, message.Seqno, ping)
		case *mtproto.TLPingDelayDisconnect:	// PUSH和GENERIC
			ping, _ := message.Object.(*mtproto.TLPingDelayDisconnect)
			c.onPingDelayDisconnect(connID, md, message.MsgId, message.Seqno, ping)
		case *mtproto.TLDestroySession:			// GENERIC
			destroySession, _ := message.Object.(*mtproto.TLDestroySession)
			c.onDestroySession(connID, md, message.MsgId, message.Seqno, destroySession)
		case *mtproto.TLMsgsAck:				// 所有链接都有可能
			msgsAck, _ := message.Object.(*mtproto.TLMsgsAck)
			c.onMsgsAck(connID, md, message.MsgId, message.Seqno, msgsAck)
		case *mtproto.TLMsgsStateReq:	// android未用
			c.onMsgsStateReq(connID, md, message.MsgId, message.Seqno, message.Object)
		case *mtproto.TLMsgsStateInfo:	// android未用
			c.onMsgsStateInfo(connID, md, message.MsgId, message.Seqno, message.Object)
		case *mtproto.TLMsgsAllInfo:	// android未用
			c.onMsgsAllInfo(connID, md, message.MsgId, message.Seqno, message.Object)
		case *mtproto.TLMsgResendReq:	// 都有可能
			c.onMsgResendReq(connID, md, message.MsgId, message.Seqno, message.Object)
		case *mtproto.TLMsgDetailedInfo:	// 都有可能
			// glog.Error("client side msg: ", object)
		case *mtproto.TLMsgNewDetailedInfo:	// 都有可能
			// glog.Error("client side msg: ", object)
		case *mtproto.TLContestSaveDeveloperInfo:	// 未知
			contestSaveDeveloperInfo, _ := message.Object.(*mtproto.TLContestSaveDeveloperInfo)
			c.onContestSaveDeveloperInfo(connID, md, message.MsgId, message.Seqno, contestSaveDeveloperInfo)
		case *TLInvokeAfterMsgExt:	// 未知
			invokeAfterMsgExt, _ := message.Object.(*TLInvokeAfterMsgExt)
			// c.onRpcRequest(md, message.MsgId, message.Seqno, invokeAfterMsgExt.Query)
			c.onInvokeAfterMsgExt(connID, md, message.MsgId, message.Seqno, invokeAfterMsgExt)
		case *TLInvokeAfterMsgsExt:	// 未知
			invokeAfterMsgsExt, _ := message.Object.(*TLInvokeAfterMsgsExt)
			// c.onRpcRequest(md, message.MsgId, message.Seqno, invokeAfterMsgsExt.Query)
			c.onInvokeAfterMsgsExt(connID, md, message.MsgId, message.Seqno, invokeAfterMsgsExt)
		case *TLInitConnectionExt:	// 都有可能
			initConnectionExt, _ := message.Object.(*TLInitConnectionExt)
			c.onInitConnectionEx(connID, md, message.MsgId, message.Seqno, initConnectionExt)
			// c.onRpcRequest(md, message.MsgId, message.Seqno, initConnectionExt.Query)
		case *TLInvokeWithoutUpdatesExt:
			invokeWithoutUpdatesExt, _ := message.Object.(*TLInvokeWithoutUpdatesExt)
			c.onInvokeWithoutUpdatesExt(connID, md, message.MsgId, message.Seqno, invokeWithoutUpdatesExt)
		default:
			c.onRpcRequest(connID, md, message.MsgId, message.Seqno, message.Object)
		}
	}
}

//============================================================================================
func (c *clientSession) onPing(connID ClientConnID, md *mtproto.ZProtoMetadata, msgId int64, seqNo int32, ping *mtproto.TLPing) {
	// ping, _ := request.(*mtproto.TLPing)
	glog.Info("processPing - request data: ", ping)
	// c.setOnline()
	pong := &mtproto.TLPong{Data2: &mtproto.Pong_Data{
		MsgId:  msgId,
		PingId: ping.PingId,
	}}
	c.sendToClient(connID, md, 0, false, pong)

	c.closeDate = time.Now().Unix() + kDefaultPingTimeout + kPingAddTimeout

	// _ = pong
	// c.sendMessageList = append(c.sendMessageList, &messageData{false, false, pong})

	// 启动定时器
	// timingWheel.AddTimer(c, kDefaultPingTimeout + kPingAddTimeout)
}

func (c *clientSession) onPingDelayDisconnect(connID ClientConnID, md *mtproto.ZProtoMetadata, msgId int64, seqNo int32, pingDelayDisconnect *mtproto.TLPingDelayDisconnect) {
	// pingDelayDisconnect, _ := request.(*mtproto.TLPingDelayDisconnect)
	glog.Info("onPingDelayDisconnect - request data: ", pingDelayDisconnect)
	pong := &mtproto.TLPong{Data2: &mtproto.Pong_Data{
		MsgId:  msgId,
		PingId: pingDelayDisconnect.PingId,
	}}

	c.sendToClient(connID, md, 0, false, pong)

	c.closeDate = time.Now().Unix() + int64(pingDelayDisconnect.DisconnectDelay) + kPingAddTimeout

	// _ = pong
	// c.sendMessageList = append(c.sendMessageList, &messageData{false, false, pong})

	// 启动定时器
	// timingWheel.AddTimer(c, int(pingDelayDisconnect.DisconnectDelay) + kPingAddTimeout)
}

func (c *clientSession) onMsgsAck(connID ClientConnID, md *mtproto.ZProtoMetadata, msgId int64, seqNo int32, request *mtproto.TLMsgsAck) {
	glog.Infof("onMsgsAck - request: %s", request)

	for _, id := range request.GetMsgIds() {
		// reqMsgId := msgId
		for e := c.apiMessages.Front(); e != nil; e = e.Next() {
			v, _ := e.Value.(*networkApiMessage)
			if v.rpcMsgId == id {
				v.state = kNetworkMessageStateAck
				glog.Info("onMsgsAck - networkSyncMessage change kNetworkMessageStateAck")
			}
		}

		for e := c.syncMessages.Front(); e != nil; e = e.Next() {
			v, _ := e.Value.(*networkSyncMessage)
			if v.update.MsgId == id {
				v.state = kNetworkMessageStateAck
				glog.Info("onMsgsAck - networkSyncMessage change kNetworkMessageStateAck")
				// TODO(@benqi): update pts, qts, seq etc...
			}
		}
	}
}

func (c *clientSession) onHttpWait(connID ClientConnID, md *mtproto.ZProtoMetadata, msgId int64, seqNo int32, request mtproto.TLObject) {
	glog.Infof("onHttpWait - request: %s", request.String())
}

func (c *clientSession) onMsgsStateReq(connID ClientConnID, md *mtproto.ZProtoMetadata, msgId int64, seqNo int32, request mtproto.TLObject) {
	glog.Infof("onMsgsStateReq - request: %s", request.String())
}

func (c *clientSession) onInitConnectionEx(connID ClientConnID, md *mtproto.ZProtoMetadata, msgId int64, seqNo int32, request *TLInitConnectionExt) {
	glog.Infof("onInitConnection - request: %s", request.String())
	c.onRpcRequest(connID, md, msgId, seqNo, request.Query)
}

func (c *clientSession) onMsgResendReq(connID ClientConnID, md *mtproto.ZProtoMetadata, msgId int64, seqNo int32, request mtproto.TLObject) {
	glog.Infof("onMsgResendReq - request: %s", request.String())
}

func (c *clientSession) onMsgsStateInfo(connID ClientConnID, md *mtproto.ZProtoMetadata, msgId int64, seqNo int32, request mtproto.TLObject) {
	glog.Infof("onMsgResendReq - request: %s", request.String())
}

func (c *clientSession) onMsgsAllInfo(connID ClientConnID, md *mtproto.ZProtoMetadata, msgId int64, seqNo int32, request mtproto.TLObject) {
	glog.Infof("onMsgResendReq - request: %s", request.String())
}

func (c *clientSession) onDestroySession(connID ClientConnID, md *mtproto.ZProtoMetadata, msgId int64, seqNo int32, request *mtproto.TLDestroySession) {
	glog.Info("onDestroySession - request data: ", request)
	//
	//// TODO(@benqi): 实现destroySession处理逻辑
	destroySessionOk := &mtproto.TLDestroySessionOk{Data2: &mtproto.DestroySessionRes_Data{
		SessionId: request.SessionId,
	}}

	c.sendToClient(connID, md, 0, false, destroySessionOk)

	//c.sendMessageList = append(c.sendMessageList, &messageData{false, false, destroySessionOk})
}

func (c *clientSession) onGetFutureSalts(connID ClientConnID, md *mtproto.ZProtoMetadata, msgId int64, seqNo int32, request *mtproto.TLGetFutureSalts) {
	// getFutureSalts, _ := request.(*mtproto.TLGetFutureSalts)
	glog.Info("onGetFutureSalts - request data: ", request)

	salts, _ := GetOrInsertSaltList(c.manager.authKeyId, int(request.Num))
	futureSalts := &mtproto.TLFutureSalts{Data2: &mtproto.FutureSalts_Data{
		ReqMsgId: msgId,
		Now:      int32(time.Now().Unix()),
		Salts:    salts,
	}}

	c.sendToClient(connID, md, 0, false, futureSalts)
}

// sendToClient:
// 	rpc_answer_unknown#5e2ad36e = RpcDropAnswer;
// 	rpc_answer_dropped_running#cd78e586 = RpcDropAnswer;
// 	rpc_answer_dropped#a43ad8b7 msg_id:long seq_no:int bytes:int = RpcDropAnswer;
func (c *clientSession) onRpcDropAnswer(connID ClientConnID, md *mtproto.ZProtoMetadata, msgId int64, seqNo int32, request *mtproto.TLRpcDropAnswer) {
	glog.Info("processRpcDropAnswer - request data: ", request)
	//
	// reqMsgId := msgId

	rpcAnswer := &mtproto.RpcDropAnswer{Data2: &mtproto.RpcDropAnswer_Data{}}

	var found = false
	for e := c.apiMessages.Front(); e != nil; e = e.Next() {
		v, _ := e.Value.(*networkApiMessage)
		if v.rpcRequest.MsgId == request.ReqMsgId {
			if v.state == kNetworkMessageStateReceived {
				rpcAnswer.Constructor = mtproto.TLConstructor_CRC32_rpc_answer_dropped
				rpcAnswer.Data2.MsgId = request.ReqMsgId
				// TODO(@benqi): set seqno and bytes
				// rpcAnswer.Data2.SeqNo = 0
				// rpcAnswer.Data2.Bytes = 0
			} else if v.state == kNetworkMessageStateInvoked {
				rpcAnswer.Constructor = mtproto.TLConstructor_CRC32_rpc_answer_dropped_running
			} else {
				rpcAnswer.Constructor = mtproto.TLConstructor_CRC32_rpc_answer_unknown
			}
			found = true
			break
		}
	}

	if !found {
		rpcAnswer.Constructor = mtproto.TLConstructor_CRC32_rpc_answer_unknown
	}

	// android client code:
	/*
	 if (notifyServer) {
		TL_rpc_drop_answer *dropAnswer = new TL_rpc_drop_answer();
		dropAnswer->req_msg_id = request->messageId;
		sendRequest(dropAnswer, nullptr, nullptr, RequestFlagEnableUnauthorized | RequestFlagWithoutLogin | RequestFlagFailOnServerErrors, request->datacenterId, request->connectionType, true);
	 }
	 */

	// and both of these responses require an acknowledgment from the client.
	c.sendToClient(connID, md, 0, true, &mtproto.TLRpcResult{ReqMsgId: msgId, Result: rpcAnswer})
}

func (c *clientSession) onContestSaveDeveloperInfo(connID ClientConnID, md *mtproto.ZProtoMetadata, msgId int64, seqNo int32, request *mtproto.TLContestSaveDeveloperInfo) {
	// contestSaveDeveloperInfo, _ := request.(*mtproto.TLContestSaveDeveloperInfo)
	glog.Info("processGetFutureSalts - request data: ", request)

	// TODO(@benqi): 实现scontestSaveDeveloperInfo处理逻辑
	// r := &mtproto.TLTrue{}
	// c.sendToClient(connID, md, false, &mtproto.TLTrue{})

	// _ = r
}

func (c *clientSession) onInvokeAfterMsgExt(connID ClientConnID, md *mtproto.ZProtoMetadata, msgId int64, seqNo int32, request *TLInvokeAfterMsgExt) {
	glog.Info("processInvokeAfterMsg - request data: ", request)
	//		if invokeAfterMsg.GetQuery() == nil {
	//			glog.Errorf("invokeAfterMsg Query is nil, query: {%v}", invokeAfterMsg)
	//			return
	//		}
	//
	//		dbuf := mtproto.NewDecodeBuf(invokeAfterMsg.Query)
	//		query := dbuf.Object()
	//		if query == nil {
	//			glog.Errorf("Decode query error: %s", hex.EncodeToString(invokeAfterMsg.Query))
	//			return
	//		}
	//
	//		var found = false
	//		for j := 0; j < i; j++ {
	//			if messages[j].MsgId == invokeAfterMsg.MsgId {
	//				messages[i].Object = query
	//				found = true
	//				break
	//			}
	//		}
	//
	//		if !found {
	//			for j := i + 1; j < len(messages); j++ {
	//				if messages[j].MsgId == invokeAfterMsg.MsgId {
	//					// c.messages[i].Object = query
	//					messages[i].Object = query
	//					found = true
	//					messages = append(messages, messages[i])
	//
	//					// set messages[i] = nil, will ignore this.
	//					messages[i] = nil
	//					break
	//				}
	//			}
	//		}
	//
	//		if !found {
	//			// TODO(@benqi): backup message, wait.
	//
	//			messages[i].Object = query
	//		}

	c.onRpcRequest(connID, md, msgId, seqNo, request.Query)
}

func (c *clientSession) onInvokeAfterMsgsExt(connID ClientConnID, md *mtproto.ZProtoMetadata, msgId int64, seqNo int32, request *TLInvokeAfterMsgsExt) {
	//		invokeAfterMsgs, _ := messages[i].Object.(*mtproto.TLInvokeAfterMsgs)
	glog.Info("processInvokeAfterMsgs - request data: ", request)
	//		if invokeAfterMsgs.GetQuery() == nil {
	//			glog.Errorf("invokeAfterMsgs Query is nil, query: {%v}", invokeAfterMsgs)
	//			return
	//		}
	//
	//		dbuf := mtproto.NewDecodeBuf(invokeAfterMsgs.Query)
	//		query := dbuf.Object()
	//		if query == nil {
	//			glog.Errorf("Decode query error: %s", hex.EncodeToString(invokeAfterMsgs.Query))
	//			return
	//		}
	//
	//		if len(invokeAfterMsgs.MsgIds) == 0 {
	//			// TODO(@benqi): invalid msgIds, ignore??
	//
	//			messages[i].Object = query
	//		} else {
	//			var maxMsgId = invokeAfterMsgs.MsgIds[0]
	//			for j := 1; j < len(invokeAfterMsgs.MsgIds); j++ {
	//				if maxMsgId > invokeAfterMsgs.MsgIds[j] {
	//					maxMsgId = invokeAfterMsgs.MsgIds[j]
	//				}
	//			}
	//
	//
	//			var found = false
	//			for j := 0; j < i; j++ {
	//				if messages[j].MsgId == maxMsgId {
	//					messages[i].Object = query
	//					found = true
	//					break
	//				}
	//			}
	//
	//			if !found {
	//				for j := i + 1; j < len(messages); j++ {
	//					if messages[j].MsgId == maxMsgId {
	//						// c.messages[i].Object = query
	//						messages[i].Object = query
	//						found = true
	//						messages = append(messages, messages[i])
	//
	//						// set messages[i] = nil, will ignore this.
	//						messages[i] = nil
	//						break
	//					}
	//				}
	//			}
	//
	//			if !found {
	//				// TODO(@benqi): backup message, wait.
	//
	//				messages[i].Object = query
	//			}

	c.onRpcRequest(connID, md, msgId, seqNo, request.Query)
}

func (c *clientSession) onInvokeWithoutUpdatesExt(connID ClientConnID, md *mtproto.ZProtoMetadata, msgId int64, seqNo int32, request *TLInvokeWithoutUpdatesExt) {
	//		glog.Error("android client not use invokeWithoutUpdates: ", messages[i])
	c.onRpcRequest(connID, md, msgId, seqNo, request.Query)
}

func (c *clientSession) onRpcRequest(connID ClientConnID, md *mtproto.ZProtoMetadata, msgId int64, seqNo int32, object mtproto.TLObject) {
	requestMessage := &mtproto.TLMessage2{
		MsgId:  msgId,
		Seqno:  seqNo,
		Object: object,
	}


	// reqMsgId := msgId
	for e := c.apiMessages.Front(); e != nil; e = e.Next() {
		v, _ := e.Value.(*networkApiMessage)
		if v.rpcRequest.MsgId == msgId {
			if v.state >= kNetworkMessageStateInvoked {
				c.sendToClient(connID, md, v.rpcMsgId, true, v.rpcResult)
				return
			}
		}
	}

	apiMessage := &networkApiMessage{
		date:       time.Now().Unix(),
		rpcRequest: requestMessage,
		state:      kNetworkMessageStateReceived,
	}
	// c.apiMessages = append(c.apiMessages, apiMessage)
	c.apiMessages.PushBack(apiMessage)
	c.manager.rpcQueue.Push(&rpcApiMessage{connID: connID, sessionId: c.sessionId, rpcMessage: apiMessage})
}

