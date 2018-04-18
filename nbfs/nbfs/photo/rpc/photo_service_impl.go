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
	"context"
	"github.com/nebulaim/telegramd/mtproto"
	"github.com/golang/glog"
	"github.com/nebulaim/telegramd/baselib/logger"
	"fmt"
	"github.com/nebulaim/telegramd/nbfs/biz/core/photo"
	"github.com/nebulaim/telegramd/nbfs/biz/core/file"
	"github.com/nebulaim/telegramd/nbfs/biz"
	"github.com/nebulaim/telegramd/biz/base"
)

type PhotoServiceImpl struct {
}

// rpc
// rpc nbfs_uploadPhotoFile(UploadPhotoFileRequest) returns (PhotoDataRsp);
func (s *PhotoServiceImpl) NbfsUploadPhotoFile(ctx context.Context, request *mtproto.UploadPhotoFileRequest) (*mtproto.PhotoDataRsp, error) {
	glog.Infof("nbfs.uploadPhotoFile - request: %s", logger.JsonDebugData(request))
	if request.GetFile() == nil {
		return nil, fmt.Errorf("bad request")
	}

	var reply *mtproto.PhotoDataRsp

	inputFile := request.GetFile().GetData2()
	switch request.GetFile().GetConstructor() {
	case mtproto.TLConstructor_CRC32_inputFile:
		// TODO(@benqi): 出错以后，回滚数据库操作
		filePart, err := file.MakeFilePartData(request.OwnerId, inputFile.Id, false, false)
		if err != nil {
			glog.Error(err)
			return nil, err
		}

		var filename = core.NBFS_DATA_PATH + filePart.FilePath
		md5Checksum, _ := core.CalcMd5File(filename)
		if md5Checksum != inputFile.Md5Checksum {
			return nil, fmt.Errorf("check md5 error")
		}

		fileData, err := file.NewFileData(filePart.FilePartId,
			filePart.FilePath,
			inputFile.Name,
			filePart.FileSize,
			md5Checksum)
		if err != nil {
			glog.Error(err)
			return nil, err
		}

		photoId := base.NextSnowflakeId()
		szList, err := photo.UploadPhotoFile(photoId, fileData.FilePath, fileData.Ext, false)
		if err != nil {
			glog.Error(err)
			return nil, err
		}

		reply = &mtproto.PhotoDataRsp{
			PhotoId:  photoId,
			SizeList: szList,
		}
	case mtproto.TLConstructor_CRC32_inputFileBig:
	default:
		return nil, fmt.Errorf("bad request")
	}

	glog.Infof("nbfs.uploadPhotoFile - reply: %s", logger.JsonDebugData(reply))
	return reply, nil
}

// rpc nbfs_getPhotoFileData(GetPhotoFileDataRequest) returns (PhotoDataRsp);
func (s *PhotoServiceImpl) NbfsGetPhotoFileData(ctx context.Context, request *mtproto.GetPhotoFileDataRequest) (*mtproto.PhotoDataRsp, error) {
	glog.Infof("nbfs.getPhotoFileData - request: %s", logger.JsonDebugData(request))

	var photoId = request.GetPhotoId()
	szList := photo.GetPhotoSizeList(photoId)
	reply := &mtproto.PhotoDataRsp{
		PhotoId:  photoId,
		SizeList: szList,
	}

	glog.Infof("nbfs.getPhotoFileData - reply: %s", logger.JsonDebugData(reply))
	return reply, nil
}
