# Telegramd - Unofficial open source telegram server written in golang
> 打造高性能、稳定并且功能完善的开源telegram服务端，建设开源telegram客户端生态系统非官方首选服务！

## Chinese

### 简介
Go语言非官方开源telegram服务端，包括但不限于如下一些特色：

- [mtprotoc](https://github.com/nebulaim/mtprotoc)代码生成器
	- 可自动将tl转换成protobuf协议
	- 自动生成tl二进制数据的的codec代码，可将接收到客户端tl的二进制数据转换成protobuf对象，并通过grpc接入到内部各服务节点处理，这样就可以借助很完善的grpc生态环境来实现我们的系统
- [dalgen](https://github.com/nebulaim/nebula-dal-generator)数据访问层代码生成器
	- 集成了sqlparser解析器，通过可配置的sql自动生成dao代码
	- 代码生成时检查sql语法，极大减少传统手写sql实现的出错几率和手写sql调用的工作量
- 支持可切换的多个服务注册和发现系统
- 集成了grpc的recovery等中间件

### 架构图
![架构图](doc/image/architecture-001.jpeg)

### 文档
[RoadMap](doc/road-map.md)

[Diffie–Hellman key exchange](doc/dh-key-exchange.md)

[Creating an Authorization Key](doc/Creating_an_Authorization_Key.md)

[Mobile Protocol: Detailed Description (v.1.0, DEPRECATED)](doc/Mobile_Protocol-Detailed_Description_v.1.0_DEPRECATED.md)

### 编译和安装

[编译和安装](doc/build.md)

[编译和运行脚本](scripts/build.sh)

[依赖脚本](scripts/prerequisite.sh)

### 配套客户端
#### 官方开源客户端修改适配版本
[Android client for telegramd](https://github.com/nebulaim/TelegramAndroid)

[macOS client for telegramd](https://github.com/nebulaim/TelegramSwift)

[iOS client for telegramd](https://github.com/nebulaim/TelegramiOS)

Web客户端（敬请期待）

桌面客户端（敬请期待）

#### 开源客户端库修改适配版本
tdlib

### TODO
channels, Secret Chats, bots and payments这几大功能还未实现

## English

### Introduce
open source mtproto server implement by golang, which compatible telegram client.

### Install
[Build and install](doc/build.md)

[build](scripts/build.sh)

[prerequisite](scripts/prerequisite.sh)

## Feedback
Please report bugs, concerns, suggestions by issues, or join telegram group [Telegramd](https://t.me/joinchat/D8b0DRJiuH8EcIHNZQmCxQ) to discuss problems around source code.
