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
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/golang/glog"
	"github.com/nebulaim/telegramd/proto/mtproto"
	base2 "github.com/nebulaim/telegramd/server/nbfs/biz/base"
	"github.com/nebulaim/telegramd/server/nbfs/biz/dal/dataobject"
	"image"
	"math/rand"
	// "os"
	"github.com/nebulaim/telegramd/server/nbfs/biz/core"
	"github.com/nebulaim/telegramd/server/nbfs/biz/dal/dao"
	"io/ioutil"
	"os"
	"time"
)

const (
	kPhotoSizeOriginalType = "0" // client upload original photo
	kPhotoSizeSmallType    = "s"
	kPhotoSizeMediumType   = "m"
	kPhotoSizeXLargeType   = "x"
	kPhotoSizeYLargeType   = "y"
	kPhotoSizeAType        = "a"
	kPhotoSizeBType        = "b"
	kPhotoSizeCType        = "c"

	kPhotoSizeOriginalSize = 0 // client upload original photo
	kPhotoSizeSmallSize    = 90
	kPhotoSizeMediumSize   = 320
	kPhotoSizeXLargeSize   = 800
	kPhotoSizeYLargeSize   = 1280
	kPhotoSizeASize        = 160
	kPhotoSizeBSize        = 320
	kPhotoSizeCSize        = 640

	kPhotoSizeAIndex = 4
)

var sizeList = []int{
	kPhotoSizeOriginalSize,
	kPhotoSizeSmallSize,
	kPhotoSizeMediumSize,
	kPhotoSizeXLargeSize,
	kPhotoSizeYLargeSize,
	kPhotoSizeASize,
	kPhotoSizeBSize,
	kPhotoSizeCSize,
}

func getSizeType(idx int) string {
	switch idx {
	case 0:
		return kPhotoSizeOriginalType
	case 1:
		return kPhotoSizeSmallType
	case 2:
		return kPhotoSizeMediumType
	case 3:
		return kPhotoSizeXLargeType
	case 4:
		return kPhotoSizeYLargeType
	case 5:
		return kPhotoSizeAType
	case 6:
		return kPhotoSizeBType
	case 7:
		return kPhotoSizeCType
	}

	return ""
}

//storage.fileUnknown#aa963b05 = storage.FileType;
//storage.filePartial#40bc6f52 = storage.FileType;
//storage.fileJpeg#7efe0e = storage.FileType;
//storage.fileGif#cae1aadf = storage.FileType;
//storage.filePng#a4f63c0 = storage.FileType;
//storage.filePdf#ae1e508d = storage.FileType;
//storage.fileMp3#528a0677 = storage.FileType;
//storage.fileMov#4b09ebbc = storage.FileType;
//storage.fileMp4#b3cea0e4 = storage.FileType;
//storage.fileWebp#1081464c = storage.FileType;

//func checkIsABC(idx int) bool {
//	if idx ==
//}

//import "github.com/golang/glog"
//

type resizeInfo struct {
	isWidth bool
	size    int
}

func makeResizeInfo(img image.Image) resizeInfo {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()

	if w >= h {
		return resizeInfo{
			isWidth: true,
			size:    w,
		}
	} else {
		return resizeInfo{
			isWidth: false,
			size:    h,
		}
	}
}

func getFileSize(filename string) int32 {
	fi, e := os.Stat(filename)
	if e != nil {
		return 0
	}
	// get the size
	return int32(fi.Size())
}

// TODO(@benqi):
// 	我们未来的图片存储系统可能会按facebook的Haystack论文来实现
// 	mtproto协议也定义了一套自己的文件存储方案，fileLocation#53d69076 dc_id:int volume_id:long local_id:int secret:long = FileLocation;
// 	在这里，我们重新定义mtproto的volume_id和local_id，对应Haystack的key和alternate_key，secret对应cookie
//  在当前简单实现里，volume_id由sonwflake生成，local_id对应于图片类型，secret为access_hash
// TODO(@benqi):
//  参数使用mtproto.File
func UploadPhotoFile(photoId, accessHash int64, filePath, extName string, isABC bool) ([]*mtproto.PhotoSize, error) {
	fileName := core.NBFS_DATA_PATH + filePath

	f, err := os.Open(fileName)
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	defer f.Close()

	sizes := make([]*mtproto.PhotoSize, 0, 4)

	//fileDatas, err := ioutil.ReadFile(fileName)
	//if err != nil {
	//	glog.Error(err)
	//	return nil, err
	//}

	// bufio.Reader{}
	img, err := imaging.Decode(f)
	// img, err := imaging.Decode(bytes.NewReader(fileDatas))
	if err != nil {
		glog.Errorf("Decode %s error: {%v}", fileName, err)
		return nil, err
	}

	imgSz := makeResizeInfo(img)

	vId := base2.NextSnowflakeId()
	for i, sz := range sizeList {
		if i != 0 {
			if isABC {
				if i <= kPhotoSizeAIndex {
					continue
				}
			} else {
				if i > kPhotoSizeAIndex {
					continue
				}
			}
		}
		photoDatasDO := &dataobject.PhotoDatasDO{
			PhotoId:    photoId,
			DcId:       2,
			VolumeId:   vId,
			LocalId:    int32(i),
			AccessHash: rand.Int63(),
			// Bytes:   []byte{0},
			Ext: extName,
		}

		if i == 0 {
			photoDatasDO.Width = int32(img.Bounds().Dx())
			photoDatasDO.Height = int32(img.Bounds().Dy())
			photoDatasDO.FileSize = getFileSize(fileName)
			photoDatasDO.FilePath = filePath
			photoDatasDO.AccessHash = accessHash
		} else {
			var dst *image.NRGBA
			if imgSz.isWidth {
				dst = imaging.Resize(img, sz, 0, imaging.Lanczos)
			} else {
				dst = imaging.Resize(img, 0, sz, imaging.Lanczos)
			}

			photoDatasDO.Width = int32(dst.Bounds().Dx())
			photoDatasDO.Height = int32(dst.Bounds().Dy())

			// imgBuf := base2.MakeBuffer(0, len(fileDatas))
			photoDatasDO.FilePath = fmt.Sprintf("/%s/%d%s", getSizeType(i), vId, extName)
			dstFileName := core.NBFS_DATA_PATH + photoDatasDO.FilePath
			// fmt.Sprintf("%s/%s/%d%s", core.NBFS_DATA_PATH, getSizeType(i), vId, extName)
			err = imaging.Save(dst, dstFileName)
			// .Encode(imgBuf, dst, imaging.JPEG)
			if err != nil {
				glog.Errorf("Encode error: {%v}", err)
				return nil, err
			}
			photoDatasDO.FileSize = getFileSize(dstFileName)
		}

		// photoDatasDO.Bytes = imgBuf.Bytes()
		dao.GetPhotoDatasDAO(dao.DB_MASTER).Insert(photoDatasDO)

		photoSizeData := &mtproto.PhotoSize_Data{
			Type: getSizeType(i),
			W:    photoDatasDO.Width,
			H:    photoDatasDO.Height,
			// Size: int32(len(photoDatasDO.Bytes)),
			Size: photoDatasDO.FileSize,
			Location: &mtproto.FileLocation{
				Constructor: mtproto.TLConstructor_CRC32_fileLocation,
				Data2: &mtproto.FileLocation_Data{
					VolumeId: photoDatasDO.VolumeId,
					LocalId:  int32(i),
					Secret:   photoDatasDO.AccessHash,
					DcId:     photoDatasDO.DcId}}}

		if i == 0 {
			continue
		} else if i == 1 {
			sizes = append(sizes, &mtproto.PhotoSize{
				Constructor: mtproto.TLConstructor_CRC32_photoCachedSize,
				Data2:       photoSizeData})
			// TODO(@benqi): 如上预先存起来
			photoSizeData.Bytes, _ = ioutil.ReadFile(core.NBFS_DATA_PATH + photoDatasDO.FilePath)
			// photoDatasDO.Bytes
		} else {
			sizes = append(sizes, &mtproto.PhotoSize{
				Constructor: mtproto.TLConstructor_CRC32_photoSize,
				Data2:       photoSizeData})
		}
	}

	return sizes, nil
}

////////////////////////////////////////////////////////////////////////////
func GetPhotoSizeList(photoId int64) (sizes []*mtproto.PhotoSize) {
	doList := dao.GetPhotoDatasDAO(dao.DB_SLAVE).SelectListByPhotoId(photoId)
	sizes = make([]*mtproto.PhotoSize, 0, len(doList))
	for i := 1; i < len(doList); i++ {
		sizeData := &mtproto.PhotoSize_Data{
			Type: getSizeType(int(doList[i].LocalId)),
			W:    doList[i].Width,
			H:    doList[i].Height,
			Size: doList[i].FileSize,
			Location: &mtproto.FileLocation{
				Constructor: mtproto.TLConstructor_CRC32_fileLocation,
				Data2: &mtproto.FileLocation_Data{
					VolumeId: doList[i].VolumeId,
					LocalId:  int32(doList[i].LocalId),
					Secret:   doList[i].AccessHash,
					DcId:     doList[i].DcId,
				},
			},
		}

		if i == 1 {
			var filename = core.NBFS_DATA_PATH + doList[i].FilePath
			cacheData, err := ioutil.ReadFile(filename)
			if err != nil {
				glog.Errorf("read file %s error: %v", filename, err)
				sizeData.Bytes = []byte{}
			} else {
				sizeData.Bytes = cacheData
			}
			sizes = append(sizes, &mtproto.PhotoSize{
				Constructor: mtproto.TLConstructor_CRC32_photoCachedSize,
				Data2:       sizeData,
			})
		} else {
			sizes = append(sizes, &mtproto.PhotoSize{
				Constructor: mtproto.TLConstructor_CRC32_photoSize,
				Data2:       sizeData,
			})
		}
	}
	return
}

func GetPhotoFileData(volumeId int64, localId int32, secret int64, offset int32, limit int32) (*mtproto.Upload_File, error) {
	// inputFileLocation#14637196 volume_id:long local_id:int secret:long = InputFileLocation;
	do := dao.GetPhotoDatasDAO(dao.DB_MASTER).SelectByFileLocation(volumeId, localId, secret)
	if do == nil {
		return nil, fmt.Errorf("not found photo <%d, %d, %d>", volumeId, localId, secret)
	}

	if offset > do.FileSize {
		limit = 0
	} else if offset+limit > do.FileSize {
		limit = do.FileSize - offset
	}

	var filename = core.NBFS_DATA_PATH + do.FilePath
	f, err := os.Open(filename)
	if err != nil {
		glog.Error("open ", filename, " error: ", err)
		return nil, err
	}
	defer f.Close()

	bytes := make([]byte, limit)
	_, err = f.ReadAt(bytes, int64(offset))
	if err != nil {
		glog.Error("read file ", filename, " error: ", err)
		return nil, err
	}

	uploadFile := &mtproto.TLUploadFile{Data2: &mtproto.Upload_File_Data{
		Type:  core.MakeStorageFileType(do.Ext),
		Mtime: int32(time.Now().Unix()),
		Bytes: bytes,
	}}
	return uploadFile.To_Upload_File(), nil
}
