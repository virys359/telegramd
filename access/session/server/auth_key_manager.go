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

package server

import (
	"github.com/golang/glog"
	"encoding/base64"
	"github.com/nebulaim/telegramd/biz/dal/dao"
	"github.com/nebulaim/telegramd/biz/dal/dataobject"
	"fmt"
)

type AuthKeyStorager interface {
	GetAuthKey(int64) []byte
	PutAuthKey(int64, []byte) error
}

// "root:@/nebulaim?charset=utf8"
// 30
func NewAuthKeyCacheManager() *AuthKeyCacheManager {
	return &AuthKeyCacheManager{}
}

type AuthKeyCacheManager struct {
}

func (s *AuthKeyCacheManager) GetAuthKey(keyID int64) (authKey []byte) {
	defer func() {
		if r := recover(); r != nil {
			authKey = nil
		}
	}()

	// find by cache
	// var cacheKey CacheAuthKeyItem
	if k, ok := cacheAuthKey.Load(keyID); ok {
		// 本地缓存命中
		cacheKey := k.([]byte)
		if cacheKey != nil {
			authKey = cacheKey
			return
		}
	}

	do := dao.GetAuthKeysDAO(dao.DB_SLAVE).SelectByAuthId(keyID)
	if do == nil {
		glog.Errorf("Read keyData error: not find keyId = %d", keyID)
		return nil
	}

	glog.Info("keyID: ", keyID, ", do: ", do)
	authKey, _ = base64.RawStdEncoding.DecodeString(do.Body)
	cacheAuthKey.Store(keyID, authKey)

	return
}

func (s *AuthKeyCacheManager) PutAuthKey(keyID int64, key []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	do := &dataobject.AuthKeysDO{ AuthId: keyID}
	do.Body = base64.RawStdEncoding.EncodeToString(key)
	dao.GetAuthKeysDAO(dao.DB_MASTER).Insert(do)

	// store cache
	cacheAuthKey.Store(keyID, key)
	return
}

