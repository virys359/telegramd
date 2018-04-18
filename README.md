# telegramd

## Chinese

### 简介
Go语言开源mtproto服务器，兼容telegram客户端

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
