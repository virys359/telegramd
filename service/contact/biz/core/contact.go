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

package core

import (
	"github.com/nebulaim/telegramd/baselib/mysql_client"
	"fmt"
	"github.com/golang/glog"
	"github.com/nebulaim/telegramd/service/contact/biz/dal/dao/mysql_dao"
)

// var modelInstance *contactModel

type contactsDAO struct {
	*mysql_dao.UnregisteredContactsDAO
	*mysql_dao.UserContactsDAO
	*mysql_dao.PopularContactsDAO
}

type ContactModel struct {
	dao *contactsDAO
}

func InitContactModel(dbName string) (*ContactModel, error) {
	// mysql_dao.UnregisteredContactsDAO{}
	dbClient := mysql_client.GetMysqlClient(dbName)
	if dbClient == nil {
		err := fmt.Errorf("invalid dbName: %s", dbName)
		glog.Error(err)
		return nil, err
	}

	m := &ContactModel{dao: &contactsDAO{
		UnregisteredContactsDAO: mysql_dao.NewUnregisteredContactsDAO(dbClient),
		UserContactsDAO:         mysql_dao.NewUserContactsDAO(dbClient),
		PopularContactsDAO:      mysql_dao.NewPopularContactsDAO(dbClient),
	}}
	return m, nil
}

type inputContactData struct {
	userId    int32
	phone     string
	firstName string
	lastName  string
}

type importedContactData struct {
	userId        int32
	importers     int32
	mutualUpdated bool
}

func (m *ContactModel) ImportContacts(selfUserId int32, contacts []*inputContactData) []*importedContactData {
	if len(contacts) == 0 {
		glog.Errorf("phoneContacts not empty.")
		return []*importedContactData{}
		// return
	}


	logic := MakeContactLogic(selfUserId)
	if len(contacts) == 1 {
		return []*importedContactData{logic.importContact(contacts[0])}
	} else {
		// sync phone book
		return logic.importContacts(contacts)
	}
}

//func (m *ContactModel) GetPopularContacts(selfUserId int32, phones []string) {
//	logic := MakeContactLogic(selfUserId)
//	logic.getPopularContacts(phones)
//}
//

type deleteResult struct {
	userId int32
	state  int32
}

func (m *ContactModel) DeleteContact(selfUserId, contactUserId int32) *deleteResult {
	logic := MakeContactLogic(selfUserId)
	return logic.deleteContact(contactUserId)
}

func (m *ContactModel) DeleteContacts(selfUserId int32, contactUserIdList []int32) []*deleteResult {
	logic := MakeContactLogic(selfUserId)
	return logic.deleteContacts(contactUserIdList)
}


func (m *ContactModel) BlockUser(selfUserId, id int32) bool {
	logic := MakeContactLogic(selfUserId)
	return logic.BlockUser(id)
}

func (m *ContactModel) UnBlockUser(selfUserId, id int32) bool {
	logic := MakeContactLogic(selfUserId)
	return logic.UnBlockUser(id)
}
