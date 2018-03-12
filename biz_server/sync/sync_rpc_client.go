/*
 *  Copyright (c) 2017, https://github.com/nebulaim
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

package sync

import (
	"context"
	"github.com/golang/glog"
	"github.com/nebulaim/telegramd/grpc_util/service_discovery"
	"github.com/nebulaim/telegramd/grpc_util"
	"github.com/nebulaim/telegramd/mtproto"
)

type syncClient struct {
	client mtproto.RPCSyncClient
}

var (
	syncInstance = &syncClient{}
)

func GetSyncClient() *syncClient {
	return syncInstance
}

func InstallSyncClient(discovery *service_discovery.ServiceDiscoveryClientConfig) {
	conn, err := grpc_util.NewRPCClientByServiceDiscovery(discovery)

	if err != nil {
		glog.Error(err)
		panic(err)
	}

	syncInstance.client = mtproto.NewRPCSyncClient(conn)
}

func (c *syncClient) SyncUpdateShortMessage(authKeyId, sessionId, netlibSessionId int64, senderId, peerId int32, update *mtproto.TLUpdateShortMessage) (reply *mtproto.ClientUpdatesState, err error) {
	m := &mtproto.SyncShortMessageRequest{
		ClientId: &mtproto.PushClientID{
			AuthKeyId: authKeyId,
			SessionId: sessionId,
			NetlibSessionId: netlibSessionId,
		},
		SenderUserId: senderId,
		// PeerUserId: peerId,
		PushData: &mtproto.PushShortMessage{
			PushType: mtproto.SyncType_SYNC_TYPE_USER_NOTME,
			PushUserId: peerId,
			PushData: update,
		},
	}
	reply, err = c.client.SyncUpdateShortMessage(context.Background(), m)
	return
}

func (c *syncClient) PushUpdateShortMessage(senderId, peerId int32, update *mtproto.TLUpdateShortMessage) (reply *mtproto.VoidRsp, err error) {
	m := &mtproto.UpdateShortMessageRequest{
		PeerUserId: peerId,
		PushData: &mtproto.PushShortMessage{
			PushType: mtproto.SyncType_SYNC_TYPE_USER,
			PushUserId: peerId,
			PushData: update,
		},
	}
	reply, err = c.client.PushUpdateShortMessage(context.Background(), m)
	return
}

func (c *syncClient) SyncUpdateShortChatMessage(authKeyId, sessionId, netlibSessionId int64, senderId, peerId int32, update *mtproto.TLUpdateShortChatMessage) (reply *mtproto.ClientUpdatesState, err error) {
	m := &mtproto.SyncShortChatMessageRequest{
		ClientId: &mtproto.PushClientID{
			AuthKeyId: authKeyId,
			SessionId: sessionId,
			NetlibSessionId: netlibSessionId,
		},
		SenderUserId: senderId,
		// PeerUserId: peerId,
		PushData: &mtproto.PushShortChatMessage{
			PushType: mtproto.SyncType_SYNC_TYPE_USER_NOTME,
			PushUserId: peerId,
			PushData: update,
		},
	}
	reply, err = c.client.SyncUpdateShortChatMessage(context.Background(), m)
	return
}

func (c *syncClient) PushUpdateShortChatMessage(senderId, peerId int32, update *mtproto.TLUpdateShortChatMessage) (reply *mtproto.VoidRsp, err error) {
	return
}
