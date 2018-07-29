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

package core

import (
	"github.com/nebulaim/telegramd/baselib/base"
	"github.com/nebulaim/telegramd/service/idgen/client"
	"github.com/golang/glog"
)

const (
	// TODO(@benqi): 使用更紧凑的前缀
	boxUpdatesNgenId        = "message_box_ngen_"
	channelBoxUpdatesNgenId = "channel_message_box_ngen_"
)

var seqIDGen idgen.SeqIDGen

func initSeqIDGen(redisName string) {
	var err error
	seqIDGen, err = idgen.NewSeqIDGen("redis", redisName)
	if err != nil {
		glog.Fatal("seqidGen init error: ", err)
	}
}

func NextMessageBoxId(key int32) (seq int64) {
	seq, _ = seqIDGen.GetNextSeqID(boxUpdatesNgenId + base.Int32ToString(key))
	return
}

func CurrentMessageBoxId(key int32) (seq int64) {
	seq, _ = seqIDGen.GetCurrentSeqID(boxUpdatesNgenId + base.Int32ToString(key))
	return
}

func NextChannelMessageBoxId(key int32) (seq int64) {
	seq, _ = seqIDGen.GetNextSeqID(channelBoxUpdatesNgenId + base.Int32ToString(key))
	return
}

func CurrentChannelMessageBoxId(key int32) (seq int64) {
	seq, _ = seqIDGen.GetCurrentSeqID(channelBoxUpdatesNgenId + base.Int32ToString(key))
	return
}
