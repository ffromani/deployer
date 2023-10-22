/*
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2023 Red Hat, Inc.
 */

package manifests

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestDecodeSchedulerConfigFromData(t *testing.T) {
	type testCase struct {
		name           string
		data           []byte
		schedulerName  string
		expectedParams ConfigParams
		expectedError  bool
	}
	testCases := []testCase{
		{
			name:           "nil",
			data:           nil,
			schedulerName:  "",
			expectedParams: ConfigParams{},
			expectedError:  false,
		},
		{
			name: "bad scheduler name",
			data: []byte(`apiVersion: kubescheduler.config.k8s.io/v1beta2
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: false
profiles:
- pluginConfig:
  - args: {}
    name: NodeResourceTopologyMatch
  plugins:
    filter:
      enabled:
      - name: NodeResourceTopologyMatch
    reserve:
      enabled:
      - name: NodeResourceTopologyMatch
    score:
      enabled:
      - name: NodeResourceTopologyMatch
  schedulerName: topology-aware-scheduler
`),
			schedulerName:  "topo-aware-scheduler",
			expectedParams: ConfigParams{},
			expectedError:  true,
		},
		{
			name: "bad scheduler params name",
			data: []byte(`apiVersion: kubescheduler.config.k8s.io/v1beta2
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: false
profiles:
- pluginConfig:
  - args: {}
    name: noderestopo
  plugins:
    filter:
      enabled:
      - name: NodeResourceTopologyMatch
    reserve:
      enabled:
      - name: NodeResourceTopologyMatch
    score:
      enabled:
      - name: NodeResourceTopologyMatch
  schedulerName: topology-aware-scheduler
`),
			schedulerName:  "topology-aware-scheduler",
			expectedParams: ConfigParams{},
			expectedError:  true,
		},
		{
			name: "empty params",
			data: []byte(`apiVersion: kubescheduler.config.k8s.io/v1beta2
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: false
profiles:
- pluginConfig:
  - args: {}
    name: NodeResourceTopologyMatch
  plugins:
    filter:
      enabled:
      - name: NodeResourceTopologyMatch
    reserve:
      enabled:
      - name: NodeResourceTopologyMatch
    score:
      enabled:
      - name: NodeResourceTopologyMatch
  schedulerName: topology-aware-scheduler
`),
			schedulerName: "topology-aware-scheduler",
			expectedParams: ConfigParams{
				Cache: &ConfigCacheParams{},
			},
			expectedError: false,
		},
		{
			name: "nonzero resync period",
			data: []byte(`apiVersion: kubescheduler.config.k8s.io/v1beta2
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: false
profiles:
- pluginConfig:
  - args:
      cacheResyncPeriodSeconds: 5
    name: NodeResourceTopologyMatch
  plugins:
    filter:
      enabled:
      - name: NodeResourceTopologyMatch
    reserve:
      enabled:
      - name: NodeResourceTopologyMatch
    score:
      enabled:
      - name: NodeResourceTopologyMatch
  schedulerName: topology-aware-scheduler
`),
			schedulerName: "topology-aware-scheduler",
			expectedParams: ConfigParams{
				Cache: &ConfigCacheParams{
					ResyncPeriodSeconds: newInt64(5),
				},
			},
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params, err := DecodeSchedulerConfigFromData(tc.data, tc.schedulerName)
			if (err != nil) != tc.expectedError {
				t.Fatalf("unexpected error [%v] expected=%v", err, tc.expectedError)
			}
			if !reflect.DeepEqual(params, tc.expectedParams) {
				t.Fatalf("params got %q expected %q", toJSON(params), toJSON(tc.expectedParams))
			}
		})
	}
}

func toJSON(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("<err=%v>", err)
	}
	return string(data)
}

func newInt64(value int64) *int64 {
	return &value
}
