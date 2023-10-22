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
	"fmt"

	"sigs.k8s.io/yaml"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
)

const (
	SchedulerConfigFileName = "scheduler-config.yaml" // TODO duplicate from yaml
	SchedulerPluginName     = "NodeResourceTopologyMatch"
)

type ConfigCacheParams struct {
	ResyncPeriodSeconds *int64
}

type ConfigParams struct {
	Cache *ConfigCacheParams
}

func DecodeSchedulerConfigFromData(data []byte, schedulerName string) (ConfigParams, error) {
	params := ConfigParams{}
	if schedulerName == "" {
		klog.InfoS("missing scheduler name")
		return params, nil
	}

	var r unstructured.Unstructured
	if err := yaml.Unmarshal(data, &r.Object); err != nil {
		klog.ErrorS(err, "cannot unmarshal scheduler config")
		return params, nil
	}

	profiles, ok, err := unstructured.NestedSlice(r.Object, "profiles")
	if !ok || err != nil {
		klog.ErrorS(err, "failed to process unstructured data", "profiles", ok)
		return params, nil
	}
	for _, prof := range profiles {
		profile, ok := prof.(map[string]interface{})
		if !ok {
			klog.V(1).InfoS("unexpected profile data")
			return params, nil
		}

		profileName, ok, err := unstructured.NestedString(profile, "schedulerName")
		if !ok || err != nil {
			klog.ErrorS(err, "failed to get profile name", "profileName", ok)
			return params, nil
		}

		if profileName != schedulerName {
			continue
		}

		pluginConfigs, ok, err := unstructured.NestedSlice(profile, "pluginConfig")
		if !ok || err != nil {
			klog.ErrorS(err, "failed to process unstructured data", "pluginConfig", ok)
			return params, nil
		}
		for _, plConf := range pluginConfigs {
			pluginConf, ok := plConf.(map[string]interface{})
			if !ok {
				klog.V(1).InfoS("unexpected profile coonfig data")
				return params, nil
			}

			name, ok, err := unstructured.NestedString(pluginConf, "name")
			if !ok || err != nil {
				klog.ErrorS(err, "failed to process unstructured data", "name", ok)
				return params, nil
			}
			if name != SchedulerPluginName {
				continue
			}
			args, ok, err := unstructured.NestedMap(pluginConf, "args")
			if !ok || err != nil {
				klog.ErrorS(err, "failed to process unstructured data", "args", ok)
				return params, nil
			}

			return extractParams(args)
		}
	}

	return params, fmt.Errorf("failed to find parameters for schedulerName %q", schedulerName)
}

func extractParams(args map[string]interface{}) (ConfigParams, error) {
	params := ConfigParams{
		Cache: &ConfigCacheParams{},
	}
	// json quirk: we know it's int64, yet it's detected as float64
	resyncPeriod, ok, err := unstructured.NestedFloat64(args, "cacheResyncPeriodSeconds")
	if !ok {
		// nothing to do
		return params, nil
	}
	if err != nil {
		return params, fmt.Errorf("cannot process field cacheResyncPeriodSeconds: %w", err)
	}

	val := int64(resyncPeriod)
	params.Cache.ResyncPeriodSeconds = &val
	return params, nil
}
