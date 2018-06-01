# telegramd

## Chinese

### 简介
Go语言开源telegram服务器，兼容telegram客户端，一些特色：

- 通过 mtprotoc 自动将 tl 转换成protobuf 协议，并生成 tl 的 codec 代码，将客户端收到的 tl 的二进制数据转换成 protobuf 对象 后，通过 grpc 接入到内部各服务节点处理，这样就可以利用很完善的 grpc 生态环境来实现我们 的系统。
- dalgen 数据访问层代码生成器(github.com/nebulaim/nebula-dal-generator):集成了 sqlparser 解析器，从配置文件里读入每条 sql，并为每条 sql 生成一个 dao 方法，而且生成时可 以预先检查 sql 语法，可以极大减少手写 sql 的出错几率以及手写 sql 的工作量。
- 基于 grpc 集成了可替换的服务注册和发现系统(当前代码库里使用了 etcd)
- 集成了 grpc 的 recovery 等中间件，使得业务服务器尽可能少地关注各种异常处理
- 目前已经实现了 handshake、auth、contacts、updates、dialogs、help 等的基本功能，以及发送单聊消息、创建和修改群组、发送群聊消息和发送图片消息等核心功能。
  
最终目标是打造一个高性能、稳定并且功能完善的 telegram 服务端，能让整个开源 telegram 客户端生 态系统除了官方服务之外能有多一个选择。

### Road map
[road map](./doc/reoad-map.md)

### 编译

#### 下载代码

    mkdir $GOPATH/src/github.com/nebulaim/
    cd $GOPATH/src/github.com/nebulaim/
    git clone https://github.com/nebulaim/telegramd.git

#### 编译代码

编译frontend

    cd $GOPATH/src/github.com/nebulaim/telegramd/access/frontend
    go get
    go build

编译session

    cd $GOPATH/src/github.com/nebulaim/telegramd/access/session
    go get
    go build

编译auth_key

    cd $GOPATH/src/github.com/nebulaim/telegramd/access/auth_key
    go get
    go build
    
编译sync

    cd $GOPATH/src/github.com/nebulaim/telegramd/push/sync
    go get
    go build

编译nbfs

    cd $GOPATH/src/github.com/nebulaim/telegramd/nbfs/nbfs
    go get
    go build

编译biz_server

    cd $GOPATH/src/github.com/nebulaim/telegramd/biz_server
    go get
    go build


### 运行

    cd $GOPATH/src/github.com/nebulaim/telegramd/access/frontend
    ./frontend

    cd $GOPATH/src/github.com/nebulaim/telegramd/access/auth_key
    ./auth_key

    cd $GOPATH/src/github.com/nebulaim/telegramd/push/sync
    ./sync
    
    cd $GOPATH/src/github.com/nebulaim/telegramd/nbfs/nbfs
    mkdir /opt/nbfs/0
    mkdir /opt/nbfs/s
    mkdir /opt/nbfs/m
    mkdir /opt/nbfs/x
    mkdir /opt/nbfs/y
    mkdir /opt/nbfs/a
    mkdir /opt/nbfs/b
    mkdir /opt/nbfs/c
    ./nbfs

    cd $GOPATH/src/github.com/nebulaim/telegramd/biz_server
    ./biz_server

    cd $GOPATH/src/github.com/nebulaim/telegramd/access/session
    ./session

## English

### Introduce
open source mtproto server implement by golang, which compatible telegram client.

### Compile

## Feedback
Please report bugs, concerns, suggestions by issues, or join telegram group [Telegramd](https://t.me/joinchat/D8b0DRJiuH8EcIHNZQmCxQ) to discuss problems around source code.
