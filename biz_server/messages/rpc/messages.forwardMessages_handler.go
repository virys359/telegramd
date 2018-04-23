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
	"github.com/nebulaim/telegramd/biz/base"
	"github.com/nebulaim/telegramd/biz/core/message"
)

func makeForwardMessagesData(selfId int32, idList []int32, ridList []int64) ([]*message.MessageBox, []int64) {
	findRandomIdById := func(id int32) int64 {
		for i := 0; i < len(idList); i++ {
			if id == idList[i] {
				return ridList[i]
			}
		}
		return 0
	}
	boxList := message.GetMessageBoxListByMessageIdList(selfId, idList)
	randomIdList := make([]int64, 0, len(boxList))
	for _, box := range boxList {
		randomIdList = append(randomIdList, findRandomIdById(box.MessageId))
	}

	return boxList, randomIdList
}

// messages.forwardMessages#708e0195 flags:# silent:flags.5?true background:flags.6?true with_my_score:flags.8?true from_peer:InputPeer id:Vector<int> random_id:Vector<long> to_peer:InputPeer = Updates;
func (s *MessagesServiceImpl) MessagesForwardMessages(ctx context.Context, request *mtproto.TLMessagesForwardMessages) (*mtproto.Updates, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("MessagesForwardMessages - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	//// peer
	var (
		// fromPeer = helper.FromInputPeer2(md.UserId, request.GetFromPeer())
		toPeer = base.FromInputPeer2(md.UserId, request.GetToPeer())
		// err error
	)

	boxList, ridList := makeForwardMessagesData(md.UserId, request.GetId(), request.GetRandomId())
	for i := 0; i < len(boxList); i++ {
		outboxId, dialogMessageId := message.SendMessageToOutbox(md.UserId, toPeer, ridList[i], boxList[i].Message)
		_ = outboxId
		_ = dialogMessageId
	}

	//shortMessage := model.MessageToUpdateShortMessage(outbox.To_Message())
	//state, err := sync_client.GetSyncClient().SyncUpdatesData(md.AuthId, md.SessionId, md.UserId, shortMessage.To_Updates())
	//if err != nil {
	//	glog.Error(err)
	//	return nil, err
	//}
	//// 更新会话信息
	//model.GetDialogModel().CreateOrUpdateByOutbox(md.UserId, peer.PeerType, peer.PeerId, messageId, outbox.GetMentioned(), request.GetClearDraft())
	//
	//// 返回给客户端
	//sentMessage = model.MessageToUpdateShortSentMessage(outbox.To_Message())
	//sentMessage.SetPts(state.Pts)
	//sentMessage.SetPtsCount(state.PtsCount)


	//
	//if request.GetPeer().GetConstructor() ==  mtproto.TLConstructor_CRC32_inputPeerSelf {
	//	peer = &helper.PeerUtil{PeerType: helper.PEER_USER, PeerId: md.UserId}
	//} else {
	//	peer = helper.FromInputPeer(request.GetPeer())
	//}
	//// SelectDialogMessageListByMessageId
	//forwardMessage := model.GetMessageModel().GetMessageByPeerAndMessageId(md.UserId, request.GetId())
	//// TODO(@benqi): check invalid
	//
	//setEditMessageData(request, editOutbox)
	//
	//syncUpdates := makeUpdateEditMessageUpdates(md.UserId, editOutbox)
	//state, err := sync_client.GetSyncClient().SyncUpdatesData(md.AuthId, md.SessionId, md.UserId, syncUpdates.To_Updates())
	//if err != nil {
	//	return nil, err
	//}
	//SetupUpdatesState(state, syncUpdates)
	//model.GetMessageModel().SaveMessage(editOutbox, md.UserId, request.GetId())
	//
	//// push edit peer message
	//peerEditMessages := model.GetMessageModel().GetPeerDialogMessageListByMessageId(md.UserId, request.GetId())
	//for i := 0; i < len(peerEditMessages.UserIds); i++ {
	//	editMessage := peerEditMessages.Messages[i]
	//	editUserId := peerEditMessages.UserIds[i]
	//
	//	setEditMessageData(request, editMessage)
	//	editUpdates := makeUpdateEditMessageUpdates(editUserId, editMessage)
	//	sync_client.GetSyncClient().PushToUserUpdatesData(editUserId, editUpdates.To_Updates())
	//	model.GetMessageModel().SaveMessage(editMessage, editUserId, editMessage.GetData2().GetId())
	//}

	return nil, fmt.Errorf("Not impl MessagesForwardMessages")
}
