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

package photo

import (
	"github.com/nebulaim/telegramd/proto/mtproto"
)

func MakeUserProfilePhoto(photoId int64, sizes []*mtproto.PhotoSize) *mtproto.UserProfilePhoto {
	if len(sizes) == 0 {
		return mtproto.NewTLUserProfilePhotoEmpty().To_UserProfilePhoto()
	}

	// TODO(@benqi): check PhotoSize is photoSizeEmpty
	photo := &mtproto.TLUserProfilePhoto{Data2: &mtproto.UserProfilePhoto_Data{
		PhotoId: photoId,
		PhotoSmall: sizes[0].GetData2().GetLocation(),
		PhotoBig: sizes[len(sizes)-1].GetData2().GetLocation(),
	}}

	return photo.To_UserProfilePhoto()
}

func MakeChatPhoto(sizes []*mtproto.PhotoSize) *mtproto.ChatPhoto {
	if len(sizes) == 0 {
		return mtproto.NewTLChatPhotoEmpty().To_ChatPhoto()
	}

	// TODO(@benqi): check PhotoSize is photoSizeEmpty
	photo := &mtproto.TLChatPhoto{Data2: &mtproto.ChatPhoto_Data{
		PhotoSmall: sizes[0].GetData2().GetLocation(),
		PhotoBig: sizes[len(sizes)-1].GetData2().GetLocation(),
	}}

	return photo.To_ChatPhoto()
}

