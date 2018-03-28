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
	body: { contacts_importContacts
	  contacts: [ vector<0x0>
		{ inputPhoneContact
		  client_id: 4368649977211762445 [LONG],
		  phone: "+86 158 3819 6209" [STRING],
		  first_name: "X" [STRING],
		  last_name: "Ray" [STRING],
		},
	  ],
	},

 result:
	body: { rpc_result
	  req_msg_id: 6537537306041558152 [LONG],
	  result: { contacts_importedContacts
		imported: [ vector<0x0>
		  { importedContact
			user_id: 456246609 [INT],
			client_id: 4368649977211762445 [LONG],
		  },
		],
		popular_invites: [ vector<0x0> ],
		retry_contacts: [ vector<0x22076cba> ],
		users: [ vector<0x0>
		  { user
			flags: 2135 [INT],
			self: [ SKIPPED BY BIT 10 IN FIELD flags ],
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
			id: 456246609 [INT],
			access_hash: 9056436018490939711 [LONG],
			first_name: "X" [STRING],
			last_name: "Ray" [STRING],
			username: [ SKIPPED BY BIT 3 IN FIELD flags ],
			phone: "8615838196209" [STRING],
			photo: [ SKIPPED BY BIT 5 IN FIELD flags ],
			status: { userStatusRecently },
			bot_info_version: [ SKIPPED BY BIT 14 IN FIELD flags ],
			restriction_reason: [ SKIPPED BY BIT 18 IN FIELD flags ],
			bot_inline_placeholder: [ SKIPPED BY BIT 19 IN FIELD flags ],
			lang_code: [ SKIPPED BY BIT 22 IN FIELD flags ],
		  },
		],
	  },
	},

 updates:
	body: { updates
	  updates: [ vector<0x0>
		{ updateContactLink
		  user_id: 456246609 [INT],
		  my_link: { contactLinkContact },
		  foreign_link: { contactLinkUnknown },
		},
	  ],
	  users: [ vector<0x0>
		{ user
		  flags: 2135 [INT],
		  self: [ SKIPPED BY BIT 10 IN FIELD flags ],
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
		  id: 456246609 [INT],
		  access_hash: 9056436018490939711 [LONG],
		  first_name: "X" [STRING],
		  last_name: "Ray" [STRING],
		  username: [ SKIPPED BY BIT 3 IN FIELD flags ],
		  phone: "8615838196209" [STRING],
		  photo: [ SKIPPED BY BIT 5 IN FIELD flags ],
		  status: { userStatusRecently },
		  bot_info_version: [ SKIPPED BY BIT 14 IN FIELD flags ],
		  restriction_reason: [ SKIPPED BY BIT 18 IN FIELD flags ],
		  bot_inline_placeholder: [ SKIPPED BY BIT 19 IN FIELD flags ],
		  lang_code: [ SKIPPED BY BIT 22 IN FIELD flags ],
		},
	  ],
	  chats: [ vector<0x0> ],
	  date: 1522139017 [INT],
	  seq: 3631 [INT],
	},
  },
 */

// contacts.importContacts#2c800be5 contacts:Vector<InputContact> = contacts.ImportedContacts;
func (s *ContactsServiceImpl) ContactsImportContacts(ctx context.Context, request *mtproto.TLContactsImportContacts) (*mtproto.Contacts_ImportedContacts, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("ContactsImportContacts - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	// TODO(@benqi): Impl ContactsImportContacts logic

	return nil, fmt.Errorf("Not impl ContactsImportContacts")
}
