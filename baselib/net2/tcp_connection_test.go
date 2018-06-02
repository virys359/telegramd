/*
 *  Copyright (c) 2018, https://github.com/nebulaim
 *  All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

 // TcpConnection full buffer #73
 // Author: @yumcoder-platform (https://github.com/yumcoder-platform)
 //

package net2

import (
	"bufio"
	"bytes"
	"fmt"
	"sync"
	"testing"
)

type mockCodec struct {
	*bufio.Reader
	*bufio.Writer
	sendCounter sync.WaitGroup
}

func (c *mockCodec) Send(msg interface{}) error {
	fmt.Println("send - ", msg)
	buf := []byte(msg.(string))
	if _, err := c.Writer.Write(buf); err != nil {
		fmt.Println(err)
		return err
	}
	c.Writer.WriteByte('\n')
	c.Writer.Flush()
	c.sendCounter.Done()
	return nil
}

func (c *mockCodec) Receive() (interface{}, error) {
	line, err := c.Reader.ReadString('\n')
	return line, err
}

func (c *mockCodec) Close() error {
	return nil
}

func TestSend(t *testing.T) {

	var w bytes.Buffer
	writer := bufio.NewWriter(&w)
	reader := bufio.NewReader(&w)

	mockCodec := &mockCodec{reader, writer, sync.WaitGroup{}}

	tcpConn := NewTcpConnection(`conn0`, nil, 5, mockCodec, nil)
	sendMsgCount := 100
	for i := 0; i < sendMsgCount; i++ {
		mockCodec.sendCounter.Add(1)
		go func(n int){
			fmt.Println(n)
			err := tcpConn.Send(fmt.Sprintf(`msg%d`, n))
			if err != nil {
				fmt.Println(err)
			}
		}(i)
	}

	mockCodec.sendCounter.Wait()

	for i := 0; i < sendMsgCount; i++ {
		m,_ := reader.ReadString('\n')
		fmt.Println(m)
	}
}
