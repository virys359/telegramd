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
	"github.com/nebulaim/telegramd/baselib/grpc_util"
	"github.com/nebulaim/telegramd/mtproto"
	"golang.org/x/net/context"
	"github.com/nebulaim/telegramd/biz/core/channel"
	"github.com/nebulaim/telegramd/biz_server/sync_client"
	update2 "github.com/nebulaim/telegramd/biz/core/update"
)

// channels.inviteToChannel#199f3a6c channel:InputChannel users:Vector<InputUser> = Updates;
func (s *ChannelsServiceImpl) ChannelsInviteToChannel(ctx context.Context, request *mtproto.TLChannelsInviteToChannel) (*mtproto.Updates, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("channels.inviteToChannel#199f3a6c - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	if request.Channel.Constructor == mtproto.TLConstructor_CRC32_inputChannelEmpty {
		// TODO(@benqi): chatUser不能是inputUser和inputUserSelf
		err := mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_BAD_REQUEST)
		glog.Error("channels.exportInvite#c7560885 - error: ", err, "; InputPeer invalid")
		return nil, err
	}

	channelLogic, err := channel.NewChannelLogicById(request.GetChannel().GetData2().GetChannelId())
	if err != nil {
		glog.Error("channels.inviteToChannel#199f3a6c - error: ", err)
		return nil, err
	}

	updateChannel := &mtproto.TLUpdateChannel{Data2: &mtproto.Update_Data{
		ChannelId: channelLogic.GetChannelId(),
	}}

	for _, u := range request.Users {
		if u.GetConstructor() == mtproto.TLConstructor_CRC32_inputUserEmpty ||
			u.GetConstructor() == mtproto.TLConstructor_CRC32_inputUserSelf {
			// TODO(@benqi): handle inputUserSelf
			continue
		}
		channelLogic.AddChannelUser(md.UserId, u.GetData2().GetUserId())

		updates := update2.NewUpdatesLogic(u.GetData2().GetUserId())
		updates.AddUpdate(updateChannel.To_Update())
		updates.AddChat(channelLogic.ToChannel(u.GetData2().GetUserId()))
		sync_client.GetSyncClient().PushToUserUpdatesData(u.GetData2().GetUserId(), updates.ToUpdates())
	}

	reply := update2.NewUpdatesLogic(md.UserId)
	reply.AddUpdate(updateChannel.To_Update())
	reply.AddChat(channelLogic.ToChannel(md.UserId))

	glog.Infof("channels.inviteToChannel#199f3a6c - reply: {%v}", reply)
	return reply.ToUpdates(), nil
}
