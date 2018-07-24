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

package photo

import (
	// "github.com/nebulaim/telegramd/biz/dal/dao"
	"github.com/nebulaim/telegramd/proto/mtproto"
)

const (
	PHOTO_SIZE_ORIGINAL     = "0" // client upload original photo
	PHOTO_SIZE_SMALL_TYPE   = "s"
	PHOTO_SIZE_MEDIUMN_TYPE = "m"
	PHOTO_SIZE_XLARGE_TYPE  = "x"
	PHOTO_SIZE_YLARGE_TYPE  = "y"
	PHOTO_SIZE_A_TYPE       = "a"
	PHOTO_SIZE_B_TYPE       = "b"
	PHOTO_SIZE_C_TYPE       = "c"

	PHOTO_SIZE_SMALL_SIZE   = 90
	PHOTO_SIZE_MEDIUMN_SIZE = 320
	PHOTO_SIZE_XLARGE_SIZE  = 800
	PHOTO_SIZE_YLARGE_SIZE  = 1280
	PHOTO_SIZE_A_SIZE       = 160
	PHOTO_SIZE_B_SIZE       = 320
	PHOTO_SIZE_C_SIZE       = 640
)

/*
	inputFile#f52ff27f id:long parts:int name:string md5_checksum:string = InputFile;
	inputFileBig#fa4f0bb5 id:long parts:int name:string = InputFile;

	inputPhotoEmpty#1cd7bf0d = InputPhoto;
	inputPhoto#fb95c6c4 id:long access_hash:long = InputPhoto;

	photoEmpty#2331b22d id:long = Photo;
	photo#9288dd29 flags:# has_stickers:flags.0?true id:long access_hash:long date:int sizes:Vector<PhotoSize> = Photo;

	fileLocationUnavailable#7c596b46 volume_id:long local_id:int secret:long = FileLocation;
	fileLocation#53d69076 dc_id:int volume_id:long local_id:int secret:long = FileLocation;

	photoSizeEmpty#e17e23c type:string = PhotoSize;
	photoSize#77bfb61b type:string location:FileLocation w:int h:int size:int = PhotoSize;
	photoCachedSize#e9a734fa type:string location:FileLocation w:int h:int bytes:bytes = PhotoSize;


	photos.photos#8dca6aa5 photos:Vector<Photo> users:Vector<User> = photos.Photos;
	photos.photosSlice#15051f54 count:int photos:Vector<Photo> users:Vector<User> = photos.Photos;

	photos.photo#20212ca8 photo:Photo users:Vector<User> = photos.Photo;

	upload.file#96a18d5 type:storage.FileType mtime:int bytes:bytes = upload.File;
	upload.fileCdnRedirect#ea52fe5a dc_id:int file_token:bytes encryption_key:bytes encryption_iv:bytes cdn_file_hashes:Vector<CdnFileHash> = upload.File;

	photos.updateProfilePhoto#f0bb5152 id:InputPhoto = UserProfilePhoto;
	photos.uploadProfilePhoto#4f32c098 file:InputFile = photos.Photo;
	photos.deletePhotos#87cf7f2f id:Vector<InputPhoto> = Vector<long>;
	photos.getUserPhotos#91cd32a8 user_id:InputUser offset:int max_id:long limit:int = photos.Photos;

	userProfilePhotoEmpty#4f11bae1 = UserProfilePhoto;
	userProfilePhoto#d559d8c8 photo_id:long photo_small:FileLocation photo_big:FileLocation = UserProfilePhoto;
*/

/*
	storage.fileUnknown#aa963b05 = storage.FileType;
	storage.filePartial#40bc6f52 = storage.FileType;
	storage.fileJpeg#7efe0e = storage.FileType;
	storage.fileGif#cae1aadf = storage.FileType;
	storage.filePng#a4f63c0 = storage.FileType;
	storage.filePdf#ae1e508d = storage.FileType;
	storage.fileMp3#528a0677 = storage.FileType;
	storage.fileMov#4b09ebbc = storage.FileType;
	storage.fileMp4#b3cea0e4 = storage.FileType;
	storage.fileWebp#1081464c = storage.FileType;
*/

//var sizeList = []int{
//	PHOTO_SIZE_SMALL_SIZE,
//	PHOTO_SIZE_SMALL_SIZE,
//	PHOTO_SIZE_MEDIUMN_SIZE,
//	PHOTO_SIZE_XLARGE_SIZE,
//	PHOTO_SIZE_YLARGE_SIZE,
//	PHOTO_SIZE_A_SIZE,
//	PHOTO_SIZE_B_SIZE,
//	PHOTO_SIZE_C_SIZE,
//}
//
//func getSizeType(idx int) string {
//	switch idx {
//	case 0:
//		return PHOTO_SIZE_SMALL_TYPE
//	case 1:
//		return PHOTO_SIZE_MEDIUMN_TYPE
//	case 2:
//		return PHOTO_SIZE_XLARGE_TYPE
//	case 3:
//		return PHOTO_SIZE_YLARGE_TYPE
//	case 4:
//		return PHOTO_SIZE_A_TYPE
//	case 5:
//		return PHOTO_SIZE_B_TYPE
//	case 6:
//		return PHOTO_SIZE_C_TYPE
//	}
//
//	return ""
//}
//
//type resizeInfo struct {
//	isWidth bool
//	size int
//}
//
//func MakeResizeInfo(img image.Image) resizeInfo {
//	w := img.Bounds().Dx()
//	h := img.Bounds().Dy()
//
//	if w >= h {
//		return resizeInfo{
//			isWidth: true,
//			size: w,
//		}
//	} else {
//		return resizeInfo{
//			isWidth: false,
//			size : h,
//		}
//	}
//}

//////////////////////////////////////////////////////////////////////////////
//func GetPhotoSizeList(photoId int64) (sizes []*mtproto.PhotoSize) {
//	doList := dao.GetPhotoDatasDAO(dao.DB_SLAVE).SelectListByPhotoId(photoId)
//	sizes = make([]*mtproto.PhotoSize, 0, len(doList))
//	for _, do := range doList {
//		sizeData := &mtproto.PhotoSize_Data{
//			Type: getSizeType(int(do.LocalId)),
//			W:    do.Width,
//			H:    do.Height,
//			Size: int32(len(do.Bytes)),
//			Location: &mtproto.FileLocation{
//				Constructor: mtproto.TLConstructor_CRC32_fileLocation,
//				Data2: &mtproto.FileLocation_Data{
//					VolumeId: do.VolumeId,
//					LocalId:  int32(do.LocalId),
//					Secret:   do.AccessHash,
//					DcId: 	do.DcId,
//				},
//			},
//		}
//
//		if do.LocalId== 0 {
//			sizes = append(sizes, &mtproto.PhotoSize{
//				Constructor: mtproto.TLConstructor_CRC32_photoCachedSize,
//				Data2:       sizeData,})
//			sizeData.Bytes = do.Bytes
//		} else {
//			sizes = append(sizes, &mtproto.PhotoSize{
//				Constructor: mtproto.TLConstructor_CRC32_photoSize,
//				Data2:       sizeData,})
//		}
//	}
//	return
//}

//// TODO(@benqi):
//// 	我们未来的图片存储系统可能会按facebook的Haystack论文来实现
//// 	mtproto协议也定义了一套自己的文件存储方案，fileLocation#53d69076 dc_id:int volume_id:long local_id:int secret:long = FileLocation;
//// 	在这里，我们重新定义mtproto的volume_id和local_id，对应Haystack的key和alternate_key，secret对应cookie
////  在当前简单实现里，volume_id由sonwflake生成，local_id对应于图片类型，secret为access_hash
//// TODO(@benqi):
////  参数使用mtproto.File
//func UploadPhoto(userId int32, photoId, fileId int64, parts int32, name, md5Checksum string) ([]*mtproto.PhotoSize, error) {
//	sizes := make([]*mtproto.PhotoSize, 0, 4)
//
//	// 图片压缩和处理
//	// ext := filepath.Ext(name)
//
//	filesDO := dao.GetFilesDAO(dao.DB_MASTER).SelectByIDAndParts(fileId, parts)
//	if filesDO == nil {
//		return nil, fmt.Errorf("File exists: id = %d, parts = %d", fileId, parts)
//	}
//
//	// check md5Checksum, big file's md5 is empty
//	if md5Checksum != "" && md5Checksum != filesDO.Md5Checksum {
//		return nil, fmt.Errorf("Invalid md5Checksum: md5Checksum = %s, but filesDO = {%v}", md5Checksum, filesDO)
//	}
//
//	// select file data
//	filePartsDOList := dao.GetFilePartsDAO(dao.DB_MASTER).SelectFileParts(fileId)
//	fileDatas := []byte{}
//
//	for _, p := range filePartsDOList {
//		fileDatas = append(fileDatas, p.Bytes...)
//	}
//
//	// bufio.Reader{}
//	img, err := imaging.Decode(bytes.NewReader(fileDatas))
//	if err != nil {
//		glog.Errorf("Decode error: {%v}", err)
//		return nil, err
//	}
//
//	imgSz := MakeResizeInfo(img)
//
//	vId := base2.NextSnowflakeId()
//	for i, sz := range sizeList {
//		photoDatasDO := &dataobject.PhotoDatasDO{
//			PhotoId: photoId,
//			DcId: 2,
//			VolumeId: vId,
//			LocalId: int32(i),
//			AccessHash: rand.Int63(),
//		}
//
//		var dst *image.NRGBA
//		if imgSz.isWidth {
//			dst = imaging.Resize(img, sz, 0, imaging.Lanczos)
//		} else {
//			dst = imaging.Resize(img, 0, sz, imaging.Lanczos)
//		}
//
//		photoDatasDO.Width = int32(dst.Bounds().Dx())
//		photoDatasDO.Height = int32(dst.Bounds().Dy())
//		imgBuf := helper.MakeBuffer(0, len(fileDatas))
//		err = imaging.Encode(imgBuf, dst, imaging.JPEG)
//		if err != nil {
//			glog.Errorf("Encode error: {%v}", err)
//			return nil, err
//		}
//
//		photoDatasDO.Bytes = imgBuf.Bytes()
//		dao.GetPhotoDatasDAO(dao.DB_MASTER).Insert(photoDatasDO)
//
//		photoSizeData := &mtproto.PhotoSize_Data{
//			Type: getSizeType(i),
//			W:    photoDatasDO.Width,
//			H:    photoDatasDO.Height,
//			Size: int32(len(photoDatasDO.Bytes)),
//			Location: &mtproto.FileLocation{
//				Constructor: mtproto.TLConstructor_CRC32_fileLocation,
//				Data2: &mtproto.FileLocation_Data{
//					VolumeId: photoDatasDO.VolumeId,
//					LocalId:  int32(i),
//					Secret:   photoDatasDO.AccessHash,
//					DcId: 	photoDatasDO.DcId}}}
//
//		if i== 0 {
//			sizes = append(sizes, &mtproto.PhotoSize{
//				Constructor: mtproto.TLConstructor_CRC32_photoCachedSize,
//				Data2:       photoSizeData,})
//			photoSizeData.Bytes = photoDatasDO.Bytes
//		} else {
//			sizes = append(sizes, &mtproto.PhotoSize{
//				Constructor: mtproto.TLConstructor_CRC32_photoSize,
//				Data2:       photoSizeData,})
//		}
//	}
//
//	return sizes, nil
//}
//
//func GetFileAccessHash(fileId int64, fileParts int32) int64 {
//	do := dao.GetFilesDAO(dao.DB_MASTER).SelectByIDAndParts(fileId, fileParts)
//	if do == nil {
//		return 0
//	} else {
//		return do.AccessHash
//	}
//}

func MakeUserProfilePhoto(photoId int64, sizes []*mtproto.PhotoSize) *mtproto.UserProfilePhoto {
	if len(sizes) == 0 {
		return mtproto.NewTLUserProfilePhotoEmpty().To_UserProfilePhoto()
	}

	// TODO(@benqi): check PhotoSize is photoSizeEmpty
	photo := &mtproto.TLUserProfilePhoto{Data2: &mtproto.UserProfilePhoto_Data{
		PhotoId:    photoId,
		PhotoSmall: sizes[0].GetData2().GetLocation(),
		PhotoBig:   sizes[len(sizes)-1].GetData2().GetLocation(),
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
		PhotoBig:   sizes[len(sizes)-1].GetData2().GetLocation(),
	}}

	return photo.To_ChatPhoto()
}
