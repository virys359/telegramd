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

package file

import (
	"github.com/nebulaim/telegramd/nbfs/biz/dal/dataobject"
	"fmt"
	"github.com/nebulaim/telegramd/nbfs/biz/dal/dao"
	"os"
	"github.com/golang/glog"
	"github.com/nebulaim/telegramd/nbfs/biz/base"
	// "math/rand"
	base2 "github.com/nebulaim/telegramd/baselib/base"
	"github.com/nebulaim/telegramd/nbfs/biz/core"
)

const (
	kMaxFilePartSize = 32768
)

type filePartData struct {
	*dataobject.FilePartsDO
}

func MakeFilePartData(creatorId, filePartId int64, isNew, isBigFile bool) (*filePartData, error) {
	var data *dataobject.FilePartsDO

	if isNew {
		data = &dataobject.FilePartsDO{
			CreatorId:      creatorId,
			FilePartId:     filePartId,
			FilePart:       -1,
			IsBigFile:      base2.BoolToInt8(isBigFile),
			// FileTotalParts: fileTotalParts,
			FilePath:       fmt.Sprintf("/0/%d.cache", base.NextSnowflakeId()),
		}
	} else {
		data = dao.GetFilePartsDAO(dao.DB_SLAVE).SelectFileParts(creatorId, filePartId)
		if data == nil {
			return nil, fmt.Errorf("not found file_parts by {creator_id: %d, file_id: %d}", creatorId, filePartId)
		}
	}

	return &filePartData{FilePartsDO: data}, nil
}

func saveFileData(fileName string, filePart int32, bytes []byte) error {
	if filePart == 0 {
		f, err := os.Create(fileName)
		if err != nil {
			glog.Error(err)
			return err
		}
		defer f.Close()
		_, err = f.Write(bytes)
		if err != nil {
			glog.Error(err)
			return err
		}
		f.Sync()
	} else {
		// Check file exists and file size
		if fi, err := os.Stat(fileName); os.IsNotExist(err) {
			// exist = false;
			glog.Error(err)
			return err
		} else {
			if fi.Size() != int64(filePart)*kMaxFilePartSize {
				err = fmt.Errorf("file size error")
				return err
			}
		}

		f, err := os.OpenFile(fileName, os.O_RDWR|os.O_APPEND, 0644)  //打开文件
		if err != nil {
			glog.Error(fileName, ": ", err)
			return err
		}
		defer f.Close()
		_, err = f.Write(bytes)
		if err != nil {
			glog.Error(fileName, ": ", err)
			return err
		}
		f.Sync()
	}

	return nil
}

func (m *filePartData) SaveFilePart(filePart int32, bytes []byte) error {
	if filePart - m.FilePart != 1 {
		return fmt.Errorf("bad request")
	}

	fileName := core.NBFS_DATA_PATH + m.FilePath
	err := saveFileData(fileName, filePart, bytes)
	if err != nil {
		return err
	}

	var begin = filePart == 0
	var end = len(bytes) < kMaxFilePartSize

	m.FilePart = filePart
	if end {
		m.FileTotalParts = filePart + 1
		m.FileSize = int64(filePart)*kMaxFilePartSize + int64(len(bytes))
	}

	if begin {
		dao.GetFilePartsDAO(dao.DB_MASTER).Insert(m.FilePartsDO)
	} else {
		if end {
			dao.GetFilePartsDAO(dao.DB_MASTER).UpdateFilePartAndTotal(filePart, m.FileTotalParts, m.FileSize, m.Id)
		} else {
			dao.GetFilePartsDAO(dao.DB_MASTER).UpdateFilePart(filePart, m.Id)
		}
	}

	//if end {
	//	// 文件上传结束, 计算出文件大小和md5，存盘
	//	var (
	//		fileId = base.NextSnowflakeId()
	//		accessHash = rand.Int63()
	//	)
	//
	//	if m.data.IsBigFile == 0 {
	//		// TODO(@benqi): md5优化
	//		md5Checksum, _ := md5File(m.data.FilePath)
	//		return m.onUploadEnd(fileId, accessHash, md5Checksum)
	//	} else {
	//		return m.onUploadEnd(fileId, accessHash, "")
	//	}
	//}

	return nil
}
