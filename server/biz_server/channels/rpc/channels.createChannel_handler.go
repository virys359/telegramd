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
	"golang.org/x/net/context"
	// "time"
	"github.com/nebulaim/telegramd/biz/core/channel"
	"github.com/nebulaim/telegramd/biz/core/message"
	"github.com/nebulaim/telegramd/biz/base"
	update2 "github.com/nebulaim/telegramd/biz/core/update"
	"github.com/nebulaim/telegramd/server/sync/sync_client"
	"github.com/nebulaim/telegramd/biz/core/user"
)

/*
 request:
	body: { channels_createChannel
	  flags: 1 [INT],
	  broadcast: YES [ BY BIT 0 IN FIELD flags ],
	  megagroup: [ SKIPPED BY BIT 1 IN FIELD flags ],
	  title: "channel-001" [STRING],
	  about: "channel-001" [STRING],
	},

 response:
    result: { updates
      updates: [ vector<0x0>
        { updateMessageID
          id: 1 [INT],
          random_id: 4140694743218070306 [LONG],
        },
        { updateChannel
          channel_id: 1261873857 [INT],
        },
        { updateReadChannelInbox
          channel_id: 1261873857 [INT],
          max_id: 1 [INT],
        },
        { updateNewChannelMessage
          message: { messageService
            flags: 16386 [INT],
            out: YES [ BY BIT 1 IN FIELD flags ],
            mentioned: [ SKIPPED BY BIT 4 IN FIELD flags ],
            media_unread: [ SKIPPED BY BIT 5 IN FIELD flags ],
            silent: [ SKIPPED BY BIT 13 IN FIELD flags ],
            post: YES [ BY BIT 14 IN FIELD flags ],
            id: 1 [INT],
            from_id: [ SKIPPED BY BIT 8 IN FIELD flags ],
            to_id: { peerChannel
              channel_id: 1261873857 [INT],
            },
            reply_to_msg_id: [ SKIPPED BY BIT 3 IN FIELD flags ],
            date: 1529328456 [INT],
            action: { messageActionChannelCreate
              title: "channel-001" [STRING],
            },
          },
          pts: 2 [INT],
          pts_count: 1 [INT],
        },
      ],
      users: [ vector<0x0> ],
      chats: [ vector<0x0>
        { channel
          flags: 8225 [INT],
          creator: YES [ BY BIT 0 IN FIELD flags ],
          left: [ SKIPPED BY BIT 2 IN FIELD flags ],
          editor: [ SKIPPED BY BIT 3 IN FIELD flags ],
          broadcast: YES [ BY BIT 5 IN FIELD flags ],
          verified: [ SKIPPED BY BIT 7 IN FIELD flags ],
          megagroup: [ SKIPPED BY BIT 8 IN FIELD flags ],
          restricted: [ SKIPPED BY BIT 9 IN FIELD flags ],
          democracy: [ SKIPPED BY BIT 10 IN FIELD flags ],
          signatures: [ SKIPPED BY BIT 11 IN FIELD flags ],
          min: [ SKIPPED BY BIT 12 IN FIELD flags ],
          id: 1261873857 [INT],
          access_hash: 18367393077902002260 [LONG],
          title: "channel-001" [STRING],
          username: [ SKIPPED BY BIT 6 IN FIELD flags ],
          photo: { chatPhotoEmpty },
          date: 1529328455 [INT],
          version: 0 [INT],
          restriction_reason: [ SKIPPED BY BIT 9 IN FIELD flags ],
          admin_rights: [ SKIPPED BY BIT 14 IN FIELD flags ],
          banned_rights: [ SKIPPED BY BIT 15 IN FIELD flags ],
        },
      ],
      date: 1529328455 [INT],
      seq: 0 [INT],
    },
 */

// channels.createChannel#f4893d7f flags:# broadcast:flags.0?true megagroup:flags.1?true title:string about:string = Updates;
func (s *ChannelsServiceImpl) ChannelsCreateChannel(ctx context.Context, request *mtproto.TLChannelsCreateChannel) (*mtproto.Updates, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("channels.createChannel#f4893d7f - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	// logic.NewChatLogicByCreateChat()
	//// TODO(@benqi): Impl MessagesCreateChat logic
	//randomId := md.ClientMsgId

	channelUserIdList := []int32{md.UserId}
	channel := channel.NewChannelLogicByCreateChannel(md.UserId, channelUserIdList, request.GetTitle(), request.GetAbout())

	peer := &base.PeerUtil{
		PeerType: base.PEER_CHANNEL,
		PeerId: channel.GetChannelId(),
	}

	createChannelMessage := channel.MakeCreateChannelMessage(md.UserId)
	randomId := base.NextSnowflakeId()

	// 1. 创建channel
	// 2. 创建channel createChannel message
	
	channelBox := message.CreateChannelMessageBoxByNew(md.UserId, channel.GetChannelId(), randomId, createChannelMessage, func(messageId int32) {
		user.CreateOrUpdateByOutbox(md.UserId, peer.PeerType, peer.PeerId, messageId, false, false)
	})


	syncUpdates := update2.NewUpdatesLogic(md.UserId)
	//updateChatParticipants := &mtproto.TLUpdateChatParticipants{Data2: &mtproto.Update_Data{
	//	Participants: channel.GetChannelParticipants().To_Channels_ChannelParticipants(),
	//}}
	//syncUpdates.AddUpdate(updateChatParticipants.To_Update())
	syncUpdates.AddUpdateNewMessage(createChannelMessage)
	// syncUpdates.AddUsers(user.GetUsersBySelfAndIDList(md.UserId, chat.GetChatParticipantIdList()))
	syncUpdates.AddChat(channel.ToChannel(md.UserId))

	state, _ := sync_client.GetSyncClient().SyncUpdatesData(md.AuthId, md.SessionId, md.UserId, syncUpdates.ToUpdates())
	syncUpdates.AddUpdateMessageId(channelBox.ChannelMessageBoxId, randomId)
	updateChannel := &mtproto.TLUpdateChannel{Data2: &mtproto.Update_Data{
		ChannelId: channel.GetChannelId(),
	}}
	syncUpdates.AddUpdate(updateChannel.To_Update())
	syncUpdates.SetupState(state)
	reply := syncUpdates.ToUpdates()

	//inboxList, _ := outbox.InsertMessageToInbox(md.UserId, peer, func(inBoxUserId, messageId int32) {
	//	user.CreateOrUpdateByInbox(inBoxUserId, base.PEER_CHANNEL, peer.PeerId, messageId, false)
	//})
	//for _, inbox := range inboxList {
	//	updates := update2.NewUpdatesLogic(md.UserId)
	//	//updateChatParticipants := &mtproto.TLUpdateChatParticipants{Data2: &mtproto.Update_Data{
	//	//	Participants: chat.GetChatParticipants().To_ChatParticipants(),
	//	//}}
	//	//updates.AddUpdate(updateChatParticipants.To_Update())
	//	updates.AddUpdateNewMessage(inbox.Message)
	//	//updates.AddUsers(user.GetUsersBySelfAndIDList(md.UserId, chat.GetChatParticipantIdList()))
	//	updates.AddChat(channel.ToChannel(inbox.UserId))
	//	sync_client.GetSyncClient().PushToUserUpdatesData(inbox.UserId, updates.ToUpdates())
	//}

	glog.Infof("channels.createChannel#f4893d7f - reply: {%v}", reply)
	return reply, nil
}
