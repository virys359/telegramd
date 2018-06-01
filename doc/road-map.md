# Telegramd roadmap
## 客户端和测试服务器

## 基础库（baselib）
### 统计和监控
> 监控和统计下沉到底层基础库

- 监控和统计
	- 连接
	- 字节
	- 收发包
	- ...
	
### 补齐单元测试

## 压测和性能优化
### 压测程序
### 优化

## 文档
### 架构文档
### 核心模块设计文档
### telegram.org技术文档翻译

## 接入层
> TODO(@benqi): 接入层使用C++重构，基于nebula基础库

### frontend
- transport支持
	- 官方客户端版本支持（已经支持）
	- abridged version支持（开源客户端库会用到，比如[https://github.com/shelomentsevd/telegramgo](https://github.com/shelomentsevd/telegramgo)）
	- http支持（web客户端会用到）
	
- 连接session连接池
- session连接一致性hash支持（实现）
	- 同一个`auth_key_id`路由到同一台session服务器上
- 是否考虑在frontend层解密？
- session服务发现（已经完成）
- 支持多数据中心（DC）
- frontend触发客户端连接事件后通知session
- `quick_ack_id`实现
- 使用mtproto重构服务间tcp连接zproto协议
- ...
	
### handshake
- handshake握手
- 查询`auth_key_id`和`auth_key`
- 完善流程
	- `server_DH_params_fail`
	- `dh_gen_retry`
	- `dh_gen_fail`
- 实现`destroy_auth_key`

### session
- 各个连接类型
- 消重
- 识别快速重连（暂定1小时？）
- 废弃zproto，合并至mtproto
- 消息可靠性保证
- pub／sub
- 由frontend的客户端连接事件管理sessionClientList
- 锁保护sessionManager
- 使用LRUCache缓存auth_key（已完成）
- sessionClientList添加chann里，接收到的数据由chan处理
- 检查msg_id
- 检查msg_seqno
- `ping_delay_disconnect`消息支持定时关闭连接
- `get_future_salts.num` 限制在(1~64)
- 完善`rpc_drop_answer`
- 完善`destroy_session`
- `new_session_created`
- 解耦redis和dao相关代码

	> If, after this, the server receives a message with an even smaller msg_id within the same session, a similar notification will be generated for this msg_id as well. 
	>
	> message confirm
	
- `msg_container`支持空容器	
- `msg_ack`
- `bad_msg_notification` or `bad_server_salt` error_code
- 消息可靠性保证
- 。。。
 
### proto proxy(协议兼容)
> 是否要一个协议兼容proxy？

## service
### idgen
- snowshake
- seqsvr
- redis

### user
- user service

### status
> 在线状态服务器
> 
> 订阅发布

## biz
> 完善功能

### 使用db中间件
### 不同存储系统适配

### redis
> auth的transaction使用redis
>

### biz_server
- 某些情况可能导致串消息
	> 曾经发现过几次，需要找到原因

- stickers
	> 未实现
- channels
	> 未实现
- End-to-End Encryption, Secret Chats
- bots
- payments

## sync
### redis存储sync消息
### push

## nbfs
> 分布式或使用nfs的文档
> 
> 文件存储系统
> 
 
## relay
> 完善relay详细交互流程以及分布式relay支持

- call_session建立以后要存储或者转发给relay
