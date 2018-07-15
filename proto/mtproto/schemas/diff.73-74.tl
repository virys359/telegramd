525c525
< contacts.found#1aa1f784 results:Vector<Peer> chats:Vector<Chat> users:Vector<User> = contacts.Found;
---
> contacts.found#b3134d9d my_results:Vector<Peer> results:Vector<Peer> chats:Vector<Chat> users:Vector<User> = contacts.Found;
711c711
< exportedMessageLink#1f486803 link:string = ExportedMessageLink;
---
> exportedMessageLink#5dab1af4 link:string html:string = ExportedMessageLink;
858c858
< inputPaymentCredentialsAndroidPay#795667a6 payment_token:DataJSON = InputPaymentCredentials;
---
> inputPaymentCredentialsAndroidPay#ca05d50e payment_token:DataJSON google_transaction_id:string = InputPaymentCredentials;
896c896
< channelAdminRights#5d7ceba5 flags:# change_info:flags.0?true post_messages:flags.1?true edit_messages:flags.2?true delete_messages:flags.3?true ban_users:flags.4?true invite_users:flags.5?true invite_link:flags.6?true pin_messages:flags.7?true add_admins:flags.9?true = ChannelAdminRights;
---
> channelAdminRights#5d7ceba5 flags:# change_info:flags.0?true post_messages:flags.1?true edit_messages:flags.2?true delete_messages:flags.3?true ban_users:flags.4?true invite_users:flags.5?true invite_link:flags.6?true pin_messages:flags.7?true add_admins:flags.9?true manage_call:flags.10?true = ChannelAdminRights;

channelAdminRights#5d7ceba5 flags:# change_info:flags.0?true post_messages:flags.1?true edit_messages:flags.2?true delete_messages:flags.3?true ban_users:flags.4?true invite_users:flags.5?true invite_link:flags.6?true pin_messages:flags.7?true add_admins:flags.9?true manage_call:flags.10?true = ChannelAdminRights;

966,967c966,967
< account.registerDevice#637ea878 token_type:int token:string = Bool;
< account.unregisterDevice#65c55b40 token_type:int token:string = Bool;
---
> account.registerDevice#1389cc token_type:int token:string app_sandbox:Bool other_uids:Vector<int> = Bool;
> account.unregisterDevice#3076c4bf token_type:int token:string other_uids:Vector<int> = Bool;
1159c1159
< channels.exportMessageLink#c846d22d channel:InputChannel id:int = ExportedMessageLink;
---
> channels.exportMessageLink#ceb77163 channel:InputChannel id:int grouped:Bool = ExportedMessageLink;
1199c1199
< // LAYER 73
---
> // LAYER 74



contacts.found#b3134d9d my_results:Vector<Peer> results:Vector<Peer> chats:Vector<Chat> users:Vector<User> = contacts.Found;
exportedMessageLink#5dab1af4 link:string html:string = ExportedMessageLink;
inputPaymentCredentialsAndroidPay#ca05d50e payment_token:DataJSON google_transaction_id:string = InputPaymentCredentials;
channelAdminRights#5d7ceba5 flags:# change_info:flags.0?true post_messages:flags.1?true edit_messages:flags.2?true delete_messages:flags.3?true ban_users:flags.4?true invite_users:flags.5?true invite_link:flags.6?true pin_messages:flags.7?true add_admins:flags.9?true manage_call:flags.10?true = ChannelAdminRights;
account.registerDevice#1389cc token_type:int token:string app_sandbox:Bool other_uids:Vector<int> = Bool;
account.unregisterDevice#3076c4bf token_type:int token:string other_uids:Vector<int> = Bool;
channels.exportMessageLink#ceb77163 channel:InputChannel id:int grouped:Bool = ExportedMessageLink;

// LAYER 74
