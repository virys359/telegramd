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
	"github.com/nebulaim/telegramd/baselib/grpc_util"
	"github.com/nebulaim/telegramd/proto/mtproto"
	"golang.org/x/net/context"
	user2 "github.com/nebulaim/telegramd/biz/core/user"
	"github.com/nebulaim/telegramd/server/nbfs/nbfs_client"
	"time"
)

// users.getFullUser#ca30a5b1 id:InputUser = UserFull;
func (s *UsersServiceImpl) UsersGetFullUser(ctx context.Context, request *mtproto.TLUsersGetFullUser) (*mtproto.UserFull, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("users.getFullUser#ca30a5b1 - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	var user *mtproto.User
	fullUser := mtproto.NewTLUserFull()
	fullUser.SetPhoneCallsAvailable(true)
	fullUser.SetPhoneCallsPrivate(false)
	fullUser.SetAbout("@Benqi")
	fullUser.SetCommonChatsCount(0)

	switch request.GetId().GetConstructor() {
	case mtproto.TLConstructor_CRC32_inputUserSelf:
	    // User
	    //userDO := dao.GetUsersDAO(dao.DB_SLAVE).SelectById(md.UserId)
	    //user := &mtproto.TLUser{ Data2: &mtproto.User_Data{
			//Self:       true,
			//Contact:    true,
			//Id:         userDO.Id,
			//FirstName:  userDO.FirstName,
			//LastName:   userDO.LastName,
			//Username:   userDO.Username,
			//AccessHash: userDO.AccessHash,
			//Phone:      userDO.Phone,
		//}}
		user = user2.GetUserById(md.UserId, md.UserId).To_User()
	    fullUser.SetUser(user)
	    //GetUser()user.To_User())

	    // Link
	    link := &mtproto.TLContactsLink{ Data2: &mtproto.Contacts_Link_Data{
	    	MyLink:	mtproto.NewTLContactLinkContact().To_ContactLink(),
	    	ForeignLink: mtproto.NewTLContactLinkContact().To_ContactLink(),
	    	User: user,
		}}
	    fullUser.SetLink(link.To_Contacts_Link())
	case mtproto.TLConstructor_CRC32_inputUser:
	    inputUser := request.GetId().To_InputUser()
	    // request.Id.Payload.(*mtproto.InputUser_InputUser).InputUser
	    // User
	    //userDO := dao.GetUsersDAO(dao.DB_SLAVE).SelectById(inputUser.GetUserId())
		//user := &mtproto.TLUser{ Data2: &mtproto.User_Data{
			//Self:       md.UserId == inputUser.GetUserId(),
			//Contact:    true,
			//Id:         userDO.Id,
			//FirstName:  userDO.FirstName,
			//LastName:   userDO.LastName,
			//Username:   userDO.Username,
			//AccessHash: userDO.AccessHash,
			//Phone:      userDO.Phone,
		//}}

		user = user2.GetUserById(md.UserId, inputUser.GetUserId()).To_User()
		fullUser.SetUser(user)

	    // Link
		link := &mtproto.TLContactsLink{ Data2: &mtproto.Contacts_Link_Data{
			MyLink:	mtproto.NewTLContactLinkContact().To_ContactLink(),
			ForeignLink: mtproto.NewTLContactLinkContact().To_ContactLink(),
			User: user,
		}}
		fullUser.SetLink(link.To_Contacts_Link())
	case mtproto.TLConstructor_CRC32_inputUserEmpty:
	    // TODO(@benqi): BAD_REQUEST: 400
		err := mtproto.NewRpcError2(mtproto.TLRpcErrorCodes_BAD_REQUEST)
		glog.Error(err)
		return nil, err
	}

	// NotifySettings
	peerNotifySettings := &mtproto.TLPeerNotifySettings{ Data2: &mtproto.PeerNotifySettings_Data{
		ShowPreviews: true,
		MuteUntil:    0,
		Sound:        "default",
	}}

	fullUser.SetNotifySettings(peerNotifySettings.To_PeerNotifySettings())

	photoId := user.GetData2().GetPhoto().GetData2().GetPhotoId()
	// profilePhoto := user.GetData2().GetPhoto()
	// profilePhoto.GetData2().
	// photoId := user2.GetDefaultUserPhotoID(request.GetId().GetData2().GetUserId())
	sizes, _ := nbfs_client.GetPhotoSizeList(photoId)
	// photo2 := photo2.MakeUserProfilePhoto(photoId, sizes)
	photo := &mtproto.TLPhoto{ Data2: &mtproto.Photo_Data{
		Id:          photoId,
		HasStickers: false,
		AccessHash:  photoId, // photo2.GetFileAccessHash(file.GetData2().GetId(), file.GetData2().GetParts()),
		Date:        int32(time.Now().Unix()),
		Sizes:       sizes,
	}}
	fullUser.SetProfilePhoto(photo.To_Photo())
	//time.Now()
	//
	//var profilePhoto *mtproto.UserProfilePhoto
	//photoId := user2.GetDefaultUserPhotoID(request.GetId().GetData2().GetUserId())
	//if photoId == 0 {
	//	profilePhoto =  mtproto.NewTLUserProfilePhotoEmpty().To_UserProfilePhoto()
	//} else {
	//	sizeList, _ := nbfs_client.GetPhotoSizeList(photoId)
	//	profilePhoto = photo.MakeUserProfilePhoto(photoId, sizeList)
	//}
	//fullUser.SetProfilePhoto()

	glog.Infof("users.getFullUser#ca30a5b1 - reply: %s", logger.JsonDebugData(fullUser))
	return fullUser.To_UserFull(), nil
}
