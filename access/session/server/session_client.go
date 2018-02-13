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
	"github.com/golang/glog"
	"github.com/nebulaim/telegramd/mtproto"
	"github.com/nebulaim/telegramd/baselib/net2"
	"time"
)

type sessionClient struct {
	sessionType     int
	clientSession   *clientSession
	sessionId       int64
	nextSeqNo       uint32
	state           int
	sendMessageList []*messageData
	ioCallback      ClientIOCallback
}

type clientSession struct {
	conn      		*net2.TcpConnection
	clientSessionId uint64
}

type messageData struct {
	confirmFlag bool
	compressFlag bool
	obj mtproto.TLObject
}

func (c* sessionClient) encodeMessage(authKeyId int64, authKey []byte, confirm bool, tl mtproto.TLObject) ([]byte, error) {
	message := &mtproto.EncryptedMessage2{
		Salt:      getSalt(),
		SessionId: c.sessionId,
		SeqNo:     c.generateMessageSeqNo(confirm),
		Object:    tl,
	}
	return message.Encode(authKeyId, authKey)
}

func (c* sessionClient) generateMessageSeqNo(increment bool) int32 {
	value := c.nextSeqNo
	if increment {
		c.nextSeqNo++
		return int32(value*2 + 1)
	} else {
		return int32(value*2)
	}
}

// 客户端连接事件
func (c *sessionClient) onSessionClientConnected(conn *net2.TcpConnection, sessionID uint64) {
	c.clientSession = &clientSession{conn, sessionID}
	if c.state == kSessionStateOffline {
		c.state = kSessionStateOnline
	}
}

func (c *sessionClient) onCloseSessionClient() {
	c.clientSession = nil
	if c.state == kSessionStateOnline {
		c.state = kSessionStateOffline
	}
}

func (c *sessionClient) sendToClient(md *mtproto.ZProtoMetadata, obj mtproto.TLObject) error {
	return c.ioCallback.SendToClientData(c, 0, md, []mtproto.TLObject{obj})
}

func (c *sessionClient) sendDataListToClient(md *mtproto.ZProtoMetadata, objs []mtproto.TLObject) error {
	return c.ioCallback.SendToClientData(c, 0, md, objs)
}

// 客户端Session事件
//func (c *sessionClient) onNewSessionClient(sessionId, messageId int64, seqNo int32) {
//}

/*****
	///////////////////////////////
	////////////// System messages
	///////////////////////////////
	// msgs_state_req#da69fb52 msg_ids:Vector<long> = MsgsStateReq;
	// msg_resend_req#7d861a08 msg_ids:Vector<long> = MsgResendReq;
	// msgs_ack#62d6b459 msg_ids:Vector<long> = MsgsAck;
	// http_wait#9299359f max_delay:int wait_after:int max_wait:int = HttpWait;

	//rpc_result#f35c6d01 req_msg_id:long result:Object = RpcResult; // parsed manually
	// message msg_id:long seqno:int bytes:int body:Object = Message; // parsed manually
	// msg_container#73f1f8dc messages:vector<message> = MessageContainer; // parsed manually
	// msg_copy#e06046b2 orig_message:Message = MessageCopy; // parsed manually, not used - use msg_container
	// gzip_packed#3072cfa1 packed_data:string = Object; // parsed manually

	// http_wait#9299359f max_delay:int wait_after:int max_wait:int = HttpWait;
	// help.configSimple#d997c3c5 date:int expires:int dc_id:int ip_port_list:Vector<ipPort> = help.ConfigSimple;


	// rpc_drop_answer#58e4a740 req_msg_id:long = RpcDropAnswer;
	// get_future_salts#b921bd04 num:int = FutureSalts;
	// ping#7abe77ec ping_id:long = Pong;
	// ping_delay_disconnect#f3427b8c ping_id:long disconnect_delay:int = Pong;
	// destroy_session#e7512126 session_id:long = DestroySessionRes;
	// contest.saveDeveloperInfo#9a5f6e95 vk_id:int name:string phone_number:string age:int city:string = Bool;


	///////////////////////////////
	///////// Main application API
	///////////////////////////////
	// invokeAfterMsg#cb9f372d {X:Type} msg_id:long query:!X = X;
	// invokeAfterMsgs#3dc4b4f0 {X:Type} msg_ids:Vector<long> query:!X = X;
	// initConnection#c7481da6 {X:Type} api_id:int device_model:string system_version:string app_version:string system_lang_code:string lang_pack:string lang_code:string query:!X = X;
	// invokeWithLayer#da9b0d0d {X:Type} layer:int query:!X = X;
	// invokeWithoutUpdates#bf9459b7 {X:Type} query:!X = X;
 */
func (c *sessionClient) onSessionClientData(sessDataList *sessionDataList) {
	for _, message := range sessDataList.messages {
		// check new_session_created
		if c.state == kSessionStateCreated {
			c.onNewSessionCreated(message.MsgId, message.Seqno, message)
		}

		switch message.Object.(type) {
		case *mtproto.TLPing:
			c.onPing(message.MsgId, message.Seqno, message)
		case *mtproto.TLPingDelayDisconnect:
			c.onPingDelayDisconnect(message.MsgId, message.Seqno, message)
		case *mtproto.TLGetFutureSalts:
			c.onGetFutureSalts(message.MsgId, message.Seqno, message)
		case *mtproto.TLRpcDropAnswer:
			c.onRpcDropAnswer(message.MsgId, message.Seqno, message)
		case *mtproto.TLDestroySession:
			c.onDestroySession(message.MsgId, message.Seqno, message)
		case *mtproto.TLMsgsAck:
			c.onMsgsAck(message.MsgId, message.Seqno, message)
		case *mtproto.TLHttpWait:
			c.onHttpWait(message.MsgId, message.Seqno, message)
		case *mtproto.TLMsgsStateReq:
			c.onMsgsStateReq(message.MsgId, message.Seqno, message)
		case *mtproto.TLMsgResendReq:
			c.onMsgResendReq(message.MsgId, message.Seqno, message)
		default:
			c.onRpcRequest(message.MsgId, message.Seqno, message)

			//sess := s.getSessionClientBySessionId(sessionId)
			//if sess!= nil {
			//	//
			//} else {
			//	connectionType := getConntctionType(request)
			//	sess := s.getSessionClientByConnectionType(connectionType, sessionId)
			//	sess.onNewSessionClient(sessionId)
			//}
			//sess.onSessionClientData(sessionId, msgId, seqNo, request)
		}
	}
}

func (c *sessionClient) onSessionClientDestroy(sessionId int64) {
}

func (c *sessionClient) onNewSessionCreated(msgId int64, seqNo int32, request mtproto.TLObject) {
	glog.Info("onNewSessionCreated - request data: ", request)
	notify := &mtproto.TLNewSessionCreated{Data2: &mtproto.NewSession_Data{
		FirstMsgId: msgId,
		//// TODO(@benqi): gen new_session_created.unique_id
		UniqueId:   int64(c.clientSession.clientSessionId),
		ServerSalt: getSalt(),
	}}
	c.sendMessageList = append(c.sendMessageList, &messageData{true, false, notify})
	c.state = kSessionStateOnline
}

//func (c *sessionClient) onDestroyAuthKey(msgId int64, seqNo int32, request mtproto.TLObject) {
//	glog.Info("onDestroyAuthKey - request data: ", request)
//	destroyAuthKey, _ := request.(*mtproto.TLDestroyAuthKey)
//	_ = destroyAuthKey
//
//	destroyAuthKeyRes := &mtproto.TLDestroyAuthKeyOk{Data2: &mtproto.DestroyAuthKeyRes_Data{
//	}}
//	c.sendMessageList = append(c.sendMessageList, &messageData{true, false, destroyAuthKeyRes})
//}

func (c *sessionClient) onPing(msgId int64, seqNo int32, request mtproto.TLObject) {
	ping, _ := request.(*mtproto.TLPing)
	glog.Info("processPing - request data: ", ping.String())

	// c.setOnline()
	pong := &mtproto.TLPong{Data2: &mtproto.Pong_Data{
		MsgId: msgId,
		PingId: ping.PingId,
	}}

	c.sendMessageList = append(c.sendMessageList, &messageData{false, false, pong})
}

func (c *sessionClient) onPingDelayDisconnect(msgId int64, seqNo int32, request mtproto.TLObject) {
	pingDelayDissconnect, _ := request.(*mtproto.TLPingDelayDisconnect)
	glog.Info("onPingDelayDisconnect - request data: ", pingDelayDissconnect)

	//// TODO(@benqi): check android client
	//if c.ConnectionType != -1 {
	//	// 1. recv register_device
	//	// 2. save to push pushsession
	//	// 3. check session
	//	if pingDelayDissconnect.GetDisconnectDelay() == 60 * 7 {
	//		c.ConnectionType = PUSH
	//		c.AuthSession.Type = PUSH
	//		UpdateAuthSession(c.Codec.AuthKeyId, c.AuthSession)
	//	}
	//}
	// c.setOnline()
	pong := &mtproto.TLPong{ Data2: &mtproto.Pong_Data{
		MsgId: msgId,
		PingId: pingDelayDissconnect.PingId,
	}}

	c.sendMessageList = append(c.sendMessageList, &messageData{false, false, pong})
}

func (c *sessionClient) onDestroySession(msgId int64, seqNo int32, request mtproto.TLObject) {
	destroySession, _ := request.(*mtproto.TLDestroySession)
	glog.Info("onDestroySession - request data: ", destroySession)

	// TODO(@benqi): 实现destroySession处理逻辑
	destroySessionOk := &mtproto.TLDestroySessionOk{ Data2: &mtproto.DestroySessionRes_Data{
		SessionId: destroySession.SessionId,
	}}
	c.sendMessageList = append(c.sendMessageList, &messageData{false, false, destroySessionOk})
}

func (c *sessionClient) onGetFutureSalts(msgId int64, seqNo int32, request mtproto.TLObject) {
	getFutureSalts, _ := request.(*mtproto.TLGetFutureSalts)
	glog.Info("onGetFutureSalts - request data: ", getFutureSalts)

	// TODO(@benqi): 实现getFutureSalts处理逻辑

	futureSalts := &mtproto.TLFutureSalts{ Data2: &mtproto.FutureSalts_Data{
		ReqMsgId: msgId,
		Now: int32(time.Now().Unix()),
		// Salts: []mtproto.
	}}

	c.sendMessageList = append(c.sendMessageList, &messageData{true, false, futureSalts})
}

// rpc_answer_unknown#5e2ad36e = RpcDropAnswer;
// rpc_answer_dropped_running#cd78e586 = RpcDropAnswer;
// rpc_answer_dropped#a43ad8b7 msg_id:long seq_no:int bytes:int = RpcDropAnswer;
func (c *sessionClient) onRpcDropAnswer(msgId int64, seqNo int32, request mtproto.TLObject) {
	rpcDropAnswer, _ := request.(*mtproto.TLRpcDropAnswer)
	glog.Info("processRpcDropAnswer - request data: ", rpcDropAnswer.String())

	// TODO(@benqi): 实现rpcDropAnswer处理逻辑
	return
	// return nil
}

func (c *sessionClient) onContestSaveDeveloperInfo(msgId int64, seqNo int32, request mtproto.TLObject) {
	contestSaveDeveloperInfo, _ := request.(*mtproto.TLContestSaveDeveloperInfo)
	glog.Info("processGetFutureSalts - request data: ", contestSaveDeveloperInfo.String())
	// TODO(@benqi): 实现scontestSaveDeveloperInfo处理逻辑
	r := &mtproto.TLTrue{}
	_ = r
}

func (c *sessionClient) onMsgsAck(msgId int64, seqNo int32, request mtproto.TLObject) {
	glog.Info("onMsgsAck - request: %s", request.String())
}

func (c *sessionClient) onHttpWait(msgId int64, seqNo int32, request mtproto.TLObject) {
	glog.Info("onHttpWait - request: %s", request.String())
}

func (c *sessionClient) onMsgsStateReq(msgId int64, seqNo int32, request mtproto.TLObject) {
	glog.Info("onMsgsStateReq - request: %s", request.String())
}

func (c *sessionClient) onMsgResendReq(msgId int64, seqNo int32, request mtproto.TLObject) {
	glog.Info("onMsgResendReq - request: %s", request.String())
}

func (c *sessionClient) onRpcRequest(msgId int64, seqNo int32, request mtproto.TLObject) {
	glog.Info("onRpcRequest - request: %s", request.String())
}
