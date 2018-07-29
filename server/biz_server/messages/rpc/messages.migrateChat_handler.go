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
	"github.com/nebulaim/telegramd/baselib/grpc_util"
	"github.com/nebulaim/telegramd/baselib/logger"
	"github.com/nebulaim/telegramd/proto/mtproto"
	"golang.org/x/net/context"
)

/*

	msg_id: 6583462309624893992 [LONG],
	seq_no: 19275 [INT],
	bytes: 8 [INT],
	body: { messages_migrateChat
	  chat_id: 310890386 [INT],
	},

  body: { rpc_result
    req_msg_id: 6583462309624893992 [LONG],
    result: [GZIPPED] { updates
      updates: [ vector<0x0>
        { updateChannel
          channel_id: 1179206386 [INT],
        },
        { updateNotifySettings
          peer: { notifyPeer
            peer: { peerChannel
              channel_id: 1179206386 [INT],
            },
          },
          notify_settings: { peerNotifySettings
            flags: 0 [INT],
            show_previews: [ SKIPPED BY BIT 0 IN FIELD flags ],
            silent: [ SKIPPED BY BIT 1 IN FIELD flags ],
            mute_until: [ SKIPPED BY BIT 2 IN FIELD flags ],
            sound: [ SKIPPED BY BIT 3 IN FIELD flags ],
          },
        },
        { updateReadChannelInbox
          channel_id: 1179206386 [INT],
          max_id: 1 [INT],
        },
        { updateNewChannelMessage
          message: { messageService
            flags: 258 [INT],
            out: YES [ BY BIT 1 IN FIELD flags ],
            mentioned: [ SKIPPED BY BIT 4 IN FIELD flags ],
            media_unread: [ SKIPPED BY BIT 5 IN FIELD flags ],
            silent: [ SKIPPED BY BIT 13 IN FIELD flags ],
            post: [ SKIPPED BY BIT 14 IN FIELD flags ],
            id: 1 [INT],
            from_id: 264696845 [INT],
            to_id: { peerChannel
              channel_id: 1179206386 [INT],
            },
            reply_to_msg_id: [ SKIPPED BY BIT 3 IN FIELD flags ],
            date: 1532831961 [INT],
            action: { messageActionChannelMigrateFrom
              title: "1111" [STRING],
              chat_id: 310890386 [INT],
            },
          },
          pts: 2 [INT],
          pts_count: 1 [INT],
        },
        { updateNewMessage
          message: { messageService
            flags: 258 [INT],
            out: YES [ BY BIT 1 IN FIELD flags ],
            mentioned: [ SKIPPED BY BIT 4 IN FIELD flags ],
            media_unread: [ SKIPPED BY BIT 5 IN FIELD flags ],
            silent: [ SKIPPED BY BIT 13 IN FIELD flags ],
            post: [ SKIPPED BY BIT 14 IN FIELD flags ],
            id: 2266 [INT],
            from_id: 264696845 [INT],
            to_id: { peerChat
              chat_id: 310890386 [INT],
            },
            reply_to_msg_id: [ SKIPPED BY BIT 3 IN FIELD flags ],
            date: 1532831961 [INT],
            action: { messageActionChatMigrateTo
              channel_id: 1179206386 [INT],
            },
          },
          pts: 4505 [INT],
          pts_count: 1 [INT],
        },
        { updateReadHistoryOutbox
          peer: { peerChat
            chat_id: 310890386 [INT],
          },
          max_id: 2266 [INT],
          pts: 4506 [INT],
          pts_count: 1 [INT],
        },
      ],
      users: [ vector<0x0>
        { user
          flags: 7291 [INT],
          self: YES [ BY BIT 10 IN FIELD flags ],
          contact: YES [ BY BIT 11 IN FIELD flags ],
          mutual_contact: YES [ BY BIT 12 IN FIELD flags ],
          deleted: [ SKIPPED BY BIT 13 IN FIELD flags ],
          bot: [ SKIPPED BY BIT 14 IN FIELD flags ],
          bot_chat_history: [ SKIPPED BY BIT 15 IN FIELD flags ],
          bot_nochats: [ SKIPPED BY BIT 16 IN FIELD flags ],
          verified: [ SKIPPED BY BIT 17 IN FIELD flags ],
          restricted: [ SKIPPED BY BIT 18 IN FIELD flags ],
          min: [ SKIPPED BY BIT 20 IN FIELD flags ],
          bot_inline_geo: [ SKIPPED BY BIT 21 IN FIELD flags ],
          id: 264696845 [INT],
          access_hash: 15537087501845505974 [LONG],
          first_name: "本起" [STRING],
          last_name: [ SKIPPED BY BIT 2 IN FIELD flags ],
          username: "benqi" [STRING],
          phone: "8613606512716" [STRING],
          photo: { userProfilePhoto
            photo_id: 1136864293085620150 [LONG],
            photo_small: { fileLocation
              dc_id: 5 [INT],
              volume_id: 852832478 [LONG],
              local_id: 60529 [INT],
              secret: 9338736479069608191 [LONG],
            },
            photo_big: { fileLocation
              dc_id: 5 [INT],
              volume_id: 852832478 [LONG],
              local_id: 60531 [INT],
              secret: 1214473206298316587 [LONG],
            },
          },
          status: { userStatusOnline
            expires: 1532832234 [INT],
          },
          bot_info_version: [ SKIPPED BY BIT 14 IN FIELD flags ],
          restriction_reason: [ SKIPPED BY BIT 18 IN FIELD flags ],
          bot_inline_placeholder: [ SKIPPED BY BIT 19 IN FIELD flags ],
          lang_code: [ SKIPPED BY BIT 22 IN FIELD flags ],
        },
      ],
      chats: [ vector<0x0>
        { chat
          flags: 97 [INT],
          creator: YES [ BY BIT 0 IN FIELD flags ],
          kicked: [ SKIPPED BY BIT 1 IN FIELD flags ],
          left: [ SKIPPED BY BIT 2 IN FIELD flags ],
          admins_enabled: [ SKIPPED BY BIT 3 IN FIELD flags ],
          admin: [ SKIPPED BY BIT 4 IN FIELD flags ],
          deactivated: YES [ BY BIT 5 IN FIELD flags ],
          id: 310890386 [INT],
          title: "1111" [STRING],
          photo: { chatPhotoEmpty },
          participants_count: 2 [INT],
          date: 1523857808 [INT],
          version: 9 [INT],
          migrated_to: { inputChannel
            channel_id: 1179206386 [INT],
            access_hash: 10847345435675241007 [LONG],
          },
        },
        { channel
          flags: 9473 [INT],
          creator: YES [ BY BIT 0 IN FIELD flags ],
          left: [ SKIPPED BY BIT 2 IN FIELD flags ],
          editor: [ SKIPPED BY BIT 3 IN FIELD flags ],
          broadcast: [ SKIPPED BY BIT 5 IN FIELD flags ],
          verified: [ SKIPPED BY BIT 7 IN FIELD flags ],
          megagroup: YES [ BY BIT 8 IN FIELD flags ],
          restricted: [ SKIPPED BY BIT 9 IN FIELD flags ],
          democracy: YES [ BY BIT 10 IN FIELD flags ],
          signatures: [ SKIPPED BY BIT 11 IN FIELD flags ],
          min: [ SKIPPED BY BIT 12 IN FIELD flags ],
          id: 1179206386 [INT],
          access_hash: 10847345435675241007 [LONG],
          title: "1111" [STRING],
          username: [ SKIPPED BY BIT 6 IN FIELD flags ],
          photo: { chatPhotoEmpty },
          date: 1532831959 [INT],
          version: 0 [INT],
          restriction_reason: [ SKIPPED BY BIT 9 IN FIELD flags ],
          admin_rights: [ SKIPPED BY BIT 14 IN FIELD flags ],
          banned_rights: [ SKIPPED BY BIT 15 IN FIELD flags ],
          participants_count: [ SKIPPED BY BIT 17 IN FIELD flags ],
        },
      ],
      date: 1532831960 [INT],
      seq: 0 [INT],
    },



## updates:
- 1. updates
  body: { updates
    updates: [ vector<0x0>
      { updateChannel
        channel_id: 1179206386 [INT],
      },
      { updateNotifySettings
        peer: { notifyPeer
          peer: { peerChannel
            channel_id: 1179206386 [INT],
          },
        },
        notify_settings: { peerNotifySettings
          flags: 1 [INT],
          show_previews: YES [ BY BIT 0 IN FIELD flags ],
          silent: [ SKIPPED BY BIT 1 IN FIELD flags ],
          mute_until: 0 [INT],
          sound: "default" [STRING],
        },
      },
    ],
    users: [ vector<0x0> ],
    chats: [ vector<0x0>
      { channel
        flags: 9472 [INT],
        creator: [ SKIPPED BY BIT 0 IN FIELD flags ],
        left: [ SKIPPED BY BIT 2 IN FIELD flags ],
        editor: [ SKIPPED BY BIT 3 IN FIELD flags ],
        broadcast: [ SKIPPED BY BIT 5 IN FIELD flags ],
        verified: [ SKIPPED BY BIT 7 IN FIELD flags ],
        megagroup: YES [ BY BIT 8 IN FIELD flags ],
        restricted: [ SKIPPED BY BIT 9 IN FIELD flags ],
        democracy: YES [ BY BIT 10 IN FIELD flags ],
        signatures: [ SKIPPED BY BIT 11 IN FIELD flags ],
        min: [ SKIPPED BY BIT 12 IN FIELD flags ],
        id: 1179206386 [INT],
        access_hash: 2629819016532760918 [LONG],
        title: "1111" [STRING],
        username: [ SKIPPED BY BIT 6 IN FIELD flags ],
        photo: { chatPhotoEmpty },
        date: 1532831961 [INT],
        version: 0 [INT],
        restriction_reason: [ SKIPPED BY BIT 9 IN FIELD flags ],
        admin_rights: [ SKIPPED BY BIT 14 IN FIELD flags ],
        banned_rights: [ SKIPPED BY BIT 15 IN FIELD flags ],
      },
    ],
    date: 1532831960 [INT],
    seq: 4617 [INT],
  },

- 2.
	body: { channels_getParticipant
	  channel: { inputChannel
		channel_id: 1179206386 [INT],
		access_hash: 2629819016532760918 [LONG],
	  },
	  user_id: { inputUserSelf },
	},

  body: { rpc_result
    req_msg_id: 6583463147519930592 [LONG],
    result: [GZIPPED] { channels_channelParticipant
      participant: { channelParticipantSelf
        user_id: 448603711 [INT],
        inviter_id: 264696845 [INT],
        date: 1532831960 [INT],
      },
      users: [ vector<0x0>
        { user
          flags: 3195 [INT],
          self: YES [ BY BIT 10 IN FIELD flags ],
          contact: YES [ BY BIT 11 IN FIELD flags ],
          mutual_contact: [ SKIPPED BY BIT 12 IN FIELD flags ],
          deleted: [ SKIPPED BY BIT 13 IN FIELD flags ],
          bot: [ SKIPPED BY BIT 14 IN FIELD flags ],
          bot_chat_history: [ SKIPPED BY BIT 15 IN FIELD flags ],
          bot_nochats: [ SKIPPED BY BIT 16 IN FIELD flags ],
          verified: [ SKIPPED BY BIT 17 IN FIELD flags ],
          restricted: [ SKIPPED BY BIT 18 IN FIELD flags ],
          min: [ SKIPPED BY BIT 20 IN FIELD flags ],
          bot_inline_geo: [ SKIPPED BY BIT 21 IN FIELD flags ],
          id: 448603711 [INT],
          access_hash: 16248249830026626652 [LONG],
          first_name: "璧君😊" [STRING],
          last_name: [ SKIPPED BY BIT 2 IN FIELD flags ],
          username: "yubijun2" [STRING],
          phone: "8613588021430" [STRING],
          photo: { userProfilePhoto
            photo_id: 1926738268065474478 [LONG],
            photo_small: { fileLocation
              dc_id: 5 [INT],
              volume_id: 852902137 [LONG],
              local_id: 144203 [INT],
              secret: 2124116499919855875 [LONG],
            },
            photo_big: { fileLocation
              dc_id: 5 [INT],
              volume_id: 852902137 [LONG],
              local_id: 144205 [INT],
              secret: 2433093284049731127 [LONG],
            },
          },
          status: { userStatusOffline
            was_online: 1532831934 [INT],
          },
          bot_info_version: [ SKIPPED BY BIT 14 IN FIELD flags ],
          restriction_reason: [ SKIPPED BY BIT 18 IN FIELD flags ],
          bot_inline_placeholder: [ SKIPPED BY BIT 19 IN FIELD flags ],
          lang_code: [ SKIPPED BY BIT 22 IN FIELD flags ],
        },
        { user
          flags: 6267 [INT],
          self: [ SKIPPED BY BIT 10 IN FIELD flags ],
          contact: YES [ BY BIT 11 IN FIELD flags ],
          mutual_contact: YES [ BY BIT 12 IN FIELD flags ],
          deleted: [ SKIPPED BY BIT 13 IN FIELD flags ],
          bot: [ SKIPPED BY BIT 14 IN FIELD flags ],
          bot_chat_history: [ SKIPPED BY BIT 15 IN FIELD flags ],
          bot_nochats: [ SKIPPED BY BIT 16 IN FIELD flags ],
          verified: [ SKIPPED BY BIT 17 IN FIELD flags ],
          restricted: [ SKIPPED BY BIT 18 IN FIELD flags ],
          min: [ SKIPPED BY BIT 20 IN FIELD flags ],
          bot_inline_geo: [ SKIPPED BY BIT 21 IN FIELD flags ],
          id: 264696845 [INT],
          access_hash: 13587416540126710881 [LONG],
          first_name: "本起" [STRING],
          last_name: [ SKIPPED BY BIT 2 IN FIELD flags ],
          username: "benqi" [STRING],
          phone: "8613606512716" [STRING],
          photo: { userProfilePhoto
            photo_id: 1136864293085620150 [LONG],
            photo_small: { fileLocation
              dc_id: 5 [INT],
              volume_id: 852832478 [LONG],
              local_id: 60529 [INT],
              secret: 9338736479069608191 [LONG],
            },
            photo_big: { fileLocation
              dc_id: 5 [INT],
              volume_id: 852832478 [LONG],
              local_id: 60531 [INT],
              secret: 1214473206298316587 [LONG],
            },
          },
          status: { userStatusOnline
            expires: 1532832234 [INT],
          },
          bot_info_version: [ SKIPPED BY BIT 14 IN FIELD flags ],
          restriction_reason: [ SKIPPED BY BIT 18 IN FIELD flags ],
          bot_inline_placeholder: [ SKIPPED BY BIT 19 IN FIELD flags ],
          lang_code: [ SKIPPED BY BIT 22 IN FIELD flags ],
        },
      ],
    },

- 3.
	body: { messages_getHistory
	  peer: { inputPeerChannel
		channel_id: 1179206386 [INT],
		access_hash: 2629819016532760918 [LONG],
	  },
	  offset_id: 0 [INT],
	  offset_date: 0 [INT],
	  add_offset: 0 [INT],
	  limit: 1 [INT],
	  max_id: 0 [INT],
	  min_id: 0 [INT],
	},


  body: { rpc_result
    req_msg_id: 6583463147519930592 [LONG],
    result: [GZIPPED] { channels_channelParticipant
      participant: { channelParticipantSelf
        user_id: 448603711 [INT],
        inviter_id: 264696845 [INT],
        date: 1532831960 [INT],
      },
      users: [ vector<0x0>
        { user
          flags: 3195 [INT],
          self: YES [ BY BIT 10 IN FIELD flags ],
          contact: YES [ BY BIT 11 IN FIELD flags ],
          mutual_contact: [ SKIPPED BY BIT 12 IN FIELD flags ],
          deleted: [ SKIPPED BY BIT 13 IN FIELD flags ],
          bot: [ SKIPPED BY BIT 14 IN FIELD flags ],
          bot_chat_history: [ SKIPPED BY BIT 15 IN FIELD flags ],
          bot_nochats: [ SKIPPED BY BIT 16 IN FIELD flags ],
          verified: [ SKIPPED BY BIT 17 IN FIELD flags ],
          restricted: [ SKIPPED BY BIT 18 IN FIELD flags ],
          min: [ SKIPPED BY BIT 20 IN FIELD flags ],
          bot_inline_geo: [ SKIPPED BY BIT 21 IN FIELD flags ],
          id: 448603711 [INT],
          access_hash: 16248249830026626652 [LONG],
          first_name: "璧君😊" [STRING],
          last_name: [ SKIPPED BY BIT 2 IN FIELD flags ],
          username: "yubijun2" [STRING],
          phone: "8613588021430" [STRING],
          photo: { userProfilePhoto
            photo_id: 1926738268065474478 [LONG],
            photo_small: { fileLocation
              dc_id: 5 [INT],
              volume_id: 852902137 [LONG],
              local_id: 144203 [INT],
              secret: 2124116499919855875 [LONG],
            },
            photo_big: { fileLocation
              dc_id: 5 [INT],
              volume_id: 852902137 [LONG],
              local_id: 144205 [INT],
              secret: 2433093284049731127 [LONG],
            },
          },
          status: { userStatusOffline
            was_online: 1532831934 [INT],
          },
          bot_info_version: [ SKIPPED BY BIT 14 IN FIELD flags ],
          restriction_reason: [ SKIPPED BY BIT 18 IN FIELD flags ],
          bot_inline_placeholder: [ SKIPPED BY BIT 19 IN FIELD flags ],
          lang_code: [ SKIPPED BY BIT 22 IN FIELD flags ],
        },
        { user
          flags: 6267 [INT],
          self: [ SKIPPED BY BIT 10 IN FIELD flags ],
          contact: YES [ BY BIT 11 IN FIELD flags ],
          mutual_contact: YES [ BY BIT 12 IN FIELD flags ],
          deleted: [ SKIPPED BY BIT 13 IN FIELD flags ],
          bot: [ SKIPPED BY BIT 14 IN FIELD flags ],
          bot_chat_history: [ SKIPPED BY BIT 15 IN FIELD flags ],
          bot_nochats: [ SKIPPED BY BIT 16 IN FIELD flags ],
          verified: [ SKIPPED BY BIT 17 IN FIELD flags ],
          restricted: [ SKIPPED BY BIT 18 IN FIELD flags ],
          min: [ SKIPPED BY BIT 20 IN FIELD flags ],
          bot_inline_geo: [ SKIPPED BY BIT 21 IN FIELD flags ],
          id: 264696845 [INT],
          access_hash: 13587416540126710881 [LONG],
          first_name: "本起" [STRING],
          last_name: [ SKIPPED BY BIT 2 IN FIELD flags ],
          username: "benqi" [STRING],
          phone: "8613606512716" [STRING],
          photo: { userProfilePhoto
            photo_id: 1136864293085620150 [LONG],
            photo_small: { fileLocation
              dc_id: 5 [INT],
              volume_id: 852832478 [LONG],
              local_id: 60529 [INT],
              secret: 9338736479069608191 [LONG],
            },
            photo_big: { fileLocation
              dc_id: 5 [INT],
              volume_id: 852832478 [LONG],
              local_id: 60531 [INT],
              secret: 1214473206298316587 [LONG],
            },
          },
          status: { userStatusOnline
            expires: 1532832234 [INT],
          },
          bot_info_version: [ SKIPPED BY BIT 14 IN FIELD flags ],
          restriction_reason: [ SKIPPED BY BIT 18 IN FIELD flags ],
          bot_inline_placeholder: [ SKIPPED BY BIT 19 IN FIELD flags ],
          lang_code: [ SKIPPED BY BIT 22 IN FIELD flags ],
        },
      ],
    },
 */

// messages.migrateChat#15a3b8e3 chat_id:int = Updates;
func (s *MessagesServiceImpl) MessagesMigrateChat(ctx context.Context, request *mtproto.TLMessagesMigrateChat) (*mtproto.Updates, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("MessagesMigrateChat - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	// TODO(@benqi): Impl MessagesMigrateChat logic

	return nil, fmt.Errorf("Not impl MessagesMigrateChat")
}
