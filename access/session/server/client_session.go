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
	"github.com/nebulaim/nebula/sync2"
	"github.com/nebulaim/telegramd/mtproto"
	"github.com/golang/glog"
	"time"
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

type messageData struct {
	confirmFlag  bool
	compressFlag bool
	obj          mtproto.TLObject
}

type clientSession struct {
	connType        int
	clientConnID    ClientConnID
	salt            int64
	nextSeqNo       uint32
	sessionId       int64
	manager			*clientSessionManager
	apiMessages     []*networkApiMessage
	syncMessages    []*networkSyncMessage
	refcount        sync2.AtomicInt32

	// sendMessageList []*messageData
}

func NewClientSession(sessionId int64, m *clientSessionManager) *clientSession{
	return &clientSession{
		connType:        UNKNOWN,
		// clientConnID:    connID,
		sessionId:       sessionId,
		manager:         m,
		apiMessages:     []*networkApiMessage{},
		syncMessages:    []*networkSyncMessage{},
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
func (c *clientSession) AddRef() {
	c.refcount.Add(1)
}

func (c *clientSession) Release() int32 {
	return c.refcount.Add(-1)
}

func (c *clientSession) TimerCallback() {
	// TODO(@benqi): disconnect client conn, notify status server offline
}

//============================================================================================
func (c *clientSession) encodeMessage(authKeyId int64, authKey []byte, confirm bool, tl mtproto.TLObject) ([]byte, error) {
	message := &mtproto.EncryptedMessage2{
		Salt:      c.salt,
		SessionId: c.sessionId,
		SeqNo:     c.generateMessageSeqNo(confirm),
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

func (c *clientSession) sendToClient(connID ClientConnID, md *mtproto.ZProtoMetadata, obj mtproto.TLObject) error {
	glog.Infof("sendToClient - manager: %v", c.manager)

	b, err := c.encodeMessage(c.manager.authKeyId, c.manager.authKey, false, obj)
	if err != nil {
		glog.Error(err)
		return err
	}
	return sendDataByConnID(connID.clientConnID, connID.frontendConnID, md, b)

}

func (c *clientSession) onMessageData(connID ClientConnID, md *mtproto.ZProtoMetadata, messages []*mtproto.TLMessage2) {
	glog.Info("onMessageData - ", messages)
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
		if message.Object == nil {
			continue
		}
		glog.Info("onMessageData - ", message)

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
			c.onMsgsAck(connID, md, message.MsgId, message.Seqno, message.Object)
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
	c.sendToClient(connID, md, pong)
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

	c.sendToClient(connID, md, pong)

	// _ = pong
	// c.sendMessageList = append(c.sendMessageList, &messageData{false, false, pong})

	// 启动定时器
	// timingWheel.AddTimer(c, int(pingDelayDisconnect.DisconnectDelay) + kPingAddTimeout)
}

func (c *clientSession) onMsgsAck(connID ClientConnID, md *mtproto.ZProtoMetadata, msgId int64, seqNo int32, request mtproto.TLObject) {
	glog.Infof("onMsgsAck - request: %s", request.String())
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
	c.sendToClient(connID, md, destroySessionOk)

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

	c.sendToClient(connID, md, futureSalts)

	//c.sendMessageList = append(c.sendMessageList, &messageData{true, false, futureSalts})
}

// rpc_answer_unknown#5e2ad36e = RpcDropAnswer;
// rpc_answer_dropped_running#cd78e586 = RpcDropAnswer;
// rpc_answer_dropped#a43ad8b7 msg_id:long seq_no:int bytes:int = RpcDropAnswer;
func (c *clientSession) onRpcDropAnswer(connID ClientConnID, md *mtproto.ZProtoMetadata, msgId int64, seqNo int32, request *mtproto.TLRpcDropAnswer) {
	glog.Info("processRpcDropAnswer - request data: ", request)
	//
	rpcAnswer := &mtproto.RpcDropAnswer{Data2: &mtproto.RpcDropAnswer_Data{}}

	// TODO(@benqi): 实现rpcDropAnswer处理逻辑
	c.sendToClient(connID, md, rpcAnswer)

	// c.sendMessageList = append(c.sendMessageList, &messageData{false, false, rpcAnswer})
}

func (c *clientSession) onContestSaveDeveloperInfo(connID ClientConnID, md *mtproto.ZProtoMetadata, msgId int64, seqNo int32, request *mtproto.TLContestSaveDeveloperInfo) {
	// contestSaveDeveloperInfo, _ := request.(*mtproto.TLContestSaveDeveloperInfo)
	glog.Info("processGetFutureSalts - request data: ", request)

	// TODO(@benqi): 实现scontestSaveDeveloperInfo处理逻辑
	// r := &mtproto.TLTrue{}
	c.sendToClient(connID, md, &mtproto.TLTrue{})

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

	apiMessage := &networkApiMessage {
		rpcRequest: requestMessage,
		state: kNetworkMessageStateReceived,
	}
	c.apiMessages = append(c.apiMessages, apiMessage)
	c.manager.rpcQueue.Push(&rpcApiMessage{connID: connID, sessionId: c.sessionId, rpcMessage: apiMessage})
}

