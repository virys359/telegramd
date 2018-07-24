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
	"fmt"
	"github.com/nebulaim/telegramd/baselib/mysql_client"
	"github.com/nebulaim/telegramd/server/nbfs/biz/core"
	"github.com/nebulaim/telegramd/server/nbfs/biz/dal/dao"
	"io/ioutil"
	"math/rand"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
	mysqlConfig1 := mysql_client.MySQLConfig{
		Name:   "immaster",
		DSN:    "root:@/nebulaim?charset=utf8",
		Active: 5,
		Idle:   2,
	}
	mysqlConfig2 := mysql_client.MySQLConfig{
		Name:   "imslave",
		DSN:    "root:@/nebulaim?charset=utf8",
		Active: 5,
		Idle:   2,
	}
	mysql_client.InstallMysqlClientManager([]mysql_client.MySQLConfig{mysqlConfig1, mysqlConfig2})
	dao.InstallMysqlDAOManager(mysql_client.GetMysqlClientManager())
}

// go test -v -run=TestSaveFilePart
func TestSaveFilePart(t *testing.T) {
	var uploadName = "./test002.jpeg"
	buf, err := ioutil.ReadFile(uploadName)
	if err != nil {
		panic(err)
	}

	sz := len(buf) / kMaxFilePartSize

	var creatorId = rand.Int63()
	var filePartId = rand.Int63()

	var blockSize = kMaxFilePartSize
	// = sz % kMaxFilePartSize
	var uploadFileName string

	var filePart *filePartData
	for i := 0; i <= sz; i++ {
		var isNew = i == 0
		filePart, err = MakeFilePartData(creatorId, filePartId, isNew, false)
		if i == sz {
			blockSize = len(buf) % kMaxFilePartSize
			uploadFileName = filePart.FilePath
		}

		filePart.SaveFilePart(int32(i), buf[i*kMaxFilePartSize:i*kMaxFilePartSize+blockSize])
		fmt.Println(*filePart.FilePartsDO, ", file_part: ", i, ", block_size: ", blockSize)
	}

	md5, _ := core.CalcMd5File(uploadName)
	fileLogic, _ := NewFileData(filePartId, uploadFileName, uploadName, int64(len(buf)), md5)
	_ = fileLogic
}
