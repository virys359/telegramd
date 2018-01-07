#!/bin/sh
SRC_DIR=.
DST_DIR=..

protoc -I. -I../../../../../github.com/nebulaim/telegramd/mtproto/schemas -I=../../../../.. --go_out=plugins=grpc:$DST_DIR/ $SRC_DIR/*.proto

