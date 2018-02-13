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

// salt cache
package server

import (
	"math/rand"
	"time"
)

var salts = newSaltCache()

type saltCache struct {
	salt int64
	time int64
}

func newSaltCache() *saltCache {
	return &saltCache{
		salt: int64(rand.Uint64()),
		time: time.Now().Unix(),
	}
}

// TODO(@benqi): refresh salt per 24 hours
func getSalt() int64 {
	return salts.salt
}

func checkSalt(salt int64) bool {
	// TODO(@benqi): check salt value and time expired
	return salts.salt == salt
}