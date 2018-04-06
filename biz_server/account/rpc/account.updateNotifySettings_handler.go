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

package rpc

import (
	"github.com/golang/glog"
	"github.com/nebulaim/telegramd/baselib/logger"
	"github.com/nebulaim/telegramd/grpc_util"
	"github.com/nebulaim/telegramd/mtproto"
	"golang.org/x/net/context"
	"github.com/nebulaim/telegramd/biz/base"
	"github.com/nebulaim/telegramd/biz_server/sync_client"
	"github.com/nebulaim/telegramd/biz/core/account"
)

// account.updateNotifySettings#84be5b93 peer:InputNotifyPeer settings:InputPeerNotifySettings = Bool;
func (s *AccountServiceImpl) AccountUpdateNotifySettings(ctx context.Context, request *mtproto.TLAccountUpdateNotifySettings) (*mtproto.Bool, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("account.updateNotifySettings#84be5b93 - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	peer := base.FromInputNotifyPeer(request.GetPeer())
	settings := request.GetSettings().To_InputPeerNotifySettings()

	account.SetNotifySettings(md.UserId, peer, settings)

	update := mtproto.NewTLUpdateNotifySettings()
	update.SetPeer(peer.ToNotifyPeer())
	updateSettings := mtproto.NewTLPeerNotifySettings()
	updateSettings.SetShowPreviews(settings.GetShowPreviews())
	updateSettings.SetSilent(settings.GetSilent())
	updateSettings.SetMuteUntil(settings.GetMuteUntil())
	updateSettings.SetSound(settings.GetSound())
	update.SetNotifySettings(updateSettings.To_PeerNotifySettings())

	sync_client.GetSyncClient().PushToUserMeOneUpdateData(md.AuthId, md.SessionId, md.UserId, update.To_Update())

	glog.Infof("account.updateNotifySettings#84be5b93 - reply: {trur}")
	return mtproto.ToBool(true), nil
}
