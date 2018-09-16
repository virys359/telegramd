/*
 *  Copyright (c) 2018, https://github.com/nebulaim
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
    "github.com/nebulaim/telegramd/proto/mtproto"
    "golang.org/x/net/context"
    "github.com/nebulaim/telegramd/baselib/logger"
)

// sync.pushChannelUpdates#bfd3d677 channel_id:int exclude_user_id:int updates:Updates = Bool;
func (s *SyncServiceImpl) SyncPushChannelUpdates(ctx context.Context, request *mtproto.TLSyncPushChannelUpdates) (*mtproto.Bool, error) {
    glog.Infof("sync.pushChannelUpdates#bfd3d677 - request: {%s}", logger.JsonDebugData(request))
    //err := s.processUpdatesRequest(request.GetUserId(), request.GetUpdates())
    //if err == nil {
    //	s.pushUpdatesToSession(syncTypeUser, request.GetUserId(), 0, 0, request.GetUpdates())
    //	glog.Infof("SyncPushUpdates - reply: %s", logger.JsonDebugData(request))
    //} else {
    //	glog.Error(err)
    //	return mtproto.ToBool(false), nil
    //}
    return mtproto.ToBool(true), nil
}
