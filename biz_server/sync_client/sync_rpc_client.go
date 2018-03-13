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

package sync_client

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

func (c *syncClient) SyncUpdateShortMessage(authKeyId, sessionId, netlibSessionId int64, pushtoUserId, peerId int32, update *mtproto.TLUpdateShortMessage) (reply *mtproto.ClientUpdatesState, err error) {
	m := &mtproto.SyncShortMessageRequest{
		ClientId: &mtproto.PushClientID{
			AuthKeyId:       authKeyId,
			SessionId:       sessionId,
			NetlibSessionId: netlibSessionId,
		},
		PushType:     mtproto.SyncType_SYNC_TYPE_USER_NOTME,
		PushtoUserId: pushtoUserId,
		PeerId:       peerId,
		PushData:     update,
	}
	reply, err = c.client.SyncUpdateShortMessage(context.Background(), m)
	return
}

func (c *syncClient) PushUpdateShortMessage(pushtoUserId, peerId int32, update *mtproto.TLUpdateShortMessage) (reply *mtproto.VoidRsp, err error) {
	m := &mtproto.UpdateShortMessageRequest{
		PushType:     mtproto.SyncType_SYNC_TYPE_USER,
		PushtoUserId: pushtoUserId,
		PeerId:       peerId,
		PushData:     update,
	}
	reply, err = c.client.PushUpdateShortMessage(context.Background(), m)
	return
}

func (c *syncClient) SyncUpdateShortChatMessage(authKeyId, sessionId, netlibSessionId int64, pushtoUserId, peerId int32, update *mtproto.TLUpdateShortChatMessage) (reply *mtproto.ClientUpdatesState, err error) {
	m := &mtproto.SyncShortChatMessageRequest{
		ClientId: &mtproto.PushClientID{
			AuthKeyId: authKeyId,
			SessionId: sessionId,
			NetlibSessionId: netlibSessionId,
		},
		PushType:     mtproto.SyncType_SYNC_TYPE_USER_NOTME,
		PushtoUserId: pushtoUserId,
		PeerChatId:   peerId,
		PushData:     update,
	}
	reply, err = c.client.SyncUpdateShortChatMessage(context.Background(), m)
	return
}

func (c *syncClient) PushUpdateShortChatMessage(senderId, peerId int32, update *mtproto.TLUpdateShortChatMessage) (reply *mtproto.VoidRsp, err error) {
	return
}

func (c *syncClient) SyncUpdateMessageData(authKeyId, sessionId, netlibSessionId int64, pushtoUserId, peerType, peerId int32, update *mtproto.Update) (reply *mtproto.ClientUpdatesState, err error) {
	m := &mtproto.SyncUpdateMessageRequest{
		ClientId: &mtproto.PushClientID{
			AuthKeyId:       authKeyId,
			SessionId:       sessionId,
			NetlibSessionId: netlibSessionId,
		},
		PushType:     mtproto.SyncType_SYNC_TYPE_USER_NOTME,
		PushtoUserId: pushtoUserId,
		PeerType:     peerType,
		PeerId:       peerId,
		PushData:     update,
	}
	reply, err = c.client.SyncUpdateMessageData(context.Background(), m)
	return
}

func (c *syncClient) PushUpdateMessageData(pushtoUserId, peerType, peerId int32, update *mtproto.Update) (reply *mtproto.VoidRsp, err error) {
	m := &mtproto.PushUpdateMessageRequest{
		PushType:     mtproto.SyncType_SYNC_TYPE_USER_NOTME,
		PushtoUserId: pushtoUserId,
		PeerType:     peerType,
		PeerId:       peerId,
		PushData:     update,
	}
	reply, err = c.client.PushUpdateMessageData(context.Background(), m)
	return
}

func (c *syncClient) PushUpdatesData(pushtoUserId int32, updates *mtproto.TLUpdates) (reply *mtproto.VoidRsp, err error) {
	m := &mtproto.PushUpdatesRequest{
		PushType:     mtproto.SyncType_SYNC_TYPE_USER,
		PushtoUserId: pushtoUserId,
		PushData:     updates,
	}
	reply, err = c.client.PushUpdatesData(context.Background(), m)
	return
}

func (c *syncClient) PushUpdateShortData(pushtoUserId int32, update *mtproto.TLUpdateShort) (reply *mtproto.VoidRsp, err error) {
	m := &mtproto.PushUpdateShortRequest{
		PushType:     mtproto.SyncType_SYNC_TYPE_USER,
		PushtoUserId: pushtoUserId,
		PushData:     update,
	}
	reply, err = c.client.PushUpdateShortData(context.Background(), m)
	return
}
