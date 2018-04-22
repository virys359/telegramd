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
	"github.com/nebulaim/telegramd/mtproto"
)

type PhoneCallSessionsDAO struct {
	db *sqlx.DB
}

func NewPhoneCallSessionsDAO(db *sqlx.DB) *PhoneCallSessionsDAO {
	return &PhoneCallSessionsDAO{db}
}

// insert into phone_call_sessions(call_session_id, admin_id, admin_access_hash, participant_id, participant_access_hash, udp_p2p, udp_reflector, min_layer, max_layer, g_a, `date`) values (:call_session_id, :admin_id, :admin_access_hash, :participant_id, :participant_access_hash, :udp_p2p, :udp_reflector, :min_layer, :max_layer, :g_a, :date)
// TODO(@benqi): sqlmap
func (dao *PhoneCallSessionsDAO) Insert(do *dataobject.PhoneCallSessionsDO) int64 {
	var query = "insert into phone_call_sessions(call_session_id, admin_id, admin_access_hash, participant_id, participant_access_hash, udp_p2p, udp_reflector, min_layer, max_layer, g_a, `date`) values (:call_session_id, :admin_id, :admin_access_hash, :participant_id, :participant_access_hash, :udp_p2p, :udp_reflector, :min_layer, :max_layer, :g_a, :date)"
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

// select id, call_session_id, admin_id, admin_access_hash, participant_id, participant_access_hash, udp_p2p, udp_reflector, min_layer, max_layer, g_a, g_b, `date` from phone_call_sessions where call_session_id = :call_session_id
// TODO(@benqi): sqlmap
func (dao *PhoneCallSessionsDAO) Select(call_session_id int64) *dataobject.PhoneCallSessionsDO {
	var query = "select id, call_session_id, admin_id, admin_access_hash, participant_id, participant_access_hash, udp_p2p, udp_reflector, min_layer, max_layer, g_a, g_b, `date` from phone_call_sessions where call_session_id = ?"
	rows, err := dao.db.Queryx(query, call_session_id)

	if err != nil {
		errDesc := fmt.Sprintf("Queryx in Select(_), error: %v", err)
		glog.Error(errDesc)
		panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_DBERR), errDesc))
	}

	defer rows.Close()

	do := &dataobject.PhoneCallSessionsDO{}
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

	return do
}
