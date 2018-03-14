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
	"time"
	"github.com/nebulaim/telegramd/biz_model/model"
	base2 "github.com/nebulaim/telegramd/baselib/base"
)

// updates.getDifference#25939651 flags:# pts:int pts_total_limit:flags.0?int date:int qts:int = updates.Difference;
func (s *UpdatesServiceImpl) UpdatesGetDifference(ctx context.Context, request *mtproto.TLUpdatesGetDifference) (*mtproto.Updates_Difference, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("UpdatesGetDifference - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	otherUpdates, boxIDList, lastPts := model.GetUpdatesModel().GetUpdatesByGtPts(md.UserId, request.GetPts())
	messages := model.GetMessageModel().GetMessagesByPeerAndMessageIdList2(md.UserId, boxIDList)
	userIdList, _, _ := model.PickAllIDListByMessages(messages)
	userList := model.GetUserModel().GetUsersBySelfAndIDList(md.UserId, userIdList)

	state := &mtproto.TLUpdatesState{Data2: &mtproto.Updates_State_Data{
		Pts:         lastPts,
		Date:        int32(time.Now().Unix()),
		UnreadCount: 0,
		Seq:         int32(model.GetSequenceModel().CurrentSeqId(base2.Int32ToString(md.UserId))),
	}}
	difference := &mtproto.TLUpdatesDifference{Data2: &mtproto.Updates_Difference_Data{
		NewMessages:  messages,
		OtherUpdates: otherUpdates,
		Users:        userList,
		// Chats:        nil,
		State:        state.To_Updates_State(),
	}}

	glog.Infof("UpdatesGetDifference - reply: %s", difference)
	return difference.To_Updates_Difference(), nil
}
