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

package model

import (
	"github.com/nebulaim/telegramd/mtproto"
	"github.com/golang/glog"
	"time"
)


//messages.sendMedia#b8d1262b flags:# silent:flags.5?true background:flags.6?true clear_draft:flags.7?true peer:InputPeer reply_to_msg_id:flags.0?int media:InputMedia message:string random_id:long reply_markup:flags.2?ReplyMarkup entities:flags.3?Vector<MessageEntity> = Updates;
//messages.forwardMessages#708e0195 flags:# silent:flags.5?true background:flags.6?true with_my_score:flags.8?true grouped:flags.9?true from_peer:InputPeer id:Vector<int> random_id:Vector<long> to_peer:InputPeer = Updates;
func sendMessageToMessageData(m *mtproto.TLMessagesSendMessage) *mtproto.TLMessage {
	//messages.sendMessage#fa88427a flags:# no_webpage:flags.1?true silent:flags.5?true background:flags.6?true clear_draft:flags.7?true peer:InputPeer reply_to_msg_id:flags.0?int message:string random_id:long reply_markup:flags.2?ReplyMarkup entities:flags.3?Vector<MessageEntity> = Updates;

	//// TODO(@benqi): ???
	//// request.Background
	//// request.NoWebpage
	//// request.ClearDraft
	//message.SetFromId(md.UserId)
	//if peer.PeerType == base.PEER_SELF {
	//	to := &mtproto.TLPeerUser{ Data2: &mtproto.Peer_Data{
	//		UserId: md.UserId,
	//	}}
	//	message.SetToId(to.To_Peer())
	//} else {
	//	message.SetToId(peer.ToPeer())
	//}

	return &mtproto.TLMessage{ Data2: &mtproto.Message_Data{
		Silent:       m.GetSilent(),
		ReplyToMsgId: m.GetReplyToMsgId(),
		Message:      m.GetMessage(),
		ReplyMarkup:  m.GetReplyMarkup(),
		Entities:     m.GetEntities(),
		Date:         int32(time.Now().Unix()),
	}}
}

func sendMediaToMessageData(m *mtproto.TLMessagesSendMedia) *mtproto.TLMessage {
	return &mtproto.TLMessage{ Data2: &mtproto.Message_Data{
		Silent:       m.GetSilent(),
		ReplyToMsgId: m.GetReplyToMsgId(),
		// Media:  m.GetMedia(),
		// Message:      m.GetMessage(),
		ReplyMarkup:  m.GetReplyMarkup(),
		// Entities:     m.GetEntities(),
		Date:         int32(time.Now().Unix()),
	}}
}

func MakeMessageBySendMessage(m mtproto.TLObject) (message *mtproto.TLMessage, err error) {
	switch m.(type) {
	case *mtproto.TLMessagesSendMessage:
		message = sendMessageToMessageData(m.(*mtproto.TLMessagesSendMessage))
	case *mtproto.TLMessagesSendMedia:
		message = sendMediaToMessageData(m.(*mtproto.TLMessagesSendMedia))
	default:
		err = mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_INTERNAL_SERVER_ERROR), "internal server error")
		glog.Error(err)
		return
	}
	return
}

