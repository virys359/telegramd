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

package server

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/nebulaim/telegramd/baselib/app"
	"github.com/nebulaim/telegramd/baselib/grpc_util"
	"github.com/nebulaim/telegramd/baselib/net2"
	"github.com/nebulaim/telegramd/mtproto"
)

type sessionClientCallback interface {
	SendToClientData(*sessionClient, int32, *mtproto.ZProtoMetadata, []*messageData) error
}

func sendDataByConnection(conn *net2.TcpConnection, sessionID uint64, md *mtproto.ZProtoMetadata, buf []byte) error {
	smsg := &mtproto.ZProtoSessionData{
		MTPMessage: &mtproto.MTPRawMessage{
			Payload: buf,
		},
	}
	zmsg := &mtproto.ZProtoMessage{
		SessionId: sessionID,
		Metadata:  md,
		SeqNum:    2,
		Message: &mtproto.ZProtoRawPayload{
			Payload: smsg.Encode(),
		},
	}
	return conn.Send(zmsg)
}

func sendDataByConnID(connID, sessionID uint64, md *mtproto.ZProtoMetadata, buf []byte) error {
	sessionServer, ok := app.GAppInstance.(*SessionServer)
	if !ok {
		err := fmt.Errorf("not use app instance framework")
		glog.Error(err)
		return err
	}
	return sessionServer.SendToClientData(connID, sessionID, md, buf)
}

func getBizRPCClient() (*grpc_util.RPCClient, error) {
	sessionServer, ok := app.GAppInstance.(*SessionServer)
	if !ok {
		err := fmt.Errorf("not use app instance framework")
		glog.Error(err)
		return nil, err
	}
	return sessionServer.bizRpcClient, nil
}

func getNbfsRPCClient() (*grpc_util.RPCClient, error) {
	sessionServer, ok := app.GAppInstance.(*SessionServer)
	if !ok {
		err := fmt.Errorf("not use app instance framework")
		glog.Error(err)
		return nil, err
	}
	return sessionServer.nbfsRpcClient, nil
}
