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
	"github.com/nebulaim/telegramd/grpc_util"
	"github.com/nebulaim/telegramd/mtproto"
	"golang.org/x/net/context"
)


/*
 request:
	body: { messages_toggleChatAdmins
	  chat_id: 259171813 [INT],
	  enabled: { boolTrue },
	},

 result:
	body: { rpc_result
	  req_msg_id: 6537655324771363928 [LONG],
	  result: { updates
		updates: [ vector<0x0> ],
		users: [ vector<0x0> ],
		chats: [ vector<0x0>
		  { chat
			flags: 9 [INT],
			creator: YES [ BY BIT 0 IN FIELD flags ],
			kicked: [ SKIPPED BY BIT 1 IN FIELD flags ],
			left: [ SKIPPED BY BIT 2 IN FIELD flags ],
			admins_enabled: YES [ BY BIT 3 IN FIELD flags ],
			admin: [ SKIPPED BY BIT 4 IN FIELD flags ],
			deactivated: [ SKIPPED BY BIT 5 IN FIELD flags ],
			id: 259171813 [INT],
			title: "2134" [STRING],
			photo: { chatPhoto
			  photo_small: { fileLocation
				dc_id: 5 [INT],
				volume_id: 852735964 [LONG],
				local_id: 123530 [INT],
				secret: 7026333565383518945 [LONG],
			  },
			  photo_big: { fileLocation
				dc_id: 5 [INT],
				volume_id: 852735964 [LONG],
				local_id: 123532 [INT],
				secret: 13296246115511276312 [LONG],
			  },
			},
			participants_count: 2 [INT],
			date: 1522165981 [INT],
			version: 2 [INT],
			migrated_to: [ SKIPPED BY BIT 6 IN FIELD flags ],
		  },
		],
		date: 1522166495 [INT],
		seq: 0 [INT],
	  },
	},

updates:
	body: { updateShort
	  update: { updateChatAdmins
		chat_id: 259171813 [INT],
		enabled: { boolTrue },
		version: 2 [INT],
	  },
	  date: 1522166495 [INT],
	},
 */
// messages.toggleChatAdmins#ec8bd9e1 chat_id:int enabled:Bool = Updates;
func (s *MessagesServiceImpl) MessagesToggleChatAdmins(ctx context.Context, request *mtproto.TLMessagesToggleChatAdmins) (*mtproto.Updates, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("MessagesToggleChatAdmins - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	// TODO(@benqi): Impl MessagesToggleChatAdmins logic

	return nil, fmt.Errorf("Not impl MessagesToggleChatAdmins")
}
