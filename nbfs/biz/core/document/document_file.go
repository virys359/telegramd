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
	"github.com/nebulaim/telegramd/nbfs/biz/dal/dataobject"
	"github.com/nebulaim/telegramd/nbfs/biz/dal/dao"
	"github.com/nebulaim/telegramd/mtproto"
	"fmt"
	"github.com/golang/glog"
	"github.com/nebulaim/telegramd/nbfs/biz/core"
	"os"
	"time"
	"math/rand"
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
