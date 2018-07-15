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

package document

import (
	"fmt"
	"os"
	"time"
	"math/rand"
	"encoding/json"
	"github.com/golang/glog"
	"github.com/nebulaim/telegramd/proto/mtproto"
	"github.com/nebulaim/telegramd/server/nbfs/biz/dal/dataobject"
	"github.com/nebulaim/telegramd/server/nbfs/biz/dal/dao"
	"github.com/nebulaim/telegramd/server/nbfs/biz/core"
	"github.com/nebulaim/telegramd/server/nbfs/biz/core/photo"
)

type documentData struct {
	*dataobject.DocumentsDO
}

func DoUploadedDocumentFile(documentId int64, filePath string, fileSize int32, uploadedFileName, extName, mimeType string, thumbId int64) (*documentData, error) {
	data := &dataobject.DocumentsDO{
		DocumentId:       documentId,
		AccessHash:       rand.Int63(),
		DcId:             2,
		FilePath:         filePath,
		FileSize:         fileSize,
		UploadedFileName: uploadedFileName,
		Ext:              extName,
		MimeType:         mimeType,
		ThumbId:          thumbId,
		Version:          0,
	}
	data.Id = dao.GetDocumentsDAO(dao.DB_MASTER).Insert(data)
	return &documentData{DocumentsDO: data}, nil
}

func GetDocumentFileData(id, accessHash int64, version int32, offset, limit int32) (*mtproto.Upload_File, error) {
	do := dao.GetDocumentsDAO(dao.DB_MASTER).SelectByFileLocation(id, accessHash, version)
	if do == nil {
		return nil, fmt.Errorf("not found document <%d, %d, %d>", id, accessHash, version)
	}

	if offset > do.FileSize {
		limit = 0
	} else if offset + limit > do.FileSize {
		limit = do.FileSize - offset
	}

	var filename = core.NBFS_DATA_PATH + do.FilePath
	f, err := os.Open(filename)
	if err != nil {
		glog.Error("open ", filename, " error: ", err)
		return nil, err
	}
	defer f.Close()

	bytes := make([]byte, limit)
	_, err = f.ReadAt(bytes, int64(offset))
	if err != nil {
		glog.Error("read file ", filename, " error: ", err)
		return nil, err
	}

	uploadFile := &mtproto.TLUploadFile{Data2: &mtproto.Upload_File_Data{
		Type: core.MakeStorageFileType(do.Ext),
		Mtime: int32(time.Now().Unix()),
		Bytes: bytes,
	}}
	return uploadFile.To_Upload_File(), nil
}

func makeDocumentByDO(do *dataobject.DocumentsDO) *mtproto.Document {
	var (
		thumb *mtproto.PhotoSize
		document *mtproto.Document
	)

	if do == nil {
		document = mtproto.NewTLDocumentEmpty().To_Document()
	} else {
		if do.ThumbId != 0 {
			sizeList := photo.GetPhotoSizeList(do.ThumbId)
			if len(sizeList) > 0 {
				thumb = sizeList[0]
			}
		}
		if thumb == nil {
			thumb = mtproto.NewTLPhotoSizeEmpty().To_PhotoSize()
		}

		attributes := &mtproto.DocumentAttributeList{}
		err := json.Unmarshal([]byte(do.Attributes), attributes)
		if err != nil {
			glog.Error(err)
			attributes.Attributes = []*mtproto.DocumentAttribute{}
		}

		// if do.Attributes
		document = &mtproto.Document{
			Constructor: mtproto.TLConstructor_CRC32_document,
			Data2: &mtproto.Document_Data{
				Id:          do.DocumentId,
				AccessHash:  do.AccessHash,
				Date:        int32(time.Now().Unix()),
				MimeType:    do.MimeType,
				Size:        do.FileSize,
				Thumb:       thumb,
				DcId:        2,
				Version:     do.Version,
				Attributes:  attributes.Attributes,
			},
		}
	}

	return document
}

func GetDocument(id, accessHash int64, version int32) (*mtproto.Document) {
	do := dao.GetDocumentsDAO(dao.DB_SLAVE).SelectByFileLocation(id, accessHash, version)
	if do == nil {
		glog.Warning("")
	}
	return makeDocumentByDO(do)
}

func GetDocumentList(idList []int64) ([]*mtproto.Document) {
	doList := dao.GetDocumentsDAO(dao.DB_SLAVE).SelectByIdList(idList)
	documetList := make([]*mtproto.Document, len(doList))
	for i := 0; i < len(doList); i++ {
		documetList[i] = makeDocumentByDO(&doList[i])
	}
	return documetList
}

//func MakeStickerDocuemnt(id, accessHash int64) *mtproto.Document {
//	//
//	thumb := &mtproto.TLPhotoSize{Data2: &mtproto.PhotoSize_Data{
//		Type:     "m",
//		Location: &mtproto.FileLocation{
//			Constructor: mtproto.TLConstructor_CRC32_fileLocation,
//			Data2: 		 &mtproto.FileLocation_Data{
//				DcId:     2,
//				VolumeId: 226607193,
//				LocalId:  35316,
//				Secret:   2434071452955778983,
//			}},
//		W:        128,
//		H:        128,
//		Size:     5648,
//	}}
//
//	document := &mtproto.TLDocument{Data2: &mtproto.Document_Data{
//		Id:          id,
//		AccessHash:  accessHash,
//		Date:        int32(time.Now().Unix()),
//		MimeType:    "image/jpg",
//		Size:        37634,
//		Thumb:       thumb.To_PhotoSize(),
//		DcId:        2,
//		Version:     0,
//		Attributes:  MakeDocumentAttributes(id, accessHash),
//	}}
//
//	return document.To_Document()
//}
//
//func MakeDocumentAttributes(id, accessHash int64) []*mtproto.DocumentAttribute {
//	attributes := make([]*mtproto.DocumentAttribute, 0, 3)
//	imageSize := &mtproto.TLDocumentAttributeImageSize{Data2: &mtproto.DocumentAttribute_Data{
//		W: 512,
//		H: 512,
//	}}
//	attributes = append(attributes, imageSize.To_DocumentAttribute())
//
//	sticker := &mtproto.TLDocumentAttributeSticker{Data2: &mtproto.DocumentAttribute_Data{
//		Alt: "😂",
//		Stickerset: &mtproto.InputStickerSet{
//			Constructor: mtproto.TLConstructor_CRC32_inputStickerSetID,
//			Data2: &mtproto.InputStickerSet_Data{
//				Id:         id,
//				AccessHash: accessHash,
//			},
//		},
//		MaskCoords: nil,
//	}}
//	attributes = append(attributes, sticker.To_DocumentAttribute())
//
//	fileName := &mtproto.TLDocumentAttributeFilename{Data2: &mtproto.DocumentAttribute_Data{
//		FileName: "sticker.wep",
//	}}
//	attributes = append(attributes, fileName.To_DocumentAttribute())
//
//	return attributes
//}
