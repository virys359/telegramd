package server

import (
	"testing"
	"github.com/nebulaim/telegramd/mtproto"
	"fmt"
	"github.com/gogo/protobuf/proto"
)

func TestReflect(t *testing.T) {
	req := &mtproto.ConnectToSessionServerReq{}
	fmt.Println(proto.MessageName(req))
	// m, _ = protoToRawPayload(req)
}
