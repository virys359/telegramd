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
	"fmt"
	"github.com/golang/glog"
	"github.com/nebulaim/telegramd/baselib/grpc_util"
	"github.com/nebulaim/telegramd/baselib/logger"
	"github.com/nebulaim/telegramd/proto/mtproto"
	"github.com/nebulaim/telegramd/server/nbfs/biz/core/document"
	photo2 "github.com/nebulaim/telegramd/server/nbfs/biz/core/photo"
	"golang.org/x/net/context"
)

// upload.getFile#e3a6cfb5 location:InputFileLocation offset:int limit:int = upload.File;
func (s *UploadServiceImpl) UploadGetFile(ctx context.Context, request *mtproto.TLUploadGetFile) (*mtproto.Upload_File, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("upload.getFile#e3a6cfb5 - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))
	var (
		uploadFile *mtproto.Upload_File
		err        error
	)
	switch request.GetLocation().GetConstructor() {
	case mtproto.TLConstructor_CRC32_inputFileLocation:
		fileLocation := request.GetLocation().To_InputFileLocation()
		uploadFile, err = photo2.GetPhotoFileData(fileLocation.GetVolumeId(),
			fileLocation.GetLocalId(),
			fileLocation.GetSecret(),
			request.GetOffset(),
			request.GetLimit())
	case mtproto.TLConstructor_CRC32_inputEncryptedFileLocation:
	case mtproto.TLConstructor_CRC32_inputDocumentFileLocation:
		fileLocation := request.GetLocation().To_InputDocumentFileLocation()
		uploadFile, err = document.GetDocumentFileData(fileLocation.GetId(),
			fileLocation.GetAccessHash(),
			fileLocation.GetVersion(),
			request.GetOffset(),
			request.GetLimit())
	case mtproto.TLConstructor_CRC32_inputDocumentFileLocationLayer11:
		fileLocation := request.GetLocation().To_InputDocumentFileLocation()
		uploadFile, err = document.GetDocumentFileData(fileLocation.GetId(),
			fileLocation.GetAccessHash(),
			fileLocation.GetVersion(),
			request.GetOffset(),
			request.GetLimit())
	default:
		err = fmt.Errorf("invalid InputFileLocation type: %d", request.GetLocation().GetConstructor())
	}

	if err != nil {
		glog.Error(err)
		return nil, err
	}

	// type:storage.FileType mtime:int bytes:bytes
	glog.Infof("upload.getFile#e3a6cfb5 - reply: {type: %v, mime: %d, len_bytes: %d}",
		uploadFile.GetData2().GetType(),
		uploadFile.GetData2().GetMtime(),
		len(uploadFile.GetData2().GetBytes()))

	return uploadFile, err
}
