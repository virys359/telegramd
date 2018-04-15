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

package app

import (
	"os/signal"
	"syscall"
	"github.com/golang/glog"
	"os"
	"flag"
)

var GAppInstance AppInstance

func init() {
	flag.Parse()
	flag.Set("alsologtostderr", "true")
	flag.Set("log_dir", "false")
}

type AppInstance interface {
	Initialize() error
	RunLoop()
	Destroy()
}

var ch = make(chan os.Signal, 1)

func DoMainAppInstance(insance AppInstance) {
	if insance == nil {
		// panic("instance is nil!!!!")
		glog.Errorf("instance is nil, will exit.")
		return
	}

	// global
	GAppInstance = insance

	glog.Info("instance initialize...")
	err := insance.Initialize()
	if err != nil {
		glog.Infof("instance initialize error: {%v}", err)
		return
	}

	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)

	glog.Info("instance run_loop...")
	go insance.RunLoop()

	// fmt.Printf("%d", os.Getpid())
	glog.Info("Wait quit...")

	s2 := <-ch
	if i, ok := s2.(syscall.Signal); ok {
		glog.Infof("instance recv os.Exit(%d) signal...", i)
	} else {
		glog.Info("instance exit...", i)
	}

	insance.Destroy()
	glog.Info("instance quited!")
}

func QuitAppInstance() {
	notifier := make(chan os.Signal, 1)
	signal.Stop(notifier)
}

// syscall.SIGTERM