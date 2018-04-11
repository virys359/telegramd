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

package rpc

import (
	"github.com/golang/glog"
	"github.com/nebulaim/telegramd/baselib/logger"
	"github.com/nebulaim/telegramd/grpc_util"
	"github.com/nebulaim/telegramd/mtproto"
	"golang.org/x/net/context"
	photo2 "github.com/nebulaim/telegramd/biz/core/photo"
	"github.com/nebulaim/telegramd/biz/base"
	"time"
	"github.com/nebulaim/telegramd/biz_server/sync_client"
	"github.com/nebulaim/telegramd/biz/core/user"
)

/*
 rpc_request:
	body: { photos_uploadProfilePhoto
	  file: { inputFile
		id: 6523161970854807437 [LONG],
		parts: 7 [INT],
		name: ".jpg" [STRING],
		md5_checksum: "1f5d86186993d946f803b4ae87348ca7" [STRING],
	  },
	},

 rpc_result:
	body: { rpc_result
	  req_msg_id: 6537202186025275052 [LONG],
	  result: { photos_photo
		photo: { photo
		  flags: 0 [INT],
		  has_stickers: [ SKIPPED BY BIT 0 IN FIELD flags ],
		  id: 1136864293085620139 [LONG],
		  access_hash: 8285278371175870058 [LONG],
		  date: 1522060992 [INT],
		  sizes: [ vector<0x0>
			{ photoSize
			  type: "a" [STRING],
			  location: { fileLocation
				dc_id: 5 [INT],
				volume_id: 852737464 [LONG],
				local_id: 120947 [INT],
				secret: 11287062036346093130 [LONG],
			  },
			  w: 160 [INT],
			  h: 160 [INT],
			  size: 12165 [INT],
			},
			{ photoSize
			  type: "b" [STRING],
			  location: { fileLocation
				dc_id: 5 [INT],
				volume_id: 852737464 [LONG],
				local_id: 120948 [INT],
				secret: 16640413773901356347 [LONG],
			  },
			  w: 320 [INT],
			  h: 320 [INT],
			  size: 42651 [INT],
			},
			{ photoSize
			  type: "c" [STRING],
			  location: { fileLocation
				dc_id: 5 [INT],
				volume_id: 852737464 [LONG],
				local_id: 120949 [INT],
				secret: 5379568561184249960 [LONG],
			  },
			  w: 640 [INT],
			  h: 640 [INT],
			  size: 148360 [INT],
			},
		  ],
		},
		users: [ vector<0x0> ],
	  },
	},

 updates:
	body: { updateShort
	  update: { updateUserPhoto
		user_id: 264696845 [INT],
		date: 1522060992 [INT],
		photo: { userProfilePhoto
		  photo_id: 1136864293085620139 [LONG],
		  photo_small: { fileLocation
			dc_id: 5 [INT],
			volume_id: 852737464 [LONG],
			local_id: 120947 [INT],
			secret: 11287062036346093130 [LONG],
		  },
		  photo_big: { fileLocation
			dc_id: 5 [INT],
			volume_id: 852737464 [LONG],
			local_id: 120949 [INT],
			secret: 5379568561184249960 [LONG],
		  },
		},
		previous: { boolFalse },
	  },
	  date: 1522060991 [INT],
	},
 */

// photos.uploadProfilePhoto#4f32c098
// Updates current user profile photo.
// file: File saved in parts by means of upload.saveFilePart method
//
// photos.uploadProfilePhoto#4f32c098 file:InputFile = photos.Photo;
func (s *PhotosServiceImpl) PhotosUploadProfilePhoto(ctx context.Context, request *mtproto.TLPhotosUploadProfilePhoto) (*mtproto.Photos_Photo, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("photos.uploadProfilePhoto#4f32c098 - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	file := request.GetFile()
	uuid := base.NextSnowflakeId()

	sizes, err := photo2.UploadPhoto(md.UserId, uuid, file.GetData2().GetId(), file.GetData2().GetParts(), file.GetData2().GetName(), file.GetData2().GetMd5Checksum())
	if err != nil {
		glog.Errorf("UploadPhoto error: %v", err)
		return nil, err
	}

	user.SetUserPhotoID(md.UserId, uuid)

	// TODO(@benqi): sync update userProfilePhoto

	// fileData := mediaData.GetFile().GetData2()
	photo := &mtproto.TLPhoto{ Data2: &mtproto.Photo_Data{
		Id:          uuid,
		HasStickers: false,
		AccessHash:  photo2.GetFileAccessHash(file.GetData2().GetId(), file.GetData2().GetParts()),
		Date:        int32(time.Now().Unix()),
		Sizes:       sizes,
	}}

	photos := &mtproto.TLPhotosPhoto{Data2: &mtproto.Photos_Photo_Data{
		Photo: photo.To_Photo(),
		Users: []*mtproto.User{},
	}}

	updateUserPhoto := &mtproto.TLUpdateUserPhoto{Data2: &mtproto.Update_Data{
		UserId: md.UserId,
		Date: int32(time.Now().Unix()),
		Photo: photo2.MakeUserProfilePhoto(uuid, sizes),
		Previous: mtproto.ToBool(false),
	}}
	sync_client.GetSyncClient().PushToUserUpdateShortData(md.UserId, updateUserPhoto.To_Update())

	glog.Infof("photos.uploadProfilePhoto#4f32c098 - reply: %s", logger.JsonDebugData(photos))
	return photos.To_Photos_Photo(), nil
}
