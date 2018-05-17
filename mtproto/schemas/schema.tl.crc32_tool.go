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

package main

import (
	"github.com/nebulaim/telegramd/mtproto"
	"fmt"
	"strconv"
	"os"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println(" ./schema.tl.crc32_tool xxx [...]")
		os.Exit(0)
	}

	for i := 1; i < len(os.Args); i++ {
		n, err := strconv.ParseInt(os.Args[i], 0, 64)
		if err != nil {
			fmt.Println(os.Args[i], " conv error: ", err)
		} else {
			if crc32, ok := mtproto.TLConstructor_name[int32(n)]; !ok {
				fmt.Printf("[%d, 0x%x] ==> %s\n", int32(n), uint32(n), "CRC32_UNKNOWN")
			} else {
				fmt.Printf("[%d, 0x%x] ==> %s\n", int32(n), uint32(n), crc32)
			}
		}
	}
}
