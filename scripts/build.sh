#!/usr/bin/env bash

docker start mysql-docker redis-docker etcd-docker

telegramd="$GOPATH/src/github.com/yumcoder-platform/telegramd/"

echo "build frontend ..."
cd ${telegramd}/access/frontend
go get
go build
./frontend &

echo "build auth_key ..."
cd ${telegramd}/access/auth_key
go get
go build
.//auth_key &

echo "build sync ..."
cd ${telegramd}/push/sync
go get
go build
./sync &

echo "build nbfs ..."
cd ${telegramd}/nbfs/nbfs
go get
go build
./nbfs &

echo "build biz_server ..."
cd ${telegramd}/biz_server
go get
go build
./biz_server &

echo "build session ..."
cd ${telegramd}/access/session
go get
go build
./session &

echo "***** wait *****"
wait