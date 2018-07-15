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

package contact

import (
	"time"

	"github.com/nebulaim/telegramd/biz/dal/dao"
	"github.com/nebulaim/telegramd/biz/dal/dataobject"
	"github.com/nebulaim/telegramd/proto/mtproto"
)

//type contactUser struct {
//	userId int32
//	phone string
//	firstName string
//	lastName string
//}

// exclude

type contactData *dataobject.UserContactsDO
type contactLogic int32

func MakeContactLogic(userId int32) contactLogic {
	return contactLogic(userId)
}

func findContaceByPhone(contacts []contactData, phone string) *dataobject.UserContactsDO {
	for _, c := range contacts {
		if c.ContactPhone == phone {
			return c
		}
	}
	return nil
}

// include deleted
func (c contactLogic) GetAllContactList() []contactData {
	doList := dao.GetUserContactsDAO(dao.DB_SLAVE).SelectAllUserContacts(int32(c))
	contactList := make([]contactData, 0, len(doList))
	for index, _ := range doList {
		contactList = append(contactList, &doList[index])
	}
	return contactList
}

// exclude deleted
func (c contactLogic) GetContactList() []contactData {
	doList := dao.GetUserContactsDAO(dao.DB_SLAVE).SelectUserContacts(int32(c))
	contactList := make([]contactData, 0, len(doList))
	for index, _ := range doList {
		contactList = append(contactList, &doList[index])
	}
	return contactList
}

func (c contactLogic) ImportContact(userId int32, phone, firstName, lastName string) bool {
	var needUpdate bool = false

	// TODO(@benqi): phone is me???

	// 我->input
	byMy := dao.GetUserContactsDAO(dao.DB_SLAVE).SelectUserContact(int32(c), userId)
	// input->我
	byInput := dao.GetUserContactsDAO(dao.DB_SLAVE).SelectUserContact(userId, int32(c))

	now := int32(time.Now().Unix())
	if byInput == nil {
		// 我不是input的联系人
		if byMy == nil {
			// input不是我的联系人
			do := &dataobject.UserContactsDO{
				OwnerUserId:      int32(c),
				ContactUserId:    userId,
				ContactPhone:     phone,
				ContactFirstName: firstName,
				ContactLastName:  lastName,
				Mutual:           0,
				Date2:            now,
			}
			do.Id = int32(dao.GetUserContactsDAO(dao.DB_MASTER).Insert(do))
		} else {
			dao.GetUserContactsDAO(dao.DB_MASTER).UpdateContactNameById(firstName, lastName, byMy.Id)
		}
	} else {
		// 我不是input的联系人
		if byMy == nil {
			// input不是我的联系人
			do := &dataobject.UserContactsDO{
				OwnerUserId:      int32(c),
				ContactUserId:    userId,
				ContactPhone:     phone,
				ContactFirstName: firstName,
				ContactLastName:  lastName,
				Mutual:           1,
				Date2:            now,
			}
			do.Id = int32(dao.GetUserContactsDAO(dao.DB_MASTER).Insert(do))
			dao.GetUserContactsDAO(dao.DB_MASTER).UpdateMutual(1, userId, int32(c))
			needUpdate = true
		} else {
			dao.GetUserContactsDAO(dao.DB_MASTER).UpdateContactNameById(firstName, lastName, byMy.Id)
			if byMy.IsDeleted == 1 {
				dao.GetUserContactsDAO(dao.DB_MASTER).UpdateMutual(1, userId, int32(c))
				dao.GetUserContactsDAO(dao.DB_MASTER).UpdateMutual(1, int32(c), userId)
				needUpdate = true
			}
		}
	}

	return needUpdate
}

func (c contactLogic) DeleteContact(deleteId int32, mutual bool) bool {
	// A 删除 B
	// 如果AB is mutual，则BA设置为非mutual

	var needUpdate = false

	dao.GetUserContactsDAO(dao.DB_MASTER).DeleteContacts(int32(c), []int32{deleteId})

	if deleteId != int32(c) && mutual {
		dao.GetUserContactsDAO(dao.DB_MASTER).UpdateMutual(0, deleteId, int32(c))
		needUpdate = true
	}

	return needUpdate
}

//// imported int64, 低32位为InputContact的index， 高32位为userId
//func (c contactLogic) AddContactList(contactList []*mtproto.InputContact) (importedList []int64, retryList []int64) {
//	contacts := c.GetAllContactList()
//	// dao.GetUserContactsDAO(dao.DB_SLAVE).SelectUserContacts(int32(c))
//	for i, v := range contactList {
//		inputContact := v.To_InputPhoneContact()
//		found := findContaceByPhone(contacts, inputContact.GetPhone())
//
//		// TODO(@benqi): ?? popularContact#5ce14175 client_id:long importers:int = PopularContact;
//		if found == nil {
//			// Not found, insert.
//			// Check user exist by phone number
//			// TODO(@benqi): mutual
//			do := &dataobject.UserContactsDO{
//				OwnerUserId:      int32(c),
//				ContactPhone:     inputContact.GetPhone(),
//				ContactFirstName: inputContact.GetFirstName(),
//				ContactLastName:  inputContact.GetLastName(),
//			}
//
//			do.Id = int32(dao.GetUserContactsDAO(dao.DB_MASTER).Insert(do))
//
//			// 低32位为InputContact的index， 高32位为userId
//			importedList = append(importedList, int64(i) | int64(do.Id) << 32)
//		} else {
//			// delete
//			if found.IsDeleted == 1 {
//				// 如果已经删除，则将delete设置为0
//				// update delete = 0
//				dao.GetUserContactsDAO(dao.DB_MASTER).UpdateContactNameById(inputContact.GetFirstName(), inputContact.GetLastName(), found.Id)
//				importedList = append(importedList, int64(i) | int64(found.Id) << 32)
//			} else {
//				if found.ContactFirstName != inputContact.GetFirstName() || found.ContactLastName != inputContact.GetLastName() {
//					// 修改联系人名字
//					dao.GetUserContactsDAO(dao.DB_MASTER).UpdateContactNameById(inputContact.GetFirstName(), inputContact.GetLastName(), found.Id)
//					importedList = append(importedList, int64(i) | int64(found.Id) << 32)
//				} else {
//					retryList = append(retryList, inputContact.GetClientId())
//				}
//			}
//		}
//	}
//
//	return
//}

/////////////////////////////////////////////////////////////////////////////////////////
func (c contactLogic) BlockUser(blockId int32) bool {
	dao.GetUserContactsDAO(dao.DB_MASTER).UpdateBlock(1, int32(c), blockId)
	return true
}

func (c contactLogic) UnBlockUser(blockedId int32) bool {
	dao.GetUserContactsDAO(dao.DB_MASTER).UpdateBlock(0, int32(c), blockedId)
	return true
}

func (c contactLogic) GetBlockedList(offset, limit int32) []*mtproto.ContactBlocked {
	// TODO(@benqi): enable offset
	doList := dao.GetUserContactsDAO(dao.DB_SLAVE).SelectBlockedList(int32(c), limit)
	bockedList := make([]*mtproto.ContactBlocked, 0, len(doList))
	for _, do := range doList {
		blocked := &mtproto.ContactBlocked{
			Constructor: mtproto.TLConstructor_CRC32_contactBlocked,
			Data2: &mtproto.ContactBlocked_Data{
				UserId: do.ContactUserId,
				Date:   do.Date2,
			},
		}
		bockedList = append(bockedList, blocked)
	}
	return bockedList
}

func (c contactLogic) SearchContacts(q string, limit int32) []int32 {
	contactList := c.GetContactList()
	idList := make([]int32, 0, len(contactList)+1)
	// {int32(c)}
	idList = append(idList, int32(c))
	for _, c2 := range contactList {
		idList = append(idList, c2.ContactUserId)
	}

	// TODO(@benqi): 区分大小写

	// 构造模糊查询字符串
	q = "%" + q + "%"
	doList := dao.GetUsersDAO(dao.DB_SLAVE).SearchByQueryNotIdList(q, idList, limit)
	founds := make([]int32, 0, len(doList))
	for _, do := range doList {
		founds = append(founds, do.Id)
	}
	return founds
}

func CheckContactAndMutualByUserId(selfId, contactId int32) (bool, bool) {
	do := dao.GetUserContactsDAO(dao.DB_SLAVE).SelectUserContact(selfId, contactId)
	if do == nil {
		return false, false
	} else {
		return true, do.Mutual == 1
	}
}
