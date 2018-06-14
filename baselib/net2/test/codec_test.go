/*
 * Copyright (c) 2018-present, Yumcoder, LLC.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree.
 */
package test

import (
	"bufio"
	"io"
	"fmt"
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
	//c.Closer.Close()
	return nil
}

func (c *TestCodec) ClearSendChan(i <-chan interface{}){
	fmt.Println(`______________________________ClearSendChan_________________`)
}

