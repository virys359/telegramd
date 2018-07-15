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

type FilePartsDAO struct {
	db *sqlx.DB
}

func NewFilePartsDAO(db *sqlx.DB) *FilePartsDAO {
	return &FilePartsDAO{db}
}

// insert into file_parts(creator_id, file_part_id, file_part, is_big_file, file_total_parts, file_path, file_size) values (:creator_id, :file_part_id, :file_part, :is_big_file, :file_total_parts, :file_path, :file_size)
// TODO(@benqi): sqlmap
func (dao *FilePartsDAO) Insert(do *dataobject.FilePartsDO) int64 {
	var query = "insert into file_parts(creator_id, file_part_id, file_part, is_big_file, file_total_parts, file_path, file_size) values (:creator_id, :file_part_id, :file_part, :is_big_file, :file_total_parts, :file_path, :file_size)"
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

// select id, creator_id, file_part_id, file_part, is_big_file, file_total_parts, file_path, file_size from file_parts where creator_id = :creator_id and file_part_id = :file_part_id
// TODO(@benqi): sqlmap
func (dao *FilePartsDAO) SelectFileParts(creator_id int64, file_part_id int64) *dataobject.FilePartsDO {
	var query = "select id, creator_id, file_part_id, file_part, is_big_file, file_total_parts, file_path, file_size from file_parts where creator_id = ? and file_part_id = ?"
	rows, err := dao.db.Queryx(query, creator_id, file_part_id)

	if err != nil {
		errDesc := fmt.Sprintf("Queryx in SelectFileParts(_), error: %v", err)
		glog.Error(errDesc)
		panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_DBERR), errDesc))
	}

	defer rows.Close()

	do := &dataobject.FilePartsDO{}
	if rows.Next() {
		err = rows.StructScan(do)
		if err != nil {
			errDesc := fmt.Sprintf("StructScan in SelectFileParts(_), error: %v", err)
			glog.Error(errDesc)
			panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_DBERR), errDesc))
		}
	} else {
		return nil
	}

	err = rows.Err()
	if err != nil {
		errDesc := fmt.Sprintf("rows in SelectFileParts(_), error: %v", err)
		glog.Error(errDesc)
		panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_DBERR), errDesc))
	}

	return do
}

// update file_parts set file_part = :file_part where id = :id
// TODO(@benqi): sqlmap
func (dao *FilePartsDAO) UpdateFilePart(file_part int32, id int64) int64 {
	var query = "update file_parts set file_part = ? where id = ?"
	r, err := dao.db.Exec(query, file_part, id)

	if err != nil {
		errDesc := fmt.Sprintf("Exec in UpdateFilePart(_), error: %v", err)
		glog.Error(errDesc)
		panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_DBERR), errDesc))
	}

	rows, err := r.RowsAffected()
	if err != nil {
		errDesc := fmt.Sprintf("RowsAffected in UpdateFilePart(_), error: %v", err)
		glog.Error(errDesc)
		panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_DBERR), errDesc))
	}

	return rows
}

// update file_parts set file_part = :file_part, file_total_parts = :file_total_parts, file_size = :file_size where id = :id
// TODO(@benqi): sqlmap
func (dao *FilePartsDAO) UpdateFilePartAndTotal(file_part int32, file_total_parts int32, file_size int64, id int64) int64 {
	var query = "update file_parts set file_part = ?, file_total_parts = ?, file_size = ? where id = ?"
	r, err := dao.db.Exec(query, file_part, file_total_parts, file_size, id)

	if err != nil {
		errDesc := fmt.Sprintf("Exec in UpdateFilePartAndTotal(_), error: %v", err)
		glog.Error(errDesc)
		panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_DBERR), errDesc))
	}

	rows, err := r.RowsAffected()
	if err != nil {
		errDesc := fmt.Sprintf("RowsAffected in UpdateFilePartAndTotal(_), error: %v", err)
		glog.Error(errDesc)
		panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_DBERR), errDesc))
	}

	return rows
}
