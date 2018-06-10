/*
 * Copyright (c) 2018-present, Yumcoder, LLC.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree.
 */
package logger

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestJsonDebugData(t *testing.T) {
	data := struct {
		Name string
		Id   int
	}{
		"telegramd",
		1,
	}
	result := string(JsonDebugData(data))
	expected := `{"Name":"telegramd","Id":1}`

	assert.Equal(t, expected, result)
}
