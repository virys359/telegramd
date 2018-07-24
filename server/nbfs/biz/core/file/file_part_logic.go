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
	"crypto/md5"
	"fmt"
	"github.com/golang/glog"
	base2 "github.com/nebulaim/telegramd/baselib/base"
	"github.com/nebulaim/telegramd/server/nbfs/biz/base"
	"github.com/nebulaim/telegramd/server/nbfs/biz/core"
	"github.com/nebulaim/telegramd/server/nbfs/biz/dal/dao"
	"github.com/nebulaim/telegramd/server/nbfs/biz/dal/dataobject"
	"io/ioutil"
	"os"
)

const (
	kMaxFilePartSize = 65536
)

type filePartData struct {
	SavedFilePath string
	SavedMd5Hash  string
	*dataobject.FilePartsDO
}

// 判断文件是否存在
func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func DoSavedFilePart(creatorId, filePartId int64) (*filePartData, error) {
	data := dao.GetFilePartsDAO(dao.DB_MASTER).SelectFileParts(creatorId, filePartId)
	if data == nil {
		return nil, fmt.Errorf("not found file_parts by {creator_id: %d, file_id: %d}", creatorId, filePartId)
	}

	// 1. calc md5, file_size, total_file_parts
	md5Hash := md5.New()
	fileSzie := 0
	// FilePath:
	newFilePath := fmt.Sprintf("/0/%d.cache", base.NextSnowflakeId())
	newFileName := core.NBFS_DATA_PATH + newFilePath
	f, err := os.Create(newFileName)
	if err != nil {
		glog.Error(err)
		return nil, err
	}
	defer f.Close()
	for i := int32(0); i <= data.FilePart; i++ {
		fileName := fmt.Sprintf("%s%s/%d.part", core.NBFS_DATA_PATH, data.FilePath, i)
		exist, err := pathExists(fileName)
		if err != nil {
			glog.Errorf("path - %s, error : %v", fileName, err)
			return nil, err
		}
		if !exist {
			err = fmt.Errorf("path not exists: %s", fileName)
			return nil, err
		}

		b, err := ioutil.ReadFile(fileName)
		if err != nil {
			err = fmt.Errorf("path not exists: %s", fileName)
			return nil, err
		}

		md5Hash.Write(b)
		fileSzie += len(b)
		_, err = f.Write(b)
		if err != nil {
			glog.Error(err)
			return nil, err
		}
	}
	f.Sync()

	data.FileSize = int64(fileSzie)
	data.FileTotalParts = data.FilePart + 1
	data.FilePath = newFilePath

	dao.GetFilePartsDAO(dao.DB_MASTER).UpdateFilePartAndTotal(data.FilePart, data.FileTotalParts, data.FileSize, data.Id)

	return &filePartData{
		SavedFilePath: newFilePath,
		SavedMd5Hash:  fmt.Sprintf("%x", md5Hash.Sum(nil)),
		FilePartsDO:   data,
	}, nil
}

func MakeFilePartData(creatorId, filePartId int64, isNew, isBigFile bool) (*filePartData, error) {
	var data *dataobject.FilePartsDO
	data = dao.GetFilePartsDAO(dao.DB_MASTER).SelectFileParts(creatorId, filePartId)
	if data == nil {
		data = &dataobject.FilePartsDO{
			CreatorId:  creatorId,
			FilePartId: filePartId,
			FilePart:   -1,
			IsBigFile:  base2.BoolToInt8(isBigFile),
			// FileTotalParts: fileTotalParts,
			FilePath: fmt.Sprintf("/0/%d.parts", base.NextSnowflakeId()),
		}
		data.Id = dao.GetFilePartsDAO(dao.DB_MASTER).Insert(data)
	}
	//if isNew {
	//} else {
	//	if data == nil {
	//		return nil, fmt.Errorf("not found file_parts by {creator_id: %d, file_id: %d}", creatorId, filePartId)
	//	}
	//}
	return &filePartData{FilePartsDO: data}, nil
}

func saveFileData(filePath string, filePart int32, bytes []byte) error {
	exist, err := pathExists(filePath)
	if err != nil {
		glog.Errorf("pathExists error![%v]", err)
		return err
	}

	if !exist {
		err := os.Mkdir(filePath, 0755)
		if err != nil {
			glog.Errorf("mkdir failed![%v]\n", err)
			return err
		}
	}

	fileName := fmt.Sprintf("%s/%d.part", filePath, filePart)
	exist, _ = pathExists(fileName)
	if !exist {
		err = ioutil.WriteFile(fileName, bytes, 0644)
		if err != nil {
			glog.Error(err)
			return err
		}
	}

	//f, err := os.Create(fileName)
	//if err != nil {
	//	glog.Error(err)
	//	return err
	//}
	//defer f.Close()
	//_, err = f.Write(bytes)
	//if err != nil {
	//	glog.Error(err)
	//	return err
	//}
	//f.Sync()

	return nil
}

func (m *filePartData) SaveFilePart(filePart int32, bytes []byte) error {
	//if filePart - m.FilePart != 1 {
	//	return fmt.Errorf("bad request, err: {filePart: %d, m.FilePart: %d}", filePart, m.FilePart)
	//}

	fileName := core.NBFS_DATA_PATH + m.FilePath
	err := saveFileData(fileName, filePart, bytes)
	if err != nil {
		return err
	}

	if m.FilePart < filePart {
		dao.GetFilePartsDAO(dao.DB_MASTER).UpdateFilePart(filePart, m.Id)
	}

	//var begin = filePart == 0
	//var end = len(bytes) < kMaxFilePartSize
	//
	//m.FilePart = filePart
	//if end {
	//	m.FileTotalParts = filePart + 1
	//	m.FileSize = int64(filePart)*kMaxFilePartSize + int64(len(bytes))
	//}
	//
	//if begin {
	//	dao.GetFilePartsDAO(dao.DB_MASTER).Insert(m.FilePartsDO)
	//} else {
	//	if end {
	//		dao.GetFilePartsDAO(dao.DB_MASTER).UpdateFilePartAndTotal(filePart, m.FileTotalParts, m.FileSize, m.Id)
	//	} else {
	//		dao.GetFilePartsDAO(dao.DB_MASTER).UpdateFilePart(filePart, m.Id)
	//	}
	//}
	//
	//if end {
	//	// 文件上传结束, 计算出文件大小和md5，存盘
	//	var (
	//		fileId = helper.NextSnowflakeId()
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
