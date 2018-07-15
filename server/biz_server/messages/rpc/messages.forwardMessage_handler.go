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
	"github.com/nebulaim/telegramd/proto/mtproto"
	"github.com/nebulaim/telegramd/biz/base"
	message2 "github.com/nebulaim/telegramd/biz/core/message"
	"github.com/nebulaim/telegramd/biz/core/user"
	"golang.org/x/net/context"
	"github.com/nebulaim/telegramd/server/sync/sync_client"
)

// @benqi: android和pc客户端未发现会发送此消息
// messages.forwardMessage#33963bf9 peer:InputPeer id:int random_id:long = Updates;
func (s *MessagesServiceImpl) MessagesForwardMessage(ctx context.Context, request *mtproto.TLMessagesForwardMessage) (*mtproto.Updates, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("messages.forwardMessage#33963bf9 - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	//// peer
	var (
		// fromPeer = helper.FromInputPeer2(md.UserId, request.GetFromPeer())

		peer = base.FromInputPeer2(md.UserId, request.GetPeer())
		messageOutboxList message2.MessageBoxList
	)

	// TODO(@benqi):
	outboxMessages, ridList := makeForwardMessagesData(md.UserId, []int32{request.GetId()}, peer, []int64{request.GetRandomId()})
	for i := 0; i < len(outboxMessages); i++ {
		messageOutbox := message2.CreateMessageOutboxByNew(md.UserId, peer, ridList[i], outboxMessages[i], func(messageId int32) {
			// 更新会话信息
			user.CreateOrUpdateByOutbox(md.UserId, peer.PeerType, peer.PeerId, messageId, outboxMessages[i].GetData2().GetMentioned(), false)
		})
		messageOutboxList = append(messageOutboxList, messageOutbox)
	}

	syncUpdates := makeUpdateNewMessageListUpdates(md.UserId, messageOutboxList)
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
	if request.GetPeer().GetConstructor() != mtproto.TLConstructor_CRC32_inputPeerSelf {
		// var inBoxes message2.MessageBoxList
		var inBoxeMap = map[int32][]*message2.MessageBox{}
		for i := 0; i < len(outboxMessages); i++ {
			inBoxes, _ := messageOutboxList[i].InsertMessageToInbox(md.UserId, peer, func(inBoxUserId, messageId int32) {
				// 更新会话信息
				switch peer.PeerType {
				case base.PEER_USER:
					user.CreateOrUpdateByInbox(inBoxUserId, peer.PeerType, md.UserId, messageId, outboxMessages[i].GetData2().GetMentioned())
				case base.PEER_CHAT, base.PEER_CHANNEL:
					user.CreateOrUpdateByInbox(inBoxUserId, peer.PeerType, peer.PeerId, messageId, outboxMessages[i].GetData2().GetMentioned())
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

		for k, v := range  inBoxeMap {

			syncUpdates = makeUpdateNewMessageListUpdates(k, v)
			sync_client.GetSyncClient().PushToUserUpdatesData(k, syncUpdates.To_Updates())
		}
	}

	glog.Infof("messages.forwardMessage#33963bf9 - reply: %s", logger.JsonDebugData(reply))
	return reply.To_Updates(), nil
}
