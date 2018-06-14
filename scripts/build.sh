#!/usr/bin/env bash

# todo(yumcoder) change dc ip
# sed -i '/ipAddress = /c\ipAddress = 127.0.0.1' a.txt
# todo(yumcoder) change folder path for nbfs

docker start mysql-docker redis-docker etcd-docker

telegramd="$GOPATH/src/github.com/nebulaim/telegramd"

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