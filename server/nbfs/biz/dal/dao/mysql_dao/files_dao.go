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

package mysql_dao

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/jmoiron/sqlx"
	"github.com/nebulaim/telegramd/proto/mtproto"
	"github.com/nebulaim/telegramd/server/nbfs/biz/dal/dataobject"
)

type FilesDAO struct {
	db *sqlx.DB
}

func NewFilesDAO(db *sqlx.DB) *FilesDAO {
	return &FilesDAO{db}
}

// insert into files(file_id, access_hash, file_size, file_path, ext, md5_checksum, upload_name) values (:file_id, :access_hash, :file_size, :file_path, :ext, :md5_checksum, :upload_name)
// TODO(@benqi): sqlmap
func (dao *FilesDAO) Insert(do *dataobject.FilesDO) int64 {
	var query = "insert into files(file_id, access_hash, file_size, file_path, ext, md5_checksum, upload_name) values (:file_id, :access_hash, :file_size, :file_path, :ext, :md5_checksum, :upload_name)"
	r, err := dao.db.NamedExec(query, do)
	if err != nil {
		errDesc := fmt.Sprintf("NamedExec in Insert(%v), error: %v", do, err)
		glog.Error(errDesc)
		panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_DBERR), errDesc))
	}

	id, err := r.LastInsertId()
	if err != nil {
		errDesc := fmt.Sprintf("LastInsertId in Insert(%v)_error: %v", do, err)
		glog.Error(errDesc)
		panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_DBERR), errDesc))
	}
	return id
}

// select id, file_id, access_hash, file_size, file_path, ext, md5_checksum, upload_name from files where file_id = :file_id
// TODO(@benqi): sqlmap
func (dao *FilesDAO) Select(file_id int64) *dataobject.FilesDO {
	var query = "select id, file_id, access_hash, file_size, file_path, ext, md5_checksum, upload_name from files where file_id = ?"
	rows, err := dao.db.Queryx(query, file_id)

	if err != nil {
		errDesc := fmt.Sprintf("Queryx in Select(_), error: %v", err)
		glog.Error(errDesc)
		panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_DBERR), errDesc))
	}

	defer rows.Close()

	do := &dataobject.FilesDO{}
	if rows.Next() {
		err = rows.StructScan(do)
		if err != nil {
			errDesc := fmt.Sprintf("StructScan in Select(_), error: %v", err)
			glog.Error(errDesc)
			panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_DBERR), errDesc))
		}
	} else {
		return nil
	}

	err = rows.Err()
	if err != nil {
		errDesc := fmt.Sprintf("rows in Select(_), error: %v", err)
		glog.Error(errDesc)
		panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_DBERR), errDesc))
	}

	return do
}
