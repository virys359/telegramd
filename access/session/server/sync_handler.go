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
	"github.com/nebulaim/telegramd/grpc_util"
	"github.com/golang/glog"
	"github.com/gogo/protobuf/proto"
)

type syncHandler struct {
}

func newSyncHandler() *syncHandler {
	s := &syncHandler{}
	return s
}

func (s *syncHandler) onSyncData(conn *net2.TcpConnection, buf []byte) error {
	dbuf := mtproto.NewDecodeBuf(buf)
	len := int(dbuf.Int())
	messageName := string(dbuf.Bytes(len))
	message, err := grpc_util.NewMessageByName(messageName)
	if err != nil {
		glog.Error(err)
		return err
	}

	err = proto.Unmarshal(buf[4+len:], message)
	if err != nil {
		glog.Error(err)
		return err
	}

	switch message.(type) {
	case *mtproto.PushUpdatesNotify:
		glog.Infof("onSyncData - sync data: {%v}", message)
	case *mtproto.PushUpdatesData:
		glog.Infof("onSyncData - sync data: {%v}", message)
	default:
		glog.Errorf("invalid register proto type: {%v}", message)
	}
	return nil
}
