/*
 * Copyright (c) 2018-present, Yumcoder, LLC.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree.
 */
package test

import (
	"io"
	"github.com/nebulaim/telegramd/baselib/net2"
	"bufio"
)

type TestProto struct {
}

func (b *TestProto) NewCodec(rw io.ReadWriter) (cc net2.Codec, err error) {
	c := &TestCodec{
		Reader: bufio.NewReader(rw),
		Writer: rw.(io.Writer),
		Closer: rw.(io.Closer),
	}
	return c, nil
}
