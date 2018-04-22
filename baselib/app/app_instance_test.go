/*
 * Copyright (c) 2018-present, Yumcoder, LLC.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree.
 */
package app

import (
	"github.com/golang/glog"
	"testing"
)

type NullInstance struct {

}

func (e *NullInstance) Initialize() error {
	glog.Info("null instance initialize...")
	return nil
}

func (e *NullInstance) RunLoop() {
	glog.Info("null run loop...")
	QuitAppInstance()
}

func (e *NullInstance) Destroy() {
	glog.Info("null destroy...")
}

func TestRun(t *testing.T) {
	instance := &NullInstance{}
	DoMainAppInstance(instance)
}

