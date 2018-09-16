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

package model

// TODO(@benqi): 配置中心管理配置
// dcOption#5d8c6cc flags:# ipv6:flags.0?true media_only:flags.1?true tcpo_only:flags.2?true cdn:flags.3?true static:flags.4?true id:int ip_address:string port:int = DcOption;
type DcOption struct {
	Ipv6      bool
	MediaOnly bool
	TcpoOnly  bool
	Cdn       bool
	Static    bool
	Id        int32
	IpAddress string
	Port      int32
	Secret    bool
}

type Config struct {
	PhonecallsEnabled       bool
	DefaultP2pContacts      bool
	PreloadFeaturedStickers bool
	IgnorePhoneEntities     bool
	RevokePmInbox           bool
	BlockedMode             bool
	Date                    int32
	TestMode                bool
	ThisDc                  int32
	DcOptions               []DcOption
	DcTxtDomainName         string
	ChatSizeMax             int32
	MegagroupSizeMax        int32
	ForwardedCountMax       int32
	OnlineUpdatePeriodMs    int32
	OfflineBlurTimeoutMs    int32
	OfflineIdleTimeoutMs    int32
	OnlineCloudTimeoutMs    int32
	NotifyCloudDelayMs      int32
	NotifyDefaultDelayMs    int32
	PushChatPeriodMs        int32
	PushChatLimit           int32
	SavedGifsLimit          int32
	EditTimeLimit           int32
	RevokeTimeLimit         int32
	RevokePmTimeLimit       int32
	RatingEDecay            int32
	StickersRecentLimit     int32
	StickersFavedLimit      int32
	ChannelsReadMediaPeriod int32
	TmpSessions             int32
	PinnedDialogsCountMax   int32
	CallReceiveTimeoutMs    int32
	CallRingTimeoutMs       int32
	CallConnectTimeoutMs    int32
	CallPacketTimeoutMs     int32
	MeUrlPrefix             string
	AutoupdateUrlPrefix     string
	GifSearchUsername       string
	VenueSearchUsername     string
	ImgSearchUsername       string
	StaticMapsProvider      bool
	CaptionLengthMax        int32
	MessageLengthMax        int32
	WebfileDcId             int32
	SuggestedLangCode       string
	LangPackVersion         int32
}

/*
     expires: 1536122938 [INT],
     test_mode: { boolFalse },
     this_dc: 5 [INT],
     dc_options: [ vector<0x0>
       { dcOption
         flags: 0 [INT],
         ipv6: [ SKIPPED BY BIT 0 IN FIELD flags ],
         media_only: [ SKIPPED BY BIT 1 IN FIELD flags ],
         tcpo_only: [ SKIPPED BY BIT 2 IN FIELD flags ],
         cdn: [ SKIPPED BY BIT 3 IN FIELD flags ],
         static: [ SKIPPED BY BIT 4 IN FIELD flags ],
         id: 1 [INT],
         ip_address: "149.154.175.50" [STRING],
         port: 443 [INT],
         secret: [ SKIPPED BY BIT 10 IN FIELD flags ],
       },
       { dcOption
         flags: 16 [INT],
         ipv6: [ SKIPPED BY BIT 0 IN FIELD flags ],
         media_only: [ SKIPPED BY BIT 1 IN FIELD flags ],
         tcpo_only: [ SKIPPED BY BIT 2 IN FIELD flags ],
         cdn: [ SKIPPED BY BIT 3 IN FIELD flags ],
         static: YES [ BY BIT 4 IN FIELD flags ],
         id: 1 [INT],
         ip_address: "149.154.175.50" [STRING],
         port: 443 [INT],
         secret: [ SKIPPED BY BIT 10 IN FIELD flags ],
       },
       { dcOption
         flags: 1 [INT],
         ipv6: YES [ BY BIT 0 IN FIELD flags ],
         media_only: [ SKIPPED BY BIT 1 IN FIELD flags ],
         tcpo_only: [ SKIPPED BY BIT 2 IN FIELD flags ],
         cdn: [ SKIPPED BY BIT 3 IN FIELD flags ],
         static: [ SKIPPED BY BIT 4 IN FIELD flags ],
         id: 1 [INT],
         ip_address: "2001:0b28:f23d:f001:0000:0000:0000:000a" [STRING],
         port: 443 [INT],
         secret: [ SKIPPED BY BIT 10 IN FIELD flags ],
       },
       { dcOption
         flags: 0 [INT],
         ipv6: [ SKIPPED BY BIT 0 IN FIELD flags ],
         media_only: [ SKIPPED BY BIT 1 IN FIELD flags ],
         tcpo_only: [ SKIPPED BY BIT 2 IN FIELD flags ],
         cdn: [ SKIPPED BY BIT 3 IN FIELD flags ],
         static: [ SKIPPED BY BIT 4 IN FIELD flags ],
         id: 2 [INT],
         ip_address: "149.154.167.51" [STRING],
         port: 443 [INT],
         secret: [ SKIPPED BY BIT 10 IN FIELD flags ],
       },
       { dcOption
         flags: 16 [INT],
         ipv6: [ SKIPPED BY BIT 0 IN FIELD flags ],
         media_only: [ SKIPPED BY BIT 1 IN FIELD flags ],
         tcpo_only: [ SKIPPED BY BIT 2 IN FIELD flags ],
         cdn: [ SKIPPED BY BIT 3 IN FIELD flags ],
         static: YES [ BY BIT 4 IN FIELD flags ],
         id: 2 [INT],
         ip_address: "149.154.167.51" [STRING],
         port: 443 [INT],
         secret: [ SKIPPED BY BIT 10 IN FIELD flags ],
       },
       { dcOption
         flags: 1 [INT],
         ipv6: YES [ BY BIT 0 IN FIELD flags ],
         media_only: [ SKIPPED BY BIT 1 IN FIELD flags ],
         tcpo_only: [ SKIPPED BY BIT 2 IN FIELD flags ],
         cdn: [ SKIPPED BY BIT 3 IN FIELD flags ],
         static: [ SKIPPED BY BIT 4 IN FIELD flags ],
         id: 2 [INT],
         ip_address: "2001:067c:04e8:f002:0000:0000:0000:000a" [STRING],
         port: 443 [INT],
         secret: [ SKIPPED BY BIT 10 IN FIELD flags ],
       },
       { dcOption
         flags: 0 [INT],
         ipv6: [ SKIPPED BY BIT 0 IN FIELD flags ],
         media_only: [ SKIPPED BY BIT 1 IN FIELD flags ],
         tcpo_only: [ SKIPPED BY BIT 2 IN FIELD flags ],
         cdn: [ SKIPPED BY BIT 3 IN FIELD flags ],
         static: [ SKIPPED BY BIT 4 IN FIELD flags ],
         id: 3 [INT],
         ip_address: "149.154.175.100" [STRING],
         port: 443 [INT],
         secret: [ SKIPPED BY BIT 10 IN FIELD flags ],
       },
       { dcOption
         flags: 16 [INT],
         ipv6: [ SKIPPED BY BIT 0 IN FIELD flags ],
         media_only: [ SKIPPED BY BIT 1 IN FIELD flags ],
         tcpo_only: [ SKIPPED BY BIT 2 IN FIELD flags ],
         cdn: [ SKIPPED BY BIT 3 IN FIELD flags ],
         static: YES [ BY BIT 4 IN FIELD flags ],
         id: 3 [INT],
         ip_address: "149.154.175.100" [STRING],
         port: 443 [INT],
         secret: [ SKIPPED BY BIT 10 IN FIELD flags ],
       },
       { dcOption
         flags: 1 [INT],
         ipv6: YES [ BY BIT 0 IN FIELD flags ],
         media_only: [ SKIPPED BY BIT 1 IN FIELD flags ],
         tcpo_only: [ SKIPPED BY BIT 2 IN FIELD flags ],
         cdn: [ SKIPPED BY BIT 3 IN FIELD flags ],
         static: [ SKIPPED BY BIT 4 IN FIELD flags ],
         id: 3 [INT],
         ip_address: "2001:0b28:f23d:f003:0000:0000:0000:000a" [STRING],
         port: 443 [INT],
         secret: [ SKIPPED BY BIT 10 IN FIELD flags ],
       },
       { dcOption
         flags: 0 [INT],
         ipv6: [ SKIPPED BY BIT 0 IN FIELD flags ],
         media_only: [ SKIPPED BY BIT 1 IN FIELD flags ],
         tcpo_only: [ SKIPPED BY BIT 2 IN FIELD flags ],
         cdn: [ SKIPPED BY BIT 3 IN FIELD flags ],
         static: [ SKIPPED BY BIT 4 IN FIELD flags ],
         id: 4 [INT],
         ip_address: "149.154.167.91" [STRING],
         port: 443 [INT],
         secret: [ SKIPPED BY BIT 10 IN FIELD flags ],
       },
       { dcOption
         flags: 16 [INT],
         ipv6: [ SKIPPED BY BIT 0 IN FIELD flags ],
         media_only: [ SKIPPED BY BIT 1 IN FIELD flags ],
         tcpo_only: [ SKIPPED BY BIT 2 IN FIELD flags ],
         cdn: [ SKIPPED BY BIT 3 IN FIELD flags ],
         static: YES [ BY BIT 4 IN FIELD flags ],
         id: 4 [INT],
         ip_address: "149.154.167.91" [STRING],
         port: 443 [INT],
         secret: [ SKIPPED BY BIT 10 IN FIELD flags ],
       },
       { dcOption
         flags: 1 [INT],
         ipv6: YES [ BY BIT 0 IN FIELD flags ],
         media_only: [ SKIPPED BY BIT 1 IN FIELD flags ],
         tcpo_only: [ SKIPPED BY BIT 2 IN FIELD flags ],
         cdn: [ SKIPPED BY BIT 3 IN FIELD flags ],
         static: [ SKIPPED BY BIT 4 IN FIELD flags ],
         id: 4 [INT],
         ip_address: "2001:067c:04e8:f004:0000:0000:0000:000a" [STRING],
         port: 443 [INT],
         secret: [ SKIPPED BY BIT 10 IN FIELD flags ],
       },
       { dcOption
         flags: 2 [INT],
         ipv6: [ SKIPPED BY BIT 0 IN FIELD flags ],
         media_only: YES [ BY BIT 1 IN FIELD flags ],
         tcpo_only: [ SKIPPED BY BIT 2 IN FIELD flags ],
         cdn: [ SKIPPED BY BIT 3 IN FIELD flags ],
         static: [ SKIPPED BY BIT 4 IN FIELD flags ],
         id: 4 [INT],
         ip_address: "149.154.166.120" [STRING],
         port: 443 [INT],
         secret: [ SKIPPED BY BIT 10 IN FIELD flags ],
       },
       { dcOption
         flags: 3 [INT],
         ipv6: YES [ BY BIT 0 IN FIELD flags ],
         media_only: YES [ BY BIT 1 IN FIELD flags ],
         tcpo_only: [ SKIPPED BY BIT 2 IN FIELD flags ],
         cdn: [ SKIPPED BY BIT 3 IN FIELD flags ],
         static: [ SKIPPED BY BIT 4 IN FIELD flags ],
         id: 4 [INT],
         ip_address: "2001:067c:04e8:f004:0000:0000:0000:000b" [STRING],
         port: 443 [INT],
         secret: [ SKIPPED BY BIT 10 IN FIELD flags ],
       },
       { dcOption
         flags: 1 [INT],
         ipv6: YES [ BY BIT 0 IN FIELD flags ],
         media_only: [ SKIPPED BY BIT 1 IN FIELD flags ],
         tcpo_only: [ SKIPPED BY BIT 2 IN FIELD flags ],
         cdn: [ SKIPPED BY BIT 3 IN FIELD flags ],
         static: [ SKIPPED BY BIT 4 IN FIELD flags ],
         id: 5 [INT],
         ip_address: "2001:0b28:f23f:f005:0000:0000:0000:000a" [STRING],
         port: 443 [INT],
         secret: [ SKIPPED BY BIT 10 IN FIELD flags ],
       },
       { dcOption
         flags: 16 [INT],
         ipv6: [ SKIPPED BY BIT 0 IN FIELD flags ],
         media_only: [ SKIPPED BY BIT 1 IN FIELD flags ],
         tcpo_only: [ SKIPPED BY BIT 2 IN FIELD flags ],
         cdn: [ SKIPPED BY BIT 3 IN FIELD flags ],
         static: YES [ BY BIT 4 IN FIELD flags ],
         id: 5 [INT],
         ip_address: "91.108.56.140" [STRING],
         port: 443 [INT],
         secret: [ SKIPPED BY BIT 10 IN FIELD flags ],
       },
       { dcOption
         flags: 0 [INT],
         ipv6: [ SKIPPED BY BIT 0 IN FIELD flags ],
         media_only: [ SKIPPED BY BIT 1 IN FIELD flags ],
         tcpo_only: [ SKIPPED BY BIT 2 IN FIELD flags ],
         cdn: [ SKIPPED BY BIT 3 IN FIELD flags ],
         static: [ SKIPPED BY BIT 4 IN FIELD flags ],
         id: 5 [INT],
         ip_address: "91.108.56.145" [STRING],
         port: 443 [INT],
         secret: [ SKIPPED BY BIT 10 IN FIELD flags ],
       },
     ],
     dc_txt_domain_name: "apv2.stel.com" [STRING],
     chat_size_max: 200 [INT],
     megagroup_size_max: 100000 [INT],
     forwarded_count_max: 100 [INT],
     online_update_period_ms: 120000 [INT],
     offline_blur_timeout_ms: 5000 [INT],
     offline_idle_timeout_ms: 30000 [INT],
     online_cloud_timeout_ms: 300000 [INT],
     notify_cloud_delay_ms: 30000 [INT],
     notify_default_delay_ms: 1500 [INT],
     push_chat_period_ms: 60000 [INT],
     push_chat_limit: 2 [INT],
     saved_gifs_limit: 200 [INT],
     edit_time_limit: 172800 [INT],
     revoke_time_limit: 172800 [INT],
     revoke_pm_time_limit: 172800 [INT],
     rating_e_decay: 2419200 [INT],
     stickers_recent_limit: 200 [INT],
     stickers_faved_limit: 5 [INT],
     channels_read_media_period: 604800 [INT],
     tmp_sessions: [ SKIPPED BY BIT 0 IN FIELD flags ],
     pinned_dialogs_count_max: 5 [INT],
     call_receive_timeout_ms: 20000 [INT],
     call_ring_timeout_ms: 90000 [INT],
     call_connect_timeout_ms: 30000 [INT],
     call_packet_timeout_ms: 10000 [INT],
     me_url_prefix: "https://t.me/" [STRING],
     autoupdate_url_prefix: [ SKIPPED BY BIT 7 IN FIELD flags ],
     gif_search_username: "gif" [STRING],
     venue_search_username: "foursquare" [STRING],
     img_search_username: "bing" [STRING],
     static_maps_provider: [ SKIPPED BY BIT 12 IN FIELD flags ],
     caption_length_max: 200 [INT],
     message_length_max: 4096 [INT],
     webfile_dc_id: 4 [INT],
     suggested_lang_code: "en" [STRING],
     lang_pack_version: 69 [INT],
   },

*/
