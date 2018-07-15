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
[Diffie–Hellman key exchange](doc/dh-key-exchange.md)

[Creating an Authorization Key](doc/Creating_an_Authorization_Key.md)

[Mobile Protocol: Detailed Description (v.1.0, DEPRECATED)](doc/Mobile_Protocol-Detailed_Description_v.1.0_DEPRECATED.md)

[Encrypted CDNs for Speed and Security](doc/cdn.md) [@steedfly](https://github.com/steedfly)翻译
### 编译和安装
#### 简单安装
- 准备
    ```
    mkdir $GOPATH/src/github.com/nebulaim/
    cd $GOPATH/src/github.com/nebulaim/
    git clone https://github.com/nebulaim/telegramd.git
    ```

- 编译代码
    ```
    编译frontend
        cd $GOPATH/src/github.com/nebulaim/telegramd/server/access/frontend
        go get
        go build
    
    编译session
        cd $GOPATH/src/github.com/nebulaim/telegramd/server/access/session
        go get
        go build
    
    编译auth_key
        cd $GOPATH/src/github.com/nebulaim/telegramd/server/access/auth_key
        go get
        go build
        
    编译sync
        cd $GOPATH/src/github.com/nebulaim/telegramd/server/sync
        go get
        go build
    
    编译nbfs
        cd $GOPATH/src/github.com/nebulaim/telegramd/server/nbfs
        go get
        go build
    
    编译biz_server
        cd $GOPATH/src/github.com/nebulaim/telegramd/server/biz_server
        go get
        go build
    ```

- 运行
    ```
    cd $GOPATH/src/github.com/nebulaim/telegramd/server/access/frontend
    ./frontend

    cd $GOPATH/src/github.com/nebulaim/telegramd/server/access/auth_key
    ./auth_key

    cd $GOPATH/src/github.com/nebulaim/telegramd/server/sync
    ./sync
    
    cd $GOPATH/src/github.com/nebulaim/telegramd/server/nbfs
    mkdir /opt/nbfs/0
    mkdir /opt/nbfs/s
    mkdir /opt/nbfs/m
    mkdir /opt/nbfs/x
    mkdir /opt/nbfs/y
    mkdir /opt/nbfs/a
    mkdir /opt/nbfs/b
    mkdir /opt/nbfs/c
    ./nbfs

    cd $GOPATH/src/github.com/nebulaim/telegramd/server/biz_server
    ./biz_server

    cd $GOPATH/src/github.com/nebulaim/telegramd/server/access/session
    ./session
    ```

#### 更多文档
[Build document](doc/build.md)

[Build script](scripts/build.sh)

[Prerequisite script](scripts/prerequisite.sh)

### 配套客户端
#### 官方开源客户端修改适配版本
[Android client for telegramd](https://github.com/nebulaim/TelegramAndroid)

[macOS client for telegramd](https://github.com/nebulaim/TelegramSwift)

[iOS client for telegramd](https://github.com/nebulaim/TelegramiOS)

[tdesktop for telegramd](https://github.com/nebulaim/tdesktop/tree/telegramd)

[webogram for telegramd](https://github.com/nebulaim/webogram)

#### 开源客户端库修改适配版本
tdlib

### TODO
channels, Secret Chats, bots and payments这几大功能还未实现

### 技术交流群
Bug反馈，意见和建议欢迎加入[Telegramd中文技术交流群](https://t.me/joinchat/D8b0DQ-CrvZXZv2dWn310g)讨论。

## English

### Introduce
open source mtproto server implement by golang, which compatible telegram client.

### Install
[Build and install](doc/build.md)

[build](scripts/build.sh)

[prerequisite](scripts/prerequisite.sh)

## Feedback
Please report bugs, concerns, suggestions by issues, or join telegram group [Telegramd](https://t.me/joinchat/D8b0DRJiuH8EcIHNZQmCxQ) to discuss problems around source code.

