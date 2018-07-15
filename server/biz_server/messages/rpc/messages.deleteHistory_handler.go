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
	"github.com/nebulaim/telegramd/biz/core/message"
	"github.com/nebulaim/telegramd/biz/base"
	"github.com/nebulaim/telegramd/server/sync/sync_client"
)

// messages.deleteHistory#1c015b09 flags:# just_clear:flags.0?true peer:InputPeer max_id:int = messages.AffectedHistory;
func (s *MessagesServiceImpl) MessagesDeleteHistory(ctx context.Context, request *mtproto.TLMessagesDeleteHistory) (*mtproto.Messages_AffectedHistory, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("messages.deleteHistory#1c015b09 - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	// TODO(@benqi): Impl MessagesDeleteHistory logic

	peer := base.FromInputPeer2(md.UserId, request.GetPeer())
	if peer.PeerType == base.PEER_SELF {
		peer.PeerType = base.PEER_USER
	}
	boxIdList := message.GetMessageIdListByDialog(md.UserId, peer)
	if len(boxIdList) > 0 {
		// TOOD(@benqi): delete dialog message.
		message.DeleteByMessageIdList(md.UserId, boxIdList)

		updateDeleteMessages := mtproto.NewTLUpdateDeleteMessages()
		updateDeleteMessages.SetMessages(boxIdList)
		// updateDeleteMessages.SetMaxId(request.MaxId)
		_, err := sync_client.GetSyncClient().SyncOneUpdateData3(md.ServerId, md.AuthId, md.SessionId, md.UserId, md.ClientMsgId, updateDeleteMessages.To_Update())
		if err != nil {
			glog.Error(err)
			return nil, err
		}

		err = mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_NOTRETURN_CLIENT)
		glog.Infof("messages.deleteHistory#1c015b09 - reply: {%v}", err)
		return nil, err
	} else {
		affectedHistory := &mtproto.TLMessagesAffectedHistory{Data2: &mtproto.Messages_AffectedHistory_Data{
			Pts:      0,
			PtsCount: 0,
			Offset:   0,
		}}
		glog.Infof("messages.deleteHistory#1c015b09 - reply: {%v}", affectedHistory)
		return affectedHistory.To_Messages_AffectedHistory(), nil
	}
}
