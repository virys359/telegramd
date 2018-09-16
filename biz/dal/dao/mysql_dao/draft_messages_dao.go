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
	"github.com/nebulaim/telegramd/biz/dal/dataobject"
	"github.com/nebulaim/telegramd/proto/mtproto"
)

type DraftMessagesDAO struct {
	db *sqlx.DB
}

func NewDraftMessagesDAO(db *sqlx.DB) *DraftMessagesDAO {
	return &DraftMessagesDAO{db}
}

// insert into draft_messages(user_id, draft_id, draft_type, draft_message_data) values (:user_id, :draft_id, 2, :draft_message_data) on duplicate key update draft_type = 2, draft_message_data = values(draft_message_data)
// TODO(@benqi): sqlmap
func (dao *DraftMessagesDAO) InsertOrUpdate(do *dataobject.DraftMessagesDO) int64 {
	var query = "insert into draft_messages(user_id, draft_id, draft_type, draft_message_data) values (:user_id, :draft_id, 2, :draft_message_data) on duplicate key update draft_type = 2, draft_message_data = values(draft_message_data)"
	r, err := dao.db.NamedExec(query, do)
	if err != nil {
		errDesc := fmt.Sprintf("NamedExec in InsertOrUpdate(%v), error: %v", do, err)
		glog.Error(errDesc)
		panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_DBERR), errDesc))
	}

	id, err := r.LastInsertId()
	if err != nil {
		errDesc := fmt.Sprintf("LastInsertId in InsertOrUpdate(%v)_error: %v", do, err)
		glog.Error(errDesc)
		panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_DBERR), errDesc))
	}
	return id
}

// update draft_messages set draft_type = 0, draft_message_data = '' where user_id = :user_id and draft_id = :draft_id and draft_type = 2
// TODO(@benqi): sqlmap
func (dao *DraftMessagesDAO) ClearDraft(user_id int32, draft_id int32) int64 {
	var query = "update draft_messages set draft_type = 0, draft_message_data = '' where user_id = ? and draft_id = ? and draft_type = 2"
	r, err := dao.db.Exec(query, user_id, draft_id)

	if err != nil {
		errDesc := fmt.Sprintf("Exec in ClearDraft(_), error: %v", err)
		glog.Error(errDesc)
		panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_DBERR), errDesc))
	}

	rows, err := r.RowsAffected()
	if err != nil {
		errDesc := fmt.Sprintf("RowsAffected in ClearDraft(_), error: %v", err)
		glog.Error(errDesc)
		panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_DBERR), errDesc))
	}

	return rows
}

// select user_id, draft_id, draft_type, draft_message_data from draft_messages where user_id = :user_id and draft_type = 2
// TODO(@benqi): sqlmap
func (dao *DraftMessagesDAO) SelectAllDraft(user_id int32) *dataobject.DraftMessagesDO {
	var query = "select user_id, draft_id, draft_type, draft_message_data from draft_messages where user_id = ? and draft_type = 2"
	rows, err := dao.db.Queryx(query, user_id)

	if err != nil {
		errDesc := fmt.Sprintf("Queryx in SelectAllDraft(_), error: %v", err)
		glog.Error(errDesc)
		panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_DBERR), errDesc))
	}

	defer rows.Close()

	do := &dataobject.DraftMessagesDO{}
	if rows.Next() {
		err = rows.StructScan(do)
		if err != nil {
			errDesc := fmt.Sprintf("StructScan in SelectAllDraft(_), error: %v", err)
			glog.Error(errDesc)
			panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_DBERR), errDesc))
		}
	} else {
		return nil
	}

	err = rows.Err()
	if err != nil {
		errDesc := fmt.Sprintf("rows in SelectAllDraft(_), error: %v", err)
		glog.Error(errDesc)
		panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_DBERR), errDesc))
	}

	return do
}

// select user_id, draft_id, draft_type, draft_message_data from draft_messages where user_id = :user_id and draft_id in (:idList) and draft_type = 2
// TODO(@benqi): sqlmap
func (dao *DraftMessagesDAO) SelectDraftList(user_id int32, idList []int32) *dataobject.DraftMessagesDO {
	var q = "select user_id, draft_id, draft_type, draft_message_data from draft_messages where user_id = ? and draft_id in (?) and draft_type = 2"
	query, a, err := sqlx.In(q, user_id, idList)
	rows, err := dao.db.Queryx(query, a...)

	if err != nil {
		errDesc := fmt.Sprintf("Queryx in SelectDraftList(_), error: %v", err)
		glog.Error(errDesc)
		panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_DBERR), errDesc))
	}

	defer rows.Close()

	do := &dataobject.DraftMessagesDO{}
	if rows.Next() {
		err = rows.StructScan(do)
		if err != nil {
			errDesc := fmt.Sprintf("StructScan in SelectDraftList(_), error: %v", err)
			glog.Error(errDesc)
			panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_DBERR), errDesc))
		}
	} else {
		return nil
	}

	err = rows.Err()
	if err != nil {
		errDesc := fmt.Sprintf("rows in SelectDraftList(_), error: %v", err)
		glog.Error(errDesc)
		panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_DBERR), errDesc))
	}

	return do
}

// select user_id, draft_id, draft_type, draft_message_data from draft_messages where user_id = :user_id and draft_id = :draft_id and draft_type = 2
// TODO(@benqi): sqlmap
func (dao *DraftMessagesDAO) SelectDraft(user_id int32, draft_id int32) *dataobject.DraftMessagesDO {
	var query = "select user_id, draft_id, draft_type, draft_message_data from draft_messages where user_id = ? and draft_id = ? and draft_type = 2"
	rows, err := dao.db.Queryx(query, user_id, draft_id)

	if err != nil {
		errDesc := fmt.Sprintf("Queryx in SelectDraft(_), error: %v", err)
		glog.Error(errDesc)
		panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_DBERR), errDesc))
	}

	defer rows.Close()

	do := &dataobject.DraftMessagesDO{}
	if rows.Next() {
		err = rows.StructScan(do)
		if err != nil {
			errDesc := fmt.Sprintf("StructScan in SelectDraft(_), error: %v", err)
			glog.Error(errDesc)
			panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_DBERR), errDesc))
		}
	} else {
		return nil
	}

	err = rows.Err()
	if err != nil {
		errDesc := fmt.Sprintf("rows in SelectDraft(_), error: %v", err)
		glog.Error(errDesc)
		panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_DBERR), errDesc))
	}

	return do
}
