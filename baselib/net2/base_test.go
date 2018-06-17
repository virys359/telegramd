/*
 * Copyright (c) 2018-present, Yumcoder, LLC.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree.
 */
package net2

import (
	"bufio"
	"io"
	"github.com/golang/glog"
)

type TestCodec struct {
	*bufio.Reader
	io.Writer
	io.Closer
	mt string
}

func (c *TestCodec) Send(msg interface{}) error {
	buf := []byte(msg.(string))
	if _, err := c.Writer.Write(buf); err != nil {
		return err
	}

	return nil
}

func (c *TestCodec) Receive() (interface{}, error) {
	line, err := c.Reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	return line, err
}

func (c *TestCodec) Close() error {
	return c.Closer.Close()
}

func (c *TestCodec) ClearSendChan(ic <-chan interface{}){
	glog.Info(`TestCodec ClearSendChan, `, ic)
}

//////////////////////////////////////////////////////////////////////////////////////////
type TestProto struct {
}

func (b *TestProto) NewCodec(rw io.ReadWriter) (cc Codec, err error) {
	c := &TestCodec{
		Reader: bufio.NewReader(rw),
		Writer: rw.(io.Writer),
		Closer: rw.(io.Closer),
	}
	return c, nil
}

