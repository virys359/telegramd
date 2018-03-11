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

package delivery

import (
	// "context"
	"github.com/golang/glog"
	// "github.com/nebulaim/telegramd/zproto"
	"github.com/nebulaim/telegramd/grpc_util/service_discovery"
	"github.com/nebulaim/telegramd/grpc_util"
)

type deliveryService struct {
	// client zproto.RPCSyncClient
}

var (
	deliveryInstance = &deliveryService{}
)

func GetDeliveryInstance() *deliveryService {
	return deliveryInstance
}

func InstallDeliveryInstance(discovery *service_discovery.ServiceDiscoveryClientConfig) {
	conn, err := grpc_util.NewRPCClientByServiceDiscovery(discovery)
	_ = conn
	if err != nil {
		glog.Error(err)
		panic(err)
	}

	// deliveryInstance.client = zproto.NewRPCSyncClient(conn)
}

func (d *deliveryService) DeliveryUpdates(authKeyId, sessionId, netlibSessionId int64, sendtoUserIdList []int32, rawData []byte) (err error) {
	//delivery := &zproto.DeliveryUpdatesToUsers{}
	//delivery.MyAuthKeyId = authKeyId
	//delivery.MySessionId = sessionId
	//delivery.MyNetlibSessionId = netlibSessionId
	//delivery.SendtoUserIdList = sendtoUserIdList
	//delivery.RawData = rawData
	//
	//glog.Infof("DeliveryUpdates - delivery: %v", delivery)
	//_, err = d.client.DeliveryUpdates(context.Background(), delivery)
	return
}

func (d *deliveryService) DeliveryUpdatesNotMe(authKeyId, sessionId, netlibSessionId int64, sendtoUserIdList []int32, rawData []byte) (err error) {
	//delivery := &zproto.DeliveryUpdatesToUsers{}
	//delivery.MyAuthKeyId = authKeyId
	//delivery.MySessionId = sessionId
	//delivery.MyNetlibSessionId = netlibSessionId
	//delivery.SendtoUserIdList = sendtoUserIdList
	//delivery.RawData = rawData
	//
	//glog.Infof("DeliveryUpdatesNotMe - delivery: %v", delivery)
	//_, err = d.client.DeliveryUpdatesNotMe(context.Background(), delivery)
	return
}

//func (d *deliveryService) DeliveryUpdates2(authKeyId, sessionId, netlibSessionId int64, pushDatas []*zproto.PushUpdates) (err error) {
//	request := &zproto.UpdatesRequest{
//		SenderAuthKeyId:       authKeyId,
//		SenderSessionId:       sessionId,
//		SenderNetlibSessionId: netlibSessionId,
//		PushDatas:             pushDatas,
//	}
//	glog.Infof("DeliveryUpdates2 - delivery: %v", request)
//	_, err = d.client.DeliveryUpdates2(context.Background(), request)
//	return
//}

//DeliveryUpdateShortMessage(ctx context.Context, in *UpdateShortMessageRequest, opts ...grpc.CallOption) (*DeliveryRsp, error)
//DeliveryUpdatShortChatMessage(ctx context.Context, in *UpdatShortChatMessageRequest, opts ...grpc.CallOption) (*DeliveryRsp, error)
//DeliveryUpdates2(ctx context.Context, in *UpdatesRequest, opts ...grpc.CallOption) (*DeliveryRsp, error)
