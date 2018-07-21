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

package dialog

/**
  1. peer_user和peer_chat使用user_dialogs存储
  2. channel和super_chat用channel_participants
 */

// dialog#e4def5db flags:#
// 	pinned:flags.2?true
// 	peer:Peer top_message:int
// 	read_inbox_max_id:int
// 	read_outbox_max_id:int
// 	unread_count:int
// 	unread_mentions_count:int
// 	notify_settings:PeerNotifySettings
// 	pts:flags.0?int
// 	draft:flags.1?DraftMessage = Dialog;
type dialogModel struct {

}

func (m *dialogModel) GetUserDialogs() {

}

func (m *dialogModel) GetUserChannels() {

}

