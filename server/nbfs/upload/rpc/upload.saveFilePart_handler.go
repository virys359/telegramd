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
	"github.com/nebulaim/telegramd/baselib/grpc_util"
	"github.com/nebulaim/telegramd/baselib/logger"
	"github.com/nebulaim/telegramd/proto/mtproto"
	"github.com/nebulaim/telegramd/server/nbfs/biz/core/file"
	"golang.org/x/net/context"
)

// upload.saveFilePart#b304a621 file_id:long file_part:int bytes:bytes = Bool;
func (s *UploadServiceImpl) UploadSaveFilePart(ctx context.Context, request *mtproto.TLUploadSaveFilePart) (*mtproto.Bool, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("upload.saveFilePart#b304a621 - metadata: %s, request: {file_id: %d, file_part: %d, bytes_len: %d}",
		logger.JsonDebugData(md),
		request.FileId,
		request.FilePart,
		len(request.Bytes))

	// TODO(@benqi): Check file_part <= bigFileSize/kMaxFileSize, Check len(bytes) <= kMaxFilePartSize

	filePartLogic, err := file.MakeFilePartData(md.AuthId, request.FileId, request.FilePart == 0, false)
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	err = filePartLogic.SaveFilePart(request.FilePart, request.Bytes)
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	glog.Infof("upload.saveFilePart#b304a621 - reply: {true}")
	return mtproto.ToBool(true), nil
}
