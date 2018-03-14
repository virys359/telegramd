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
	"github.com/nebulaim/telegramd/biz_model/base"
	"github.com/nebulaim/telegramd/biz_model/model"
	"github.com/nebulaim/telegramd/biz_server/sync_client"
	"time"
)

func (s *MessagesServiceImpl)makeOutboxMessageBySendMessage(fromId int32, peer *base.PeerUtil, request *mtproto.TLMessagesSendMessage) *mtproto.TLMessage {
	return &mtproto.TLMessage{ Data2: &mtproto.Message_Data{
		Out:          true,
		Silent:       request.GetSilent(),
		FromId:       fromId,
		ToId:         peer.ToPeer(),
		ReplyToMsgId: request.GetReplyToMsgId(),
		Message:      request.GetMessage(),
		Media: &mtproto.MessageMedia{
			Constructor: mtproto.TLConstructor_CRC32_messageMediaEmpty,
			Data2:       &mtproto.MessageMedia_Data{},
		},
		ReplyMarkup: request.GetReplyMarkup(),
		Entities:    request.GetEntities(),
		Date:        int32(time.Now().Unix()),
	}}
}

// 流程：
//  1. 入库: 1）存消息数据，2）分别存到发件箱和收件箱里
//  2. 离线推送
//  3. 在线推送
// messages.sendMessage#fa88427a flags:# no_webpage:flags.1?true silent:flags.5?true background:flags.6?true clear_draft:flags.7?true peer:InputPeer reply_to_msg_id:flags.0?int message:string random_id:long reply_markup:flags.2?ReplyMarkup entities:flags.3?Vector<MessageEntity> = Updates;
func (s *MessagesServiceImpl) MessagesSendMessage(ctx context.Context, request *mtproto.TLMessagesSendMessage) (*mtproto.Updates, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("MessagesSendMessage - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	// TODO(@benqi): check request data invalid
	if request.GetPeer().GetConstructor() ==  mtproto.TLConstructor_CRC32_inputPeerEmpty {
		// retuurn
	}

	// TODO(@benqi): ???
	// request.NoWebpage
	// request.Background
	// request.ClearDraft

	// peer
	var (
		peer *base.PeerUtil
		sentMessage *mtproto.TLUpdateShortSentMessage
		err error
	)

	if request.GetPeer().GetConstructor() ==  mtproto.TLConstructor_CRC32_inputPeerSelf {
		peer = &base.PeerUtil{PeerType: base.PEER_USER, PeerId: md.UserId}
	} else {
		peer = base.FromInputPeer(request.GetPeer())
	}

	// 发件箱
	// sendMessageToOutbox
	outbox := s.makeOutboxMessageBySendMessage(md.UserId, peer, request)
	_, dialogMessageId := model.GetMessageModel().SendMessageToOutbox(md.UserId, peer, request.GetRandomId(), outbox.To_Message())

	shortMessage := model.MessageToUpdateShortMessage(outbox.To_Message())
	state, err := sync_client.GetSyncClient().SyncUpdateShortMessage(md.AuthId, md.SessionId, md.NetlibSessionId, md.UserId, md.UserId, shortMessage)
	if err != nil {
		return nil, err
	}

	sentMessage = model.MessageToUpdateShortSentMessage(outbox.To_Message())
	sentMessage.SetPts(state.Pts)
	sentMessage.SetPtsCount(state.PtsCount)

	if request.GetPeer().GetConstructor() !=  mtproto.TLConstructor_CRC32_inputPeerSelf {
		inBoxes, _ := model.GetMessageModel().SendMessageToInbox(md.UserId, peer, request.GetRandomId(), dialogMessageId, outbox.To_Message())
		for i := 0; i < len(inBoxes.UserIds); i++ {
			shortMessage := model.MessageToUpdateShortMessage(inBoxes.Messages[i])
			sync_client.GetSyncClient().PushUpdateShortMessage(inBoxes.UserIds[i], md.UserId, shortMessage)
		}
	}

	// 收件箱
	glog.Infof("MessagesSendMessage - reply: %s", logger.JsonDebugData(sentMessage))
	return sentMessage.To_Updates(), nil
}
