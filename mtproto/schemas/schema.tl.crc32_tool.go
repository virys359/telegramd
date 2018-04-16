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
	"flag"
	"github.com/nebulaim/telegramd/mtproto"
	"fmt"
	"strconv"
)

// import "fmt"
//var TLConstructor_CRC32_message2 uint32 		= 0x5bb8e511
//var TLConstructor_CRC32_msg_container uint32 	= 0x73f1f8dc
//var TLConstructor_CRC32_msg_copy uint32 		= 0xe06046b2
//var TLConstructor_CRC32_gzip_packed uint32		= 0x3072cfa1
//var TLConstructor_CRC32_rpc_result uint32		= 0xf35c6d01
//
//var t2 = 0xc09be45f
//fmt.Printf("t2: %d\n", int32(t2))
//
//var t3 = 0x44f9b43d
//fmt.Printf("t2: %d\n", int32(t3))
//
//fmt.Printf("TLConstructor_CRC32_message2: %d\n", int32(TLConstructor_CRC32_message2))
//fmt.Printf("TLConstructor_CRC32_msg_container: %d\n", int32(TLConstructor_CRC32_msg_container))
//fmt.Printf("TLConstructor_CRC32_msg_copy: %d\n", int32(TLConstructor_CRC32_msg_copy))
//fmt.Printf("TLConstructor_CRC32_gzip_packed: %d\n", int32(TLConstructor_CRC32_gzip_packed))
//fmt.Printf("TLConstructor_CRC32_rpc_result: %d\n", int32(TLConstructor_CRC32_rpc_result))
//
//
//var c int32 = 1197350236
//var c2 = 583445000
//fmt.Printf("c2: %x\n", uint32(c2))
//fmt.Printf("c: %x\n", uint32(c))
//fmt.Printf("c: %x\n", uint32(TLConstructor_CRC32_message2))
//fmt.Printf("c: %x\n", uint32(TLConstructor_CRC32_msg_container))
//fmt.Printf("c: %x\n", uint32(TLConstructor_CRC32_msg_copy))
//fmt.Printf("c: %x\n", uint32(TLConstructor_CRC32_gzip_packed))
//fmt.Printf("c: %x\n", uint32(TLConstructor_CRC32_rpc_result))

func main() {
	dec := flag.Int("dec", 0, "Decimal digit")
	hex := flag.String("hex","","Hex digit")

	flag.Parse()

	if *dec != 0 {
		if crc32, ok := mtproto.TLConstructor_name[int32(*dec)]; !ok {
			fmt.Println(*dec, " ==> ", "CRC32_UNKNOWN")
		} else {
			fmt.Println(*dec, " ==> ", crc32)
		}
	} else if *hex != "" {
		n, err := strconv.ParseInt(*hex, 10, 64)
		if err != nil {
			fmt.Println(hex, " ==> ", "CRC32_UNKNOWN")
		} else {
			if crc32, ok := mtproto.TLConstructor_name[int32(n)]; !ok {
				fmt.Println(*hex, " ==> ", "CRC32_UNKNOWN")
			} else {
				fmt.Println(*hex, " ==> ", crc32)
			}
		}
	} else {
		fmt.Println(" ./chema.tl.crc32_tool -dec=xxx\n ./chema.tl.crc32_tool -hex=xxx")
	}
}

