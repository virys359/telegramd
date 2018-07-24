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
	"github.com/nebulaim/telegramd/biz/dal/dao"
	"github.com/nebulaim/telegramd/proto/mtproto"
	"golang.org/x/net/context"
)

// contacts.resolveUsername#f93ccba3 username:string = contacts.ResolvedPeer;
func (s *ContactsServiceImpl) ContactsResolveUsername(ctx context.Context, request *mtproto.TLContactsResolveUsername) (*mtproto.Contacts_ResolvedPeer, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("contacts.resolveUsername#f93ccba3 - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	// TODO(@benqi): Impl ContactsResolveUsername logic
	do := dao.GetUsersDAO(dao.DB_SLAVE).SelectByUsername(request.GetUsername())
	if do == nil {
		err := mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_USERNAME_INVALID)
		glog.Error(err)
		return nil, err
	}

	peer := &mtproto.TLPeerUser{Data2: &mtproto.Peer_Data{
		UserId: do.Id,
	}}
	resolvedPeer := &mtproto.TLContactsResolvedPeer{Data2: &mtproto.Contacts_ResolvedPeer_Data{
		Peer:  peer.To_Peer(),
		Chats: []*mtproto.Chat{},
		Users: []*mtproto.User{s.UserModel.GetUserById(md.UserId, do.Id).To_User()},
	}}

	glog.Infof("contacts.resolveUsername#f93ccba3 - reply: {%v}", resolvedPeer)
	return resolvedPeer.To_Contacts_ResolvedPeer(), nil
}
