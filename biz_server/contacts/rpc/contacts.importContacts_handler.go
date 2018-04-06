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
	"github.com/nebulaim/telegramd/biz/base"
	contact2 "github.com/nebulaim/telegramd/biz/core/contact"
	"github.com/nebulaim/telegramd/biz/core/user"
	"github.com/nebulaim/telegramd/biz_server/sync_client"
	updates2 "github.com/nebulaim/telegramd/biz/core/update"
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
	glog.Infof("contacts.importContacts#2c800be5 - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	var (
		err error
		// importedContacts *mtproto.TLImportedContact
	)

	if len(request.Contacts) == 0 {
		err := mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_BAD_REQUEST)
		glog.Error(err, ": contacts empty")
		return nil, err
	}

	// 注意: 目前只支持导入1个并且已经注册的联系人!!!!
	inputContact := request.Contacts[0].To_InputPhoneContact()

	if inputContact.GetFirstName() == "" {
		err := mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_FIRSTNAME_INVALID)
		glog.Error(err, ": first_name is empty")
		return nil, err
	}

	phone, err := base.CheckAndGetPhoneNumber(inputContact.GetPhone())
	if err != nil {
		err := mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_PHONE_CODE_INVALID)
		glog.Error(err, ": phone code invalid - ", inputContact.GetPhone())
		return nil, err
	}

	contactUser := user.GetUserByPhoneNumber(false, phone)
	if contactUser == nil {
		// 这里该手机号未注册，我们认为手机号出错
		err := mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_PHONE_CODE_INVALID)
		glog.Error(err, ": phone code invalid - ", inputContact.GetPhone())
		return nil, err
	}
	contactUser.SetContact(true)
	// contactUser.SetMutualContact(true)
	contactLogic := contact2.NewContactLogic(md.UserId)
	needUpdate := contactLogic.ImportContact(contactUser.GetId(), phone, inputContact.GetFirstName(), inputContact.GetLastName())
	_ = needUpdate

	updates := updates2.NewUpdatesLogic(md.UserId)
	contactLink := &mtproto.TLUpdateContactLink{Data2: &mtproto.Update_Data{
		UserId:      md.UserId,
		MyLink:      mtproto.NewTLContactLinkContact().To_ContactLink(),
		ForeignLink: mtproto.NewTLContactLinkContact().To_ContactLink(),
	}}
	updates.AddUpdate(contactLink.To_Update())
	updates.AddUser(contactUser.To_User())
	// TODO(@benqi): sync update
	sync_client.GetSyncClient().PushToUserUpdatesData(md.UserId, updates.ToUpdates())

	////////////////////////////////////////////////////////////////////////////////////////
	imported := &mtproto.TLImportedContact{Data2: &mtproto.ImportedContact_Data{
		UserId: contactUser.GetId(),
		ClientId: inputContact.GetClientId(),
	}}
	// contacts.importedContacts#77d01c3b imported:Vector<ImportedContact> popular_invites:Vector<PopularContact> retry_contacts:Vector<long> users:Vector<User> = contacts.ImportedContacts;
	importedContacts := &mtproto.TLContactsImportedContacts{Data2: &mtproto.Contacts_ImportedContacts_Data{
		Imported: []*mtproto.ImportedContact{imported.To_ImportedContact()},
		Users: []*mtproto.User{contactUser.To_User()},
	}}

	glog.Infof("contacts.importContacts#2c800be5 - reply: %s", logger.JsonDebugData(importedContacts))
	return importedContacts.To_Contacts_ImportedContacts(), nil
}
