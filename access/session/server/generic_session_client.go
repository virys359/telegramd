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

import "github.com/nebulaim/telegramd/mtproto"

type genericSessionClient struct {
	sessionType int
	online      bool
	sessionId   int64
	userId      int32
}

func (s *genericSessionClient) onNewSessionClient(sessionId int64) {
	if s.sessionId != 0 {

	} else {
		if s.sessionId != sessionId {
			// new_session_created
			s.sessionId = sessionId
		} else {
			//
		}
	}
}

func (s *genericSessionClient) onSessionClientData(sessionId, msgId int64, seqNo int32, request mtproto.TLObject) {
	needLogin := !checkWithoutLogin(request)
	if needLogin {
		if s.userId == 0 {
			// getUserId
		}
	} else {
		//
	}
	//case *mtproto.TLMsgsStateReq:
	//case *mtproto.TLMsgResendReq:
	//case *mtproto.TLMsgsAck:
	//	return s.onMsgsAck(msgId, seqNo, request)
	//case *mtproto.TLRpcDropAnswer:
	//	s.onRpcDropAnswer(msgId, seqNo, request)

	// route_table
}

func (s *genericSessionClient) onSessionClientDestroy(sessionId int64) {
}
