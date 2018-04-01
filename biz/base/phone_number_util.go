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

package base

import (
	"github.com/ttacon/libphonenumber"
	"fmt"
	"github.com/nebulaim/telegramd/mtproto"
)

// Check number
// 客户端发送的手机号格式为: "+86 111 1111 1111"，归一化
func CheckAndGetPhoneNumber(number string) (phoneNumber string, err error) {
	if number == "" {
		err = mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_PHONE_NUMBER_INVALID), "phone number empty")
		return
	}

	// Android客户端手机号格式为: 8611111111111, Parse结果为invalid country code
	// 转换成+8611111111111，再进行Parse
	if number[:1] != "+" {
		number = "+" + number
	}

	// check phone invalid
	var pnumber *libphonenumber.PhoneNumber
	pnumber, err = libphonenumber.Parse(number, "")
	if err != nil {
		err = mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_PHONE_NUMBER_INVALID), fmt.Sprintf("invalid phone number: %v", err))
	} else {
		if !libphonenumber.IsValidNumber(pnumber) {
			err = mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_PHONE_NUMBER_INVALID), "invalid phone number")
		}
	}

	if err != nil {
		// DB里存储归一化的phone
		phoneNumber = libphonenumber.NormalizeDigitsOnly(number)
	}

	return
}
