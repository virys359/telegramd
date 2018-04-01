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
 Request:
	body: { contacts_deleteContact
	  id: { inputUser
		user_id: 448603711 [INT],
		access_hash: 13014512641679536571 [LONG],
	  },
	},

 RpcResult:
    result: { contacts_link
      my_link: { contactLinkHasPhone },
      foreign_link: { contactLinkHasPhone },
      user: { user
        flags: 123 [INT],
        self: [ SKIPPED BY BIT 10 IN FIELD flags ],
        contact: [ SKIPPED BY BIT 11 IN FIELD flags ],
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
        access_hash: 13014512641679536571 [LONG],
        first_name: "xxxx" [STRING],
        last_name: [ SKIPPED BY BIT 2 IN FIELD flags ],
        username: "xxxx" [STRING],
        phone: "xxxx" [STRING],
        photo: { userProfilePhoto
          photo_id: 1926738268065474473 [LONG],
          photo_small: { fileLocation
            dc_id: 5 [INT],
            volume_id: 852524509 [LONG],
            local_id: 178719 [INT],
            secret: 17693880416897175419 [LONG],
          },
          photo_big: { fileLocation
            dc_id: 5 [INT],
            volume_id: 852524509 [LONG],
            local_id: 178721 [INT],
            secret: 112291147053671954 [LONG],
          },
        },
        status: { userStatusOffline
          was_online: 1522415924 [INT],
        },
        bot_info_version: [ SKIPPED BY BIT 14 IN FIELD flags ],
        restriction_reason: [ SKIPPED BY BIT 18 IN FIELD flags ],
        bot_inline_placeholder: [ SKIPPED BY BIT 19 IN FIELD flags ],
        lang_code: [ SKIPPED BY BIT 22 IN FIELD flags ],
      },
    },

 Updates:
  body: { updates
    updates: [ vector<0x0>
      { updateContactLink
        user_id: 264696845 [INT],
        my_link: { contactLinkContact },
        foreign_link: { contactLinkHasPhone },
      },
    ],
    users: [ vector<0x0>
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
        first_name: "xxxx" [STRING],
        last_name: [ SKIPPED BY BIT 2 IN FIELD flags ],
        username: "xxxx" [STRING],
        phone: "xxxx" [STRING],
        photo: { userProfilePhoto
          photo_id: 1136864293085620140 [LONG],
          photo_small: { fileLocation
            dc_id: 5 [INT],
            volume_id: 852715620 [LONG],
            local_id: 217304 [INT],
            secret: 7641812730259272198 [LONG],
          },
          photo_big: { fileLocation
            dc_id: 5 [INT],
            volume_id: 852715620 [LONG],
            local_id: 217306 [INT],
            secret: 17483086520490969958 [LONG],
          },
        },
        status: { userStatusOnline
          expires: 1522416235 [INT],
        },
        bot_info_version: [ SKIPPED BY BIT 14 IN FIELD flags ],
        restriction_reason: [ SKIPPED BY BIT 18 IN FIELD flags ],
        bot_inline_placeholder: [ SKIPPED BY BIT 19 IN FIELD flags ],
        lang_code: [ SKIPPED BY BIT 22 IN FIELD flags ],
      },
    ],
    chats: [ vector<0x0> ],
    date: 1522415944 [INT],
    seq: 4058 [INT],
  },
 */

// contacts.deleteContact#8e953744 id:InputUser = contacts.Link;
func (s *ContactsServiceImpl) ContactsDeleteContact(ctx context.Context, request *mtproto.TLContactsDeleteContact) (*mtproto.Contacts_Link, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("ContactsDeleteContact - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	// TODO(@benqi): Impl ContactsDeleteContact logic

	return nil, fmt.Errorf("Not impl ContactsDeleteContact")
}
