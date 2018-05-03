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
	"github.com/golang/glog"
	"fmt"
	"encoding/hex"
)

type sessionManager struct {
	// sync.RWMutex
	cache      AuthKeyStorager
	sessions   map[int64]*sessionClientList
	// ioCallback ClientIOCallback
}

func newSessionManager(cache AuthKeyStorager) *sessionManager {
	return &sessionManager{
		cache:      cache,
		sessions:   make(map[int64]*sessionClientList),
		// ioCallback: ioCallback,
	}
}

func (s *sessionManager) onSessionData(conn *net2.TcpConnection, sessionID uint64, md *mtproto.ZProtoMetadata, buf []byte) error {
	if len(buf) > 10240 {
		glog.Infof("onSessionData: peer(%v) data: {session_id: %d, md: %v, buf_len: %d, buf: %s, buf_end: %s}",
			conn.RemoteAddr(),
			sessionID,
			md,
			len(buf),
			hex.EncodeToString(buf[:256]),
			hex.EncodeToString(buf[len(buf)-256:]))
	} else {
		glog.Infof("onSessionData: peer(%v) data: {session_id: %d, md: %v, buf_len: %d}", conn.RemoteAddr(), sessionID, md, len(buf))
	}
	authKeyId := int64(binary.LittleEndian.Uint64(buf))
	sess, ok := s.sessions[authKeyId]
	if !ok {
		authKey := s.cache.GetAuthKey(authKeyId)
		if authKey == nil {
			err := fmt.Errorf("onSessionData - not found authKeyId: {%d}", authKeyId)
			glog.Error(err)
			return err
		}
		sess = newSessionClientList(authKeyId, authKey)
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

func (s *sessionManager) pushToSessionData(authKeyId, sessionId int64, md *mtproto.ZProtoMetadata, data *messageData) error {
	// var ok bool
	sessList, ok := s.sessions[authKeyId]
	if !ok {
		err := fmt.Errorf("pushToSessionData - not find sessionList by authKeyId: {%d}", authKeyId)
		glog.Warning(err)
		return err
	}
	sess, ok := sessList.sessions[sessionId]
	if !ok {
		err := fmt.Errorf("pushToSessionData - not find session by sessionId: {%d}", sessionId)
		glog.Warning(err)
		return err
	}

	return sessList.SendToClientData(sess, 0, md, []*messageData{data})
}
