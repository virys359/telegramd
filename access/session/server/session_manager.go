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
	"github.com/nebulaim/telegramd/baselib/net2"
	"github.com/nebulaim/telegramd/mtproto"
	"encoding/binary"
	//"fmt"
	//"encoding/hex"
	//"github.com/golang/glog"
)

type sessionManager struct {
	cache      AuthKeyStorager
	sessions   map[int64]*sessionClientList
	// ioCallback ClientIOCallback
}

func newSessionManager() *sessionManager {
	return &sessionManager{
		cache:      NewAuthKeyCacheManager(),
		sessions:   make(map[int64]*sessionClientList),
		// ioCallback: ioCallback,
	}
}

func (s *sessionManager) onSessionData(conn *net2.TcpConnection, sessionID uint64, md *mtproto.ZProtoMetadata, buf []byte) error {
	authKeyId := int64(binary.LittleEndian.Uint64(buf))
	sess, ok := s.sessions[authKeyId]
	if !ok {
		sess = newSessionClientList(authKeyId, s.cache.GetAuthKey(authKeyId))
		s.sessions[authKeyId] = sess
		// sess.onNewClient(sessionID)
	} else {
		//
	}

	// sessData := newSessionClientData(sess.authKeyId, sess.authKey, zmsg)
	return sess.onSessionClientData(conn, sessionID, md, buf)
}

// 客户端连接建立
func (s *sessionManager) onNewClient() error {
	return nil
}

// TODO(@benqi): 客户端连接断开
func (s *sessionManager) onClientClose() error {
	return nil
}
