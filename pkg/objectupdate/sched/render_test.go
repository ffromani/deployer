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

package sched

import (
	"testing"
)

func TestRenderConfig(t *testing.T) {
	type testCase struct {
		name     string
		params   *ConfigParams
		initial  string
		expected string
	}
	testCases := []testCase{
		{
			name:     "nil",
			params:   nil,
			initial:  configTemplateEmpty,
			expected: configTemplateEmpty,
		},
		{
			name:     "nil cache",
			params:   &ConfigParams{},
			initial:  configTemplateEmpty,
			expected: configTemplateEmpty,
		},
		{
			name: "resync=zero",
			params: &ConfigParams{
				Cache: &ConfigCacheParams{
					ResyncPeriodSeconds: newInt64(0),
				},
			},
			initial:  configTemplateEmpty,
			expected: configTemplateEmpty,
		},
		{
			name: "resync cleared if zero",
			params: &ConfigParams{
				Cache: &ConfigCacheParams{
					ResyncPeriodSeconds: newInt64(0),
				},
			},
			initial:  configTemplateAllValues,
			expected: configTemplateEmpty,
		},
		{
			name: "resync updated from non empty",
			params: &ConfigParams{
				Cache: &ConfigCacheParams{
					ResyncPeriodSeconds: newInt64(42),
				},
			},
			initial: configTemplateAllValues,
			expected: `apiVersion: kubescheduler.config.k8s.io/v1beta2
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: false
profiles:
- pluginConfig:
  - args:
      cacheResyncPeriodSeconds: 42
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
`,
		},
		{
			name: "resync updated from empty",
			params: &ConfigParams{
				Cache: &ConfigCacheParams{
					ResyncPeriodSeconds: newInt64(42),
				},
			},
			initial: configTemplateEmpty,
			expected: `apiVersion: kubescheduler.config.k8s.io/v1beta2
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: false
profiles:
- pluginConfig:
  - args:
      cacheResyncPeriodSeconds: 42
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
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := RenderConfig([]byte(tc.initial), "test-sched-name", tc.params)
			if err != nil {
				t.Errorf("RenderConfig() failed: %v", err)
			}

			rendered := string(data)
			if rendered != tc.expected {
				t.Errorf("rendering failed.\nrendered=[%s]\nexpected=[%s]\n", rendered, tc.expected)
			}
		})
	}
}

/*
 */

var configTemplateEmpty string = `apiVersion: kubescheduler.config.k8s.io/v1beta2
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
`

var configTemplateAllValues string = `apiVersion: kubescheduler.config.k8s.io/v1beta2
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
`

func newInt64(value int64) *int64 {
	return &value
}
