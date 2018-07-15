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
	"context"
	"time"
	"math/rand"
	"github.com/nebulaim/telegramd/proto/mtproto"
	"github.com/golang/glog"
	"github.com/nebulaim/telegramd/baselib/logger"
	"github.com/nebulaim/telegramd/server/nbfs/biz/core/photo"
	"github.com/nebulaim/telegramd/server/nbfs/biz/core/file"
	"github.com/nebulaim/telegramd/server/nbfs/biz/base"
	document2 "github.com/nebulaim/telegramd/server/nbfs/biz/core/document"
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

	inputFile := request.GetFile().GetData2()

	var (
		reply *mtproto.PhotoDataRsp
		// isBigFile = request.GetFile().GetConstructor() == mtproto.TLConstructor_CRC32_inputFileBig
	)

	//// TODO(@benqi): 出错以后，回滚数据库操作
	//filePart, err := file.MakeFilePartData(request.OwnerId, inputFile.Id, false, isBigFile)
	//if err != nil {
	//	glog.Error(err)
	//	return nil, err
	//}
	//
	//var filename = core.NBFS_DATA_PATH + filePart.FilePath
	//md5Checksum, _ := core.CalcMd5File(filename)
	//if md5Checksum != inputFile.Md5Checksum {
	//	return nil, fmt.Errorf("check md5 error")
	//}

	filePart, err := file.DoSavedFilePart(request.OwnerId, inputFile.Id)
	if filePart.SavedMd5Hash != inputFile.Md5Checksum {
		return nil, fmt.Errorf("check md5 error: %s, %s", filePart.SavedMd5Hash, inputFile.Md5Checksum)
	}

	fileData, err := file.NewFileData(filePart.FilePartId,
		filePart.SavedFilePath,
		inputFile.Name,
		filePart.FileSize,
		filePart.SavedMd5Hash)
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	photoId := base.NextSnowflakeId()
	accessHash := rand.Int63()
	szList, err := photo.UploadPhotoFile(photoId, accessHash, fileData.FilePath, fileData.Ext, false)
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	reply = &mtproto.PhotoDataRsp{
		PhotoId:    photoId,
		AccessHash: accessHash,
		Date:       int32(time.Now().Unix()),
		SizeList:   szList,
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

// inputMediaUploadedPhoto#2f37e231 flags:# file:InputFile caption:string stickers:flags.0?Vector<InputDocument> ttl_seconds:flags.1?int = InputMedia;
func (s *PhotoServiceImpl) NbfsUploadedPhotoMedia(ctx context.Context, request *mtproto.NbfsUploadedPhotoMedia) (*mtproto.TLMessageMediaPhoto, error) {
	glog.Infof("nbfs.uploadedPhotoMedia - request: %s", logger.JsonDebugData(request))

	inputFile := request.GetMedia().GetFile().GetData2()

	var (
		// reply *mtproto.PhotoDataRsp
		// isBigFile = request.GetMedia().GetFile().GetConstructor() == mtproto.TLConstructor_CRC32_inputFileBig
	)

	// TODO(@benqi): 出错以后，回滚数据库操作
	//filePart, err := file.MakeFilePartData(request.OwnerId, inputFile.Id, false, isBigFile)
	//if err != nil {
	//	glog.Error(err)
	//	return nil, err
	//}
	//
	//var filename = core.NBFS_DATA_PATH + filePart.FilePath
	//md5Checksum, _ := core.CalcMd5File(filename)
	//if md5Checksum != inputFile.Md5Checksum {
	//	return nil, fmt.Errorf("check md5 error")
	//}

	filePart, err := file.DoSavedFilePart(request.OwnerId, inputFile.Id)
	if filePart.SavedMd5Hash != inputFile.Md5Checksum {
		return nil, fmt.Errorf("check md5 error: %s, %s", filePart.SavedMd5Hash, inputFile.Md5Checksum)
	}

	fileData, err := file.NewFileData(filePart.FilePartId,
		filePart.SavedMd5Hash,
		inputFile.Name,
		filePart.FileSize,
		filePart.SavedMd5Hash)
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	photoId := base.NextSnowflakeId()
	accessHash := rand.Int63()
	szList, err := photo.UploadPhotoFile(photoId, accessHash, fileData.FilePath, fileData.Ext, false)
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	photo := &mtproto.TLPhoto{Data2: &mtproto.Photo_Data{
		Id:          photoId,
		HasStickers: false,
		AccessHash:  accessHash,
		Date:        int32(time.Now().Unix()),
		Sizes:       szList,
	}}

	// photo:flags.0?Photo caption:flags.1?string ttl_seconds:flags.2?int
	var reply = &mtproto.TLMessageMediaPhoto{Data2: &mtproto.MessageMedia_Data{
		Photo_1:    photo.To_Photo(),
		Caption:    request.GetMedia().GetCaption(),
		TtlSeconds: request.GetMedia().GetTtlSeconds(),
	}}

	glog.Infof("nbfs.uploadedPhotoMedia - reply: %s", logger.JsonDebugData(reply))
	return nil, nil
}

// inputMediaUploadedDocument#e39621fd flags:# file:InputFile thumb:flags.2?InputFile mime_type:string attributes:Vector<DocumentAttribute> caption:string stickers:flags.0?Vector<InputDocument> ttl_seconds:flags.1?int = InputMedia;
func (s *PhotoServiceImpl) NbfsUploadedDocumentMedia(ctx context.Context, request *mtproto.NbfsUploadedDocumentMedia) (*mtproto.TLMessageMediaDocument, error) {
	glog.Infof("nbfs.uploadedDocumentMedia - request: %s", logger.JsonDebugData(request))

	var (
		inputFile = request.GetMedia().GetFile().GetData2()
		// inputThumb = request.GetMedia().GetThumb()
		// reply *mtproto.PhotoDataRsp
		// isBigFile = request.GetMedia().GetFile().GetConstructor() == mtproto.TLConstructor_CRC32_inputFileBig
		thumb *mtproto.PhotoSize
	)

	// TODO(@benqi): 出错以后，回滚数据库操作
	//filePart, err := file.MakeFilePartData(request.OwnerId, inputFile.Id, false, isBigFile)
	//if err != nil {
	//	glog.Error(err)
	//	return nil, err
	//}

	// var filename = core.NBFS_DATA_PATH + filePart.FilePath
	filePart, err := file.DoSavedFilePart(request.OwnerId, inputFile.Id)
	// core.CalcMd5File(filename)
	if filePart.SavedMd5Hash != inputFile.Md5Checksum {
		return nil, fmt.Errorf("check md5 error: %s, %s", filePart.SavedMd5Hash, inputFile.Md5Checksum)
	}

	fileData, err := file.NewFileData(filePart.FilePartId,
		filePart.SavedFilePath,
		inputFile.Name,
		filePart.FileSize,
		filePart.SavedMd5Hash)
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	// upload document file
	documentId := base.NextSnowflakeId()
	data, _ := document2.DoUploadedDocumentFile(documentId,
		fileData.FilePath,
		int32(fileData.FileSize),
		fileData.UploadName,
		fileData.Ext,
		request.GetMedia().GetMimeType(),
		0)

	//thumb := &mtproto.TLPhotoSizeEmpty{Data2: &mtproto.PhotoSize_Data{
	//	Type: "",
	//}}
	if request.GetMedia().GetThumb() != nil {
		thumbFile := request.GetMedia().GetThumb().GetData2()
		// var filename = core.NBFS_DATA_PATH + filePart.FilePath
		filePart, err := file.DoSavedFilePart(request.OwnerId, thumbFile.Id)
		// core.CalcMd5File(filename)
		if filePart.SavedMd5Hash != thumbFile.Md5Checksum {
			return nil, fmt.Errorf("check md5 error: %s, %s", filePart.SavedMd5Hash, thumbFile.Md5Checksum)
		}

		fileData, err := file.NewFileData(filePart.FilePartId,
			filePart.SavedFilePath,
			thumbFile.Name,
			filePart.FileSize,
			filePart.SavedMd5Hash)
		if err != nil {
			glog.Error(err)
			return nil, err
		}

		photoId := base.NextSnowflakeId()
		accessHash := rand.Int63()
		szList, err := photo.UploadPhotoFile(photoId, accessHash, fileData.FilePath, fileData.Ext, false)
		if err != nil {
			glog.Error(err)
			return nil, err
		}

		thumb = &mtproto.PhotoSize{
			Constructor: mtproto.TLConstructor_CRC32_photoSize,
			Data2: szList[0].GetData2(),
		}
		if thumb.Data2.Size == 0 {
			thumb.Data2.Size = int32(len(thumb.Data2.Bytes))
		}
	} else {
		thumb = &mtproto.PhotoSize{
			Constructor: mtproto.TLConstructor_CRC32_photoSizeEmpty,
			Data2: &mtproto.PhotoSize_Data{
				Type: "s",
			},
		}
	}

	document := &mtproto.TLDocument{Data2: &mtproto.Document_Data{
		Id:          documentId,
		AccessHash:  data.AccessHash,
		Date:        int32(time.Now().Unix()),
		MimeType:    data.MimeType,
		Size:        int32(data.FileSize),
		Thumb:       thumb,
		DcId:        2,
		Version:     0,
		Attributes:  request.GetMedia().GetAttributes(),
	}}

	// messageMediaDocument#7c4414d3 flags:# document:flags.0?Document caption:flags.1?string ttl_seconds:flags.2?int = MessageMedia;
	var reply = &mtproto.TLMessageMediaDocument{Data2: &mtproto.MessageMedia_Data{
		Document:   document.To_Document(),
		Caption:    request.GetMedia().GetCaption(),
		TtlSeconds: request.GetMedia().GetTtlSeconds(),
	}}

	glog.Infof("nbfs.uploadedDocumentMedia - reply: %s", logger.JsonDebugData(reply))
	return reply, nil
}

// rpc nbfs_getDocument(DocumentId) returns (PhotoDataRsp);
func (s *PhotoServiceImpl) NbfsGetDocument(ctx context.Context, request *mtproto.DocumentId) (*mtproto.Document, error) {
	glog.Infof("nbfs_getDocument - request: %s", logger.JsonDebugData(request))

	reply := document2.GetDocument(request.Id, request.AccessHash, request.Version)

	glog.Infof("nbfs_getDocument - reply: %s", logger.JsonDebugData(reply))
	return reply, nil
}

// rpc nbfs_getDocumentList(DocumentIdList) returns (DocumentList);
func (s *PhotoServiceImpl) NbfsGetDocumentList(ctx context.Context, request *mtproto.DocumentIdList) (*mtproto.DocumentList, error) {
	glog.Infof("nbfs_getDocumentList - request: %s", logger.JsonDebugData(request))

	documents := document2.GetDocumentList(request.IdList)

	glog.Infof("nbfs_getDocumentList - reply: %s", logger.JsonDebugData(documents))
	return &mtproto.DocumentList{Documents: documents}, nil
}
