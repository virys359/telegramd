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
	"fmt"
	"github.com/golang/glog"
	"github.com/nebulaim/telegramd/baselib/logger"
	"github.com/nebulaim/telegramd/baselib/grpc_util"
	"github.com/nebulaim/telegramd/mtproto"
	"golang.org/x/net/context"
	"github.com/nebulaim/telegramd/biz/core/channel"
	"github.com/nebulaim/telegramd/biz/dal/dao"
	"github.com/nebulaim/telegramd/biz/core/user"
)

// channels.getParticipant#546dd7a6 channel:InputChannel user_id:InputUser = channels.ChannelParticipant;
func (s *ChannelsServiceImpl) ChannelsGetParticipant(ctx context.Context, request *mtproto.TLChannelsGetParticipant) (*mtproto.Channels_ChannelParticipant, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("channels.getParticipant#546dd7a6 - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	if request.Channel.Constructor == mtproto.TLConstructor_CRC32_inputChannelEmpty {
		err := mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_BAD_REQUEST)
		glog.Error("channels.exportInvite#c7560885 - error: ", err, "; InputPeer invalid")
		return nil, err
	}

	var userId = md.UserId
	if request.UserId.GetConstructor() == mtproto.TLConstructor_CRC32_inputUserEmpty {
		err := mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_BAD_REQUEST)
		glog.Error("channels.exportInvite#c7560885 - error: ", err, "; InputPeer invalid")
		return nil, err
	} else if request.UserId.GetConstructor() == mtproto.TLConstructor_CRC32_inputUser {
		userId = request.UserId.GetData2().GetUserId()
	}

	// GetData2().GetUserId()
	inputChannel := request.GetChannel().To_InputChannel()
	do := dao.GetChannelParticipantsDAO(dao.DB_SLAVE).SelectByUserId(inputChannel.GetChannelId(), userId)
	if do == nil {
		err := fmt.Errorf("not find userId in (%v, %d)", inputChannel, userId)
		glog.Error(err)
		return nil, err
	}

	channelParticipant := &mtproto.TLChannelsChannelParticipant{Data2: &mtproto.Channels_ChannelParticipant_Data{
		Participant: channel.MakeChannelParticipant2ByDO(md.UserId, do),
		Users: user.GetUsersBySelfAndIDList(md.UserId, []int32{do.UserId, do.InviterUserId}),
	}}

	glog.Infof("channels.getParticipant#546dd7a6 - reply: {%v}", channelParticipant)
	return channelParticipant.To_Channels_ChannelParticipant(), nil
}
