/*
 *  Copyright (c) 2017, https://github.com/nebulaim
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
	"github.com/nebulaim/telegramd/biz_model/dal/dataobject"
	"github.com/nebulaim/telegramd/mtproto"
)

type AuthUpdatesStateDAO struct {
	db *sqlx.DB
}

func NewAuthUpdatesStateDAO(db *sqlx.DB) *AuthUpdatesStateDAO {
	return &AuthUpdatesStateDAO{db}
}

// insert into auth_updates_state(auth_key_id, user_id, pts, qts, seq, date2, created_at) values (:auth_key_id, :user_id, :pts, :qts, :seq, :date2, :created_at)
// TODO(@benqi): sqlmap
func (dao *AuthUpdatesStateDAO) Insert(do *dataobject.AuthUpdatesStateDO) int64 {
	var query = "insert into auth_updates_state(auth_key_id, user_id, pts, qts, seq, date2, created_at) values (:auth_key_id, :user_id, :pts, :qts, :seq, :date2, :created_at)"
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

// update auth_updates_state set pts = :pts, qts = :qts where auth_key_id = :auth_key_id
// TODO(@benqi): sqlmap
func (dao *AuthUpdatesStateDAO) UpdatePtsAndQts(pts int32, qts int32, auth_key_id int64) int64 {
	var query = "update auth_updates_state set pts = ?, qts = ? where auth_key_id = ?"
	r, err := dao.db.Exec(query, pts, qts, auth_key_id)

	if err != nil {
		errDesc := fmt.Sprintf("Exec in UpdatePtsAndQts(_), error: %v", err)
		glog.Error(errDesc)
		panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_DBERR), errDesc))
	}

	rows, err := r.RowsAffected()
	if err != nil {
		errDesc := fmt.Sprintf("RowsAffected in UpdatePtsAndQts(_), error: %v", err)
		glog.Error(errDesc)
		panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_DBERR), errDesc))
	}

	return rows
}

// select pts, qts, seq, date2 from auth_updates_state where auth_key_id = :auth_key_id
// TODO(@benqi): sqlmap
func (dao *AuthUpdatesStateDAO) SelectByAuthId(auth_key_id int64) *dataobject.AuthUpdatesStateDO {
	var query = "select pts, qts, seq, date2 from auth_updates_state where auth_key_id = ?"
	rows, err := dao.db.Queryx(query, auth_key_id)

	if err != nil {
		errDesc := fmt.Sprintf("Queryx in SelectByAuthId(_), error: %v", err)
		glog.Error(errDesc)
		panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_DBERR), errDesc))
	}

	defer rows.Close()

	do := &dataobject.AuthUpdatesStateDO{}
	if rows.Next() {
		err = rows.StructScan(do)
		if err != nil {
			errDesc := fmt.Sprintf("StructScan in SelectByAuthId(_), error: %v", err)
			glog.Error(errDesc)
			panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_DBERR), errDesc))
		}
	} else {
		return nil
	}

	return do
}
