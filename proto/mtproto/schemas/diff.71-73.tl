102,104c102,104
< //message2 msg_id:long seqno:int bytes:int body:Object = Message2; // parsed manually
< //msg_container#73f1f8dc messages:vector<message2> = MessageContainer; // parsed manually
< //msg_copy#e06046b2 orig_message:Message2 = MessageCopy; // parsed manually, not used - use msg_container
---
> //message msg_id:long seqno:int bytes:int body:Object = Message; // parsed manually
> //msg_container#73f1f8dc messages:vector<message> = MessageContainer; // parsed manually
> //msg_copy#e06046b2 orig_message:Message = MessageCopy; // parsed manually, not used - use msg_container

109c109
< ipPort#d433ad73 ipv4:int port:int = IpPort;
---
> ipPort ipv4:int port:int = IpPort;

162c162
< inputMediaUploadedDocument#e39621fd flags:# file:InputFile thumb:flags.2?InputFile mime_type:string attributes:Vector<DocumentAttribute> caption:string stickers:flags.0?Vector<InputDocument> ttl_seconds:flags.1?int = InputMedia;
---
> inputMediaUploadedDocument#e39621fd flags:# nosound_video:flags.3?true file:InputFile thumb:flags.2?InputFile mime_type:string attributes:Vector<DocumentAttribute> caption:string stickers:flags.0?Vector<InputDocument> ttl_seconds:flags.1?int = InputMedia;

164c164
< inputMediaVenue#2827a81a geo_point:InputGeoPoint title:string address:string provider:string venue_id:string = InputMedia;
---
> inputMediaVenue#c13d1c11 geo_point:InputGeoPoint title:string address:string provider:string venue_id:string venue_type:string = InputMedia;

169c169,170
< inputMediaInvoice#92153685 flags:# title:string description:string photo:flags.0?InputWebDocument invoice:Invoice payload:bytes provider:string start_param:string = InputMedia;
---
> inputMediaInvoice#f4e096c3 flags:# title:string description:string photo:flags.0?InputWebDocument invoice:Invoice payload:bytes provider:string provider_data:DataJSON start_param:string = InputMedia;
> inputMediaGeoLive#7b1a118f geo_point:InputGeoPoint period:int = InputMedia;

221c222
< channel#cb44b1c flags:# creator:flags.0?true left:flags.2?true editor:flags.3?true broadcast:flags.5?true verified:flags.7?true megagroup:flags.8?true restricted:flags.9?true democracy:flags.10?true signatures:flags.11?true min:flags.12?true id:int access_hash:flags.13?long title:string username:flags.6?string photo:ChatPhoto date:int version:int restriction_reason:flags.9?string admin_rights:flags.14?ChannelAdminRights banned_rights:flags.15?ChannelBannedRights = Chat;
---
> channel#450b7115 flags:# creator:flags.0?true left:flags.2?true editor:flags.3?true broadcast:flags.5?true verified:flags.7?true megagroup:flags.8?true restricted:flags.9?true democracy:flags.10?true signatures:flags.11?true min:flags.12?true id:int access_hash:flags.13?long title:string username:flags.6?string photo:ChatPhoto date:int version:int restriction_reason:flags.9?string admin_rights:flags.14?ChannelAdminRights banned_rights:flags.15?ChannelBannedRights participants_count:flags.17?int = Chat;

225c226
< channelFull#17f45fcf flags:# can_view_participants:flags.3?true can_set_username:flags.6?true can_set_stickers:flags.7?true id:int about:string participants_count:flags.0?int admins_count:flags.1?int kicked_count:flags.2?int banned_count:flags.2?int read_inbox_max_id:int read_outbox_max_id:int unread_count:int chat_photo:Photo notify_settings:PeerNotifySettings exported_invite:ExportedChatInvite bot_info:Vector<BotInfo> migrated_from_chat_id:flags.4?int migrated_from_max_id:flags.4?int pinned_msg_id:flags.5?int stickerset:flags.8?StickerSet = ChatFull;
---
> channelFull#76af5481 flags:# can_view_participants:flags.3?true can_set_username:flags.6?true can_set_stickers:flags.7?true hidden_prehistory:flags.10?true id:int about:string participants_count:flags.0?int admins_count:flags.1?int kicked_count:flags.2?int banned_count:flags.2?int read_inbox_max_id:int read_outbox_max_id:int unread_count:int chat_photo:Photo notify_settings:PeerNotifySettings exported_invite:ExportedChatInvite bot_info:Vector<BotInfo> migrated_from_chat_id:flags.4?int migrated_from_max_id:flags.4?int pinned_msg_id:flags.5?int stickerset:flags.8?StickerSet available_min_id:flags.9?int = ChatFull;

238c239
< message#90dddc11 flags:# out:flags.1?true mentioned:flags.4?true media_unread:flags.5?true silent:flags.13?true post:flags.14?true id:int from_id:flags.8?int to_id:Peer fwd_from:flags.2?MessageFwdHeader via_bot_id:flags.11?int reply_to_msg_id:flags.3?int date:int message:string media:flags.9?MessageMedia reply_markup:flags.6?ReplyMarkup entities:flags.7?Vector<MessageEntity> views:flags.10?int edit_date:flags.15?int post_author:flags.16?string = Message;
---
> message#44f9b43d flags:# out:flags.1?true mentioned:flags.4?true media_unread:flags.5?true silent:flags.13?true post:flags.14?true id:int from_id:flags.8?int to_id:Peer fwd_from:flags.2?MessageFwdHeader via_bot_id:flags.11?int reply_to_msg_id:flags.3?int date:int message:string media:flags.9?MessageMedia reply_markup:flags.6?ReplyMarkup entities:flags.7?Vector<MessageEntity> views:flags.10?int edit_date:flags.15?int post_author:flags.16?string grouped_id:flags.17?long = Message;

248c249
< messageMediaVenue#7912b71f geo:GeoPoint title:string address:string provider:string venue_id:string = MessageMedia;
---
> messageMediaVenue#2ec0533f geo:GeoPoint title:string address:string provider:string venue_id:string venue_type:string = MessageMedia;

250a252
> messageMediaGeoLive#7c3c2609 geo:GeoPoint period:int = MessageMedia;

269a272
> messageActionCustomAction#fae69f56 message:string = MessageAction;

342a346
> messages.messagesNotModified#74535f21 count:int = messages.Messages;

355d358
< inputMessagesFilterPhotoVideoDocuments#d95e73bb = MessagesFilter;

365a369,370
> inputMessagesFilterGeo#e7026d0d = MessagesFilter;
> inputMessagesFilterContacts#e062db83 = MessagesFilter;

431a437
> updateChannelAvailableMessages#70db6837 channel_id:int available_min_id:int = Update;

458c464
< config#8df376a4 flags:# phonecalls_enabled:flags.1?true date:int expires:int test_mode:Bool this_dc:int dc_options:Vector<DcOption> chat_size_max:int megagroup_size_max:int forwarded_count_max:int online_update_period_ms:int offline_blur_timeout_ms:int offline_idle_timeout_ms:int online_cloud_timeout_ms:int notify_cloud_delay_ms:int notify_default_delay_ms:int chat_big_size:int push_chat_period_ms:int push_chat_limit:int saved_gifs_limit:int edit_time_limit:int rating_e_decay:int stickers_recent_limit:int stickers_faved_limit:int tmp_sessions:flags.0?int pinned_dialogs_count_max:int call_receive_timeout_ms:int call_ring_timeout_ms:int call_connect_timeout_ms:int call_packet_timeout_ms:int me_url_prefix:string suggested_lang_code:flags.2?string lang_pack_version:flags.2?int disabled_features:Vector<DisabledFeature> = Config;
---
> config#9c840964 flags:# phonecalls_enabled:flags.1?true default_p2p_contacts:flags.3?true date:int expires:int test_mode:Bool this_dc:int dc_options:Vector<DcOption> chat_size_max:int megagroup_size_max:int forwarded_count_max:int online_update_period_ms:int offline_blur_timeout_ms:int offline_idle_timeout_ms:int online_cloud_timeout_ms:int notify_cloud_delay_ms:int notify_default_delay_ms:int chat_big_size:int push_chat_period_ms:int push_chat_limit:int saved_gifs_limit:int edit_time_limit:int rating_e_decay:int stickers_recent_limit:int stickers_faved_limit:int channels_read_media_period:int tmp_sessions:flags.0?int pinned_dialogs_count_max:int call_receive_timeout_ms:int call_ring_timeout_ms:int call_connect_timeout_ms:int call_packet_timeout_ms:int me_url_prefix:string suggested_lang_code:flags.2?string lang_pack_version:flags.2?int disabled_features:Vector<DisabledFeature> = Config;

667a674
> channels.channelParticipantsNotModified#f0173fe9 = channels.ChannelParticipants;

683c690
< inputBotInlineMessageMediaGeo#f4a59de1 flags:# geo_point:InputGeoPoint reply_markup:flags.2?ReplyMarkup = InputBotInlineMessage;
---
> inputBotInlineMessageMediaGeo#c1b15d65 flags:# geo_point:InputGeoPoint period:int reply_markup:flags.2?ReplyMarkup = InputBotInlineMessage;

695c702
< botInlineMessageMediaGeo#3a8fd8b8 flags:# geo:GeoPoint reply_markup:flags.2?ReplyMarkup = BotInlineMessage;
---
> botInlineMessageMediaGeo#b722de65 flags:# geo:GeoPoint period:int reply_markup:flags.2?ReplyMarkup = BotInlineMessage;

702c709
< messages.botResults#ccd3563d flags:# gallery:flags.0?true query_id:long next_offset:flags.1?string switch_pm:flags.2?InlineBotSwitchPM results:Vector<BotInlineResult> cache_time:int = messages.BotResults;
---
> messages.botResults#947ca848 flags:# gallery:flags.0?true query_id:long next_offset:flags.1?string switch_pm:flags.2?InlineBotSwitchPM results:Vector<BotInlineResult> cache_time:int users:Vector<User> = messages.BotResults;

706c713
< messageFwdHeader#fadff4ac flags:# from_id:flags.0?int date:int channel_id:flags.1?int channel_post:flags.2?int post_author:flags.3?string = MessageFwdHeader;
---
> messageFwdHeader#559ebe6d flags:# from_id:flags.0?int date:int channel_id:flags.1?int channel_post:flags.2?int post_author:flags.3?string saved_from_peer:flags.4?Peer saved_from_msg_id:flags.4?int = MessageFwdHeader;

717c724
< messages.botCallbackAnswer#36585ea4 flags:# alert:flags.1?true has_url:flags.3?true message:flags.0?string url:flags.2?string cache_time:int = messages.BotCallbackAnswer;
---
> messages.botCallbackAnswer#36585ea4 flags:# alert:flags.1?true has_url:flags.3?true native_ui:flags.4?true message:flags.0?string url:flags.2?string cache_time:int = messages.BotCallbackAnswer;

819c826
< invoice#c30aa358 flags:# test:flags.0?true name_requested:flags.1?true phone_requested:flags.2?true email_requested:flags.3?true shipping_address_requested:flags.4?true flexible:flags.5?true currency:string prices:Vector<LabeledPrice> = Invoice;
---
> invoice#c30aa358 flags:# test:flags.0?true name_requested:flags.1?true phone_requested:flags.2?true email_requested:flags.3?true shipping_address_requested:flags.4?true flexible:flags.5?true phone_to_provider:flags.6?true email_to_provider:flags.7?true currency:string prices:Vector<LabeledPrice> = Invoice;

849a857,858
> inputPaymentCredentialsApplePay#aa1c39f payment_data:DataJSON = InputPaymentCredentials;
> inputPaymentCredentialsAndroidPay#795667a6 payment_token:DataJSON = InputPaymentCredentials;

905a915
> channelAdminLogEventActionTogglePreHistoryHidden#5f5c95f1 new_value:Bool = ChannelAdminLogEventAction;

919a930,939
> recentMeUrlUnknown#46e1d13d url:string = RecentMeUrl;
> recentMeUrlUser#8dbc3336 url:string user_id:int = RecentMeUrl;
> recentMeUrlChat#a01b22f9 url:string chat_id:int = RecentMeUrl;
> recentMeUrlChatInvite#eb49081d url:string chat_invite:ChatInvite = RecentMeUrl;
> recentMeUrlStickerSet#bc0a57dc url:string set:StickerSetCovered = RecentMeUrl;
> 
> help.recentMeUrls#e0310d7 urls:Vector<RecentMeUrl> chats:Vector<Chat> users:Vector<User> = help.RecentMeUrls;
> 
> inputSingleMedia#5eaa7809 media:InputMedia random_id:long = InputSingleMedia;
> 

995c1015
< messages.getHistory#afa92846 peer:InputPeer offset_id:int offset_date:int add_offset:int limit:int max_id:int min_id:int = messages.Messages;
---
> messages.getHistory#dcbb8260 peer:InputPeer offset_id:int offset_date:int add_offset:int limit:int max_id:int min_id:int hash:int = messages.Messages;

1004c1024
< messages.forwardMessages#708e0195 flags:# silent:flags.5?true background:flags.6?true with_my_score:flags.8?true from_peer:InputPeer id:Vector<int> random_id:Vector<long> to_peer:InputPeer = Updates;
---
> messages.forwardMessages#708e0195 flags:# silent:flags.5?true background:flags.6?true with_my_score:flags.8?true grouped:flags.9?true from_peer:InputPeer id:Vector<int> random_id:Vector<long> to_peer:InputPeer = Updates;

1051,1052c1071,1072
< messages.editMessage#ce91e4ca flags:# no_webpage:flags.1?true peer:InputPeer id:int message:flags.11?string reply_markup:flags.2?ReplyMarkup entities:flags.3?Vector<MessageEntity> = Updates;
< messages.editInlineBotMessage#130c2c85 flags:# no_webpage:flags.1?true id:InputBotInlineMessageID message:flags.11?string reply_markup:flags.2?ReplyMarkup entities:flags.3?Vector<MessageEntity> = Bool;
---
> messages.editMessage#5d1b8dd flags:# no_webpage:flags.1?true stop_geo_live:flags.12?true peer:InputPeer id:int message:flags.11?string reply_markup:flags.2?ReplyMarkup entities:flags.3?Vector<MessageEntity> geo_point:flags.13?InputGeoPoint = Updates;
> messages.editInlineBotMessage#b0e08243 flags:# no_webpage:flags.1?true stop_geo_live:flags.12?true id:InputBotInlineMessageID message:flags.11?string reply_markup:flags.2?ReplyMarkup entities:flags.3?Vector<MessageEntity> geo_point:flags.13?InputGeoPoint = Bool;

1082a1103,1106
> messages.readMentions#f0189d3 peer:InputPeer = messages.AffectedHistory;
> messages.getRecentLocations#249431e2 peer:InputPeer limit:int = messages.Messages;
> messages.sendMultiMedia#2095512f flags:# silent:flags.5?true background:flags.6?true clear_draft:flags.7?true peer:InputPeer reply_to_msg_id:flags.0?int multi_media:Vector<InputSingleMedia> = Updates;
> messages.uploadEncryptedFile#5057c497 peer:InputEncryptedChat file:InputEncryptedFile = EncryptedFile;

1110a1135
> help.getRecentMeUrls#3dc0f114 referer:string = help.RecentMeUrls;

1117c1142
< channels.getParticipants#24d98f92 channel:InputChannel filter:ChannelParticipantsFilter offset:int limit:int = channels.ChannelParticipants;
---
> channels.getParticipants#123e05e9 channel:InputChannel filter:ChannelParticipantsFilter offset:int limit:int hash:int = channels.ChannelParticipants;

1141a1167,1168
> channels.deleteHistory#af369d42 channel:InputChannel max_id:int = Bool;
> channels.togglePreHistoryHidden#eabbb94c channel:InputChannel enabled:Bool = Updates;

1172c1199
< // LAYER 71
---
> // LAYER 73
