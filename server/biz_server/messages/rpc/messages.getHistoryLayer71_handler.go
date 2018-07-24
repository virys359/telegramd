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
	"github.com/nebulaim/telegramd/proto/mtproto"
	"golang.org/x/net/context"
)

// request: {"peer":{"constructor":2072935910,"data2":{"user_id":4,"access_hash":405858233924775823}},"offset_id":2147483647,"offset_date":2147483647,"limit":1,"max_id":2147483647,"min_id":1}
// request: {"peer":{"constructor":2072935910,"data2":{"user_id":4,"access_hash":405858233924775823}},"offset_id":2147483647,"offset_date":2147483647,"limit":1,"max_id":2147483647,"min_id":1}
// messages.getHistory#afa92846 peer:InputPeer offset_id:int offset_date:int add_offset:int limit:int max_id:int min_id:int = messages.Messages;
func (s *MessagesServiceImpl) MessagesGetHistoryLayer71(ctx context.Context, request *mtproto.TLMessagesGetHistoryLayer71) (*mtproto.Messages_Messages, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("messages.getHistoryLayer71#afa92846 - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	requestLayer71 := &mtproto.TLMessagesGetHistory{
		Peer:       request.Peer,
		OffsetId:   request.OffsetId,
		OffsetDate: request.OffsetDate,
		AddOffset:  request.AddOffset,
		Limit:      request.Limit,
		MaxId:      request.MaxId,
		MinId:      request.MinId,
		Hash:       0,
	}

	messagesMessages := s.getHistoryMessages(md, requestLayer71)

	glog.Infof("messages.getHistoryLayer71#afa92846 - reply: %s", logger.JsonDebugData(messagesMessages))
	return messagesMessages, nil
}
