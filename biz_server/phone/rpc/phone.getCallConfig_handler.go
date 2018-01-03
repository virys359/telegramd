/*
 *  Copyright (c) 2017, https://github.com/nebulaim
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

package rpc

import (
	"github.com/golang/glog"
	"github.com/nebulaim/telegramd/base/logger"
	"github.com/nebulaim/telegramd/grpc_util"
	"github.com/nebulaim/telegramd/mtproto"
	"golang.org/x/net/context"
	"encoding/json"
)

/*
 "dataJSON":
 {
    "audio_frame_size": 60,
    "jitter_min_delay_20": 6,
    "jitter_min_delay_40": 4,
    "jitter_min_delay_60": 2,
    "jitter_max_delay_20": 25,
    "jitter_max_delay_40": 15,
    "jitter_max_delay_60": 10,
    "jitter_max_slots_20": 50,
    "jitter_max_slots_40": 30,
    "jitter_max_slots_60": 20,
    "jitter_losses_to_reset": 20,
    "jitter_resync_threshold": 0.5,
    "audio_congestion_window": 1024,
    "audio_max_bitrate": 20000,
    "audio_max_bitrate_edge": 16000,
    "audio_max_bitrate_gprs": 8000,
    "audio_max_bitrate_saving": 8000,
    "audio_init_bitrate": 16000,
    "audio_init_bitrate_edge": 8000,
    "audio_init_bitrate_gprs": 8000,
    "audio_init_bitrate_saving": 8000,
    "audio_bitrate_step_incr": 1000,
    "audio_bitrate_step_decr": 1000,
    "use_system_ns": true,
    "use_system_aec": true
 }
 */

type callConfigDataJSON struct {
	audioFrameSize         int     `json:"audio_frame_size"`
	jitterMinDelay20       int     `json:"jitter_min_delay_20"`
	jitterMinDelay40       int     `json:"jitter_min_delay_40"`
	jitterMinDelay60       int     `json:"jitter_min_delay_60"`
	jitterMaxDelay20       int     `json:"jitter_max_delay_20"`
	jitterMaxDelay40       int     `json:"jitter_max_delay_40"`
	jitterMaxDelay60       int     `json:"jitter_max_delay_60"`
	jitterMaxSlots20       int     `json:"jitter_max_slots_20"`
	jitterMaxSlots40       int     `json:"jitter_max_slots_40"`
	jitterMaxSlots60       int     `json:"jitter_max_slots_60"`
	jitterLossesToReset    int     `json:"jitter_losses_to_reset"`
	jitterResyncThreshold  float32 `json:"jitter_resync_threshold"`
	audioCongestionWindow  int     `json:"audio_congestion_window"`
	audioMaxBitrate        int     `json:"audio_max_bitrate"`
	audioMaxBitrateEdge    int     `json:"audio_max_bitrate_edge"`
	audioMaxBitrateGprs    int     `json:"audio_max_bitrate_gprs"`
	audioMaxBitrateSaving  int     `json:"audio_max_bitrate_saving"`
	audioInitBitrate       int     `json:"audio_init_bitrate"`
	audioInitBitrateEdge   int     `json:"audio_init_bitrate_edge"`
	audioInitBitrateGrps   int     `json:"audio_init_bitrate_gprs"`
	audioInitBitrateSaving int     `json:"audio_init_bitrate_saving"`
	audioBitrateStepIncr   int     `json:"audio_bitrate_step_incr"`
	audioBitrateStepDecr   int     `json:"audio_bitrate_step_decr"`
	useSystemNs            bool    `json:"use_system_ns"`
	useSystemAec           bool    `json:"us audioInitBitrateGrps inte_system_aec"`
}

// TODO(@benqi): 写死配置
func NewCallConfigDataJSON() *callConfigDataJSON {
	return &callConfigDataJSON{
		audioFrameSize:         60,
		jitterMinDelay20:       6,
		jitterMinDelay40:       4,
		jitterMinDelay60:       2,
		jitterMaxDelay20:       25,
		jitterMaxDelay40:       15,
		jitterMaxDelay60:       10,
		jitterMaxSlots20:       50,
		jitterMaxSlots40:       30,
		jitterMaxSlots60:       20,
		jitterLossesToReset:    20,
		jitterResyncThreshold:  0.5,
		audioCongestionWindow:  1024,
		audioMaxBitrate:        20000,
		audioMaxBitrateEdge:    16000,
		audioMaxBitrateGprs:    8000,
		audioMaxBitrateSaving:  8000,
		audioInitBitrate:       16000,
		audioInitBitrateEdge:   8000,
		audioInitBitrateGrps:   8000,
		audioInitBitrateSaving: 8000,
		audioBitrateStepIncr:   1000,
		audioBitrateStepDecr:   1000,
		useSystemNs:            true,
		useSystemAec:           true,
	}
}

// phone.getCallConfig#55451fa9 = DataJSON;
func (s *PhoneServiceImpl) PhoneGetCallConfig(ctx context.Context, request *mtproto.TLPhoneGetCallConfig) (*mtproto.DataJSON, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("PhoneGetCallConfig - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	callConfig := NewCallConfigDataJSON()
	callConfigData, _ := json.Marshal(callConfig)
	reply := mtproto.NewTLDataJSON()
	j := string(callConfigData)
	reply.SetData(j)

	glog.Infof("PhoneGetCallConfig - reply %s", logger.JsonDebugData(reply))
	return reply.To_DataJSON(), nil
}
