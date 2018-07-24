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
	"github.com/nebulaim/telegramd/baselib/grpc_util"
	"github.com/nebulaim/telegramd/baselib/logger"
	"github.com/nebulaim/telegramd/biz/base"
	message2 "github.com/nebulaim/telegramd/biz/core/message"
	"github.com/nebulaim/telegramd/proto/mtproto"
	"github.com/nebulaim/telegramd/server/sync/sync_client"
	"golang.org/x/net/context"
	"time"
)

func (s *MessagesServiceImpl) makeForwardMessagesData(selfId int32, idList []int32, peer *base.PeerUtil, ridList []int64) ([]*mtproto.Message, []int64) {
	findRandomIdById := func(id int32) int64 {
		for i := 0; i < len(idList); i++ {
			if id == idList[i] {
				return ridList[i]
			}
		}
		return 0
	}

	// TODO(@benqi): process channel

	// 通过idList找到message
	messages := s.MessageModel.GetMessagesByPeerAndMessageIdList2(selfId, idList)
	randomIdList := make([]int64, 0, len(messages))
	for _, m := range messages {
		// TODO(@benqi): rid is 0
		randomIdList = append(randomIdList, findRandomIdById(m.GetData2().GetId()))

		fwdFrom := &mtproto.TLMessageFwdHeader{Data2: &mtproto.MessageFwdHeader_Data{
			Date:   int32(time.Now().Unix()),
			FromId: m.GetData2().GetFromId(),
		}}

		// make message
		m.Data2.ToId = peer.ToPeer()
		m.Data2.FromId = selfId
		m.Data2.FwdFrom = fwdFrom.To_MessageFwdHeader()
	}

	return messages, randomIdList
}

// messages.forwardMessages#708e0195 flags:# silent:flags.5?true background:flags.6?true with_my_score:flags.8?true grouped:flags.9?true from_peer:InputPeer id:Vector<int> random_id:Vector<long> to_peer:InputPeer = Updates;
func (s *MessagesServiceImpl) MessagesForwardMessages(ctx context.Context, request *mtproto.TLMessagesForwardMessages) (*mtproto.Updates, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("messages.forwardMessages#708e0195 - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	//// peer
	var (
		// fromPeer = helper.FromInputPeer2(md.UserId, request.GetFromPeer())
		peer              = base.FromInputPeer2(md.UserId, request.GetToPeer())
		messageOutboxList message2.MessageBoxList
	)

	outboxMessages, ridList := s.makeForwardMessagesData(md.UserId, request.GetId(), peer, request.GetRandomId())
	for i := 0; i < len(outboxMessages); i++ {
		messageOutbox := s.MessageModel.CreateMessageOutboxByNew(md.UserId, peer, ridList[i], outboxMessages[i], func(messageId int32) {
			// 更新会话信息
			s.UserModel.CreateOrUpdateByOutbox(md.UserId, peer.PeerType, peer.PeerId, messageId, outboxMessages[i].GetData2().GetMentioned(), false)
		})
		messageOutboxList = append(messageOutboxList, messageOutbox)
	}

	syncUpdates := s.makeUpdateNewMessageListUpdates(md.UserId, messageOutboxList)
	state, err := sync_client.GetSyncClient().SyncUpdatesData(md.AuthId, md.SessionId, md.UserId, syncUpdates.To_Updates())
	if err != nil {
		return nil, err
	}

	reply := SetupUpdatesState(state, syncUpdates)
	updateList := []*mtproto.Update{}
	for i := 0; i < len(messageOutboxList); i++ {
		updateMessageID := &mtproto.TLUpdateMessageID{Data2: &mtproto.Update_Data{
			Id_4:     messageOutboxList[i].MessageId,
			RandomId: ridList[i],
		}}
		updateList = append(updateList, updateMessageID.To_Update())
	}
	updateList = append(updateList, reply.GetUpdates()...)

	reply.SetUpdates(updateList)

	/////////////////////////////////////////////////////////////////////////////////////
	// 收件箱
	if request.GetToPeer().GetConstructor() != mtproto.TLConstructor_CRC32_inputPeerSelf {
		// var inBoxes message2.MessageBoxList
		var inBoxeMap = map[int32][]*message2.MessageBox{}
		for i := 0; i < len(outboxMessages); i++ {
			inBoxes, _ := messageOutboxList[i].InsertMessageToInbox(md.UserId, peer, func(inBoxUserId, messageId int32) {
				// 更新会话信息
				switch peer.PeerType {
				case base.PEER_USER:
					s.UserModel.CreateOrUpdateByInbox(inBoxUserId, peer.PeerType, md.UserId, messageId, outboxMessages[i].GetData2().GetMentioned())
				case base.PEER_CHAT, base.PEER_CHANNEL:
					s.UserModel.CreateOrUpdateByInbox(inBoxUserId, peer.PeerType, peer.PeerId, messageId, outboxMessages[i].GetData2().GetMentioned())
				}
			})

			for j := 0; j < len(inBoxes); j++ {
				if boxList, ok := inBoxeMap[inBoxes[j].UserId]; !ok {
					inBoxeMap[inBoxes[j].UserId] = []*message2.MessageBox{inBoxes[j]}
				} else {
					boxList = append(boxList, inBoxes[j])
					inBoxeMap[inBoxes[j].UserId] = boxList
				}
			}
		}

		for k, v := range inBoxeMap {

			syncUpdates = s.makeUpdateNewMessageListUpdates(k, v)
			sync_client.GetSyncClient().PushToUserUpdatesData(k, syncUpdates.To_Updates())
		}
	}

	glog.Infof("messages.forwardMessages#708e0195 - reply: %s", logger.JsonDebugData(reply))
	return reply.To_Updates(), nil

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

	// return nil, fmt.Errorf("Not impl MessagesForwardMessages")
}
