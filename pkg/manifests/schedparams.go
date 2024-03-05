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
	"time"

	"sigs.k8s.io/yaml"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
)

const (
	SchedulerConfigFileName = "scheduler-config.yaml" // TODO duplicate from yaml
	SchedulerPluginName     = "NodeResourceTopologyMatch"
)

const (
	ForeignPodsDetectNone                   = "None"
	ForeignPodsDetectAll                    = "All"
	ForeignPodsDetectOnlyExclusiveResources = "OnlyExclusiveResources"
)

const (
	CacheResyncAutodetect             = "Autodetect"
	CacheResyncAll                    = "All"
	CacheResyncOnlyExclusiveResources = "OnlyExclusiveResources"
)

const (
	CacheInformerShared    = "Shared"
	CacheInformerDedicated = "Dedicated"
)

const (
	ScoringStrategyMostAllocated      = "MostAllocated"
	ScoringStrategyBalancedAllocation = "BalancedAllocation"
	ScoringStrategyLeastAllocated     = "LeastAllocated"
)

func ValidateForeignPodsDetectMode(value string) error {
	switch value {
	case ForeignPodsDetectNone:
		return nil
	case ForeignPodsDetectAll:
		return nil
	case ForeignPodsDetectOnlyExclusiveResources:
		return nil
	default:
		return fmt.Errorf("unsupported foreignPodsDetectMode: %v", value)
	}
}

func ValidateCacheResyncMethod(value string) error {
	switch value {
	case CacheResyncAutodetect:
		return nil
	case CacheResyncAll:
		return nil
	case CacheResyncOnlyExclusiveResources:
		return nil
	default:
		return fmt.Errorf("unsupported cacheResyncMethod: %v", value)
	}
}

func ValidateCacheInformerMode(value string) error {
	switch value {
	case CacheInformerShared:
		return nil
	case CacheInformerDedicated:
		return nil
	default:
		return fmt.Errorf("unsupported cacheInformerMode: %v", value)
	}
}

type ConfigCacheParams struct {
	ResyncPeriodSeconds   *int64  `json:"-"`
	ResyncMethod          *string `json:"resyncMethod,omitempty"`
	ForeignPodsDetectMode *string `json:"foreignPodsDetect,omitempty"`
	InformerMode          *string `json:"informerMode,omitempty"`
}

func SetDefaultsConfigCacheParams(params *ConfigCacheParams) {
	params.ResyncPeriodSeconds = newInt64(int64(5 * time.Second))
	params.ResyncMethod = newString(CacheResyncAutodetect)
	params.ForeignPodsDetectMode = newString(ForeignPodsDetectOnlyExclusiveResources)
	// InformerMode support intentionally left unspecified
}

func NewConfigCacheParams() *ConfigCacheParams {
	params := ConfigCacheParams{}
	SetDefaultsConfigCacheParams(&params)
	return &params
}

type ResourceSpecParams struct {
	// Name of the resource.
	Name string `json:"name"`
	// Weight of the resource.
	Weight int64 `json:"weight,omitempty"`
}

type ScoringStrategyParams struct {
	Type      string               `json:"type,omitempty"`
	Resources []ResourceSpecParams `json:"resources,omitempty"`
}

func ValidateScoringStrategyType(value string) error {
	switch value {
	case ScoringStrategyMostAllocated:
		return nil
	case ScoringStrategyBalancedAllocation:
		return nil
	case ScoringStrategyLeastAllocated:
		return nil
	default:
		return fmt.Errorf("unsupported scoringStrategyType: %v", value)
	}
}

type ConfigParams struct {
	ProfileName     string // can't be empty, so no need for pointer
	Cache           *ConfigCacheParams
	ScoringStrategy *ScoringStrategyParams
}

func DecodeSchedulerProfilesFromData(data []byte) ([]ConfigParams, error) {
	params := []ConfigParams{}

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

		pluginConfigs, ok, err := unstructured.NestedSlice(profile, "pluginConfig")
		if !ok || err != nil {
			klog.ErrorS(err, "failed to process unstructured data", "pluginConfig", ok)
			return params, nil
		}
		for _, plConf := range pluginConfigs {
			pluginConf, ok := plConf.(map[string]interface{})
			if !ok {
				klog.V(1).InfoS("unexpected profile config data")
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

			profileParams, err := extractParams(profileName, args)
			if err != nil {
				klog.ErrorS(err, "failed to extract params", "name", name, "profile", profileName)
				continue
			}

			params = append(params, profileParams)
		}
	}

	return params, nil
}

func FindSchedulerProfileByName(profileParams []ConfigParams, schedulerName string) *ConfigParams {
	for idx := range profileParams {
		params := &profileParams[idx]
		if params.ProfileName == schedulerName {
			return params
		}
	}
	return nil
}

func extractParams(profileName string, args map[string]interface{}) (ConfigParams, error) {
	params := ConfigParams{
		ProfileName: profileName,
		Cache:       &ConfigCacheParams{},
	}
	// json quirk: we know it's int64, yet it's detected as float64
	resyncPeriod, ok, err := unstructured.NestedFloat64(args, "cacheResyncPeriodSeconds")
	if err != nil {
		return params, fmt.Errorf("cannot process field cacheResyncPeriodSeconds: %w", err)
	}
	if ok {
		val := int64(resyncPeriod)
		params.Cache.ResyncPeriodSeconds = &val
	}

	cacheArgs, ok, err := unstructured.NestedMap(args, "cache")
	if err != nil {
		return params, fmt.Errorf("cannot process field cache: %w", err)
	}
	if ok {
		resyncMethod, cacheOk, err := unstructured.NestedString(cacheArgs, "resyncMethod")
		if err != nil {
			return params, fmt.Errorf("cannot process field cache.resyncMethod: %w", err)
		}
		if cacheOk {
			if err := ValidateCacheResyncMethod(resyncMethod); err != nil {
				return params, err
			}
			params.Cache.ResyncMethod = &resyncMethod
		}

		foreignPodsDetect, cacheOk, err := unstructured.NestedString(cacheArgs, "foreignPodsDetect")
		if err != nil {
			return params, fmt.Errorf("cannot process field cache.foreignPodsDetect: %w", err)
		}
		if cacheOk {
			if err := ValidateForeignPodsDetectMode(foreignPodsDetect); err != nil {
				return params, err
			}
			params.Cache.ForeignPodsDetectMode = &foreignPodsDetect
		}

		informerMode, cacheOk, err := unstructured.NestedString(cacheArgs, "informerMode")
		if err != nil {
			return params, fmt.Errorf("cannot process field cache.informerMode: %w", err)
		}
		if cacheOk {
			if err := ValidateCacheInformerMode(informerMode); err != nil {
				return params, err
			}
			params.Cache.InformerMode = &informerMode
		}
	}

	scoringStratArgs, ok, err := unstructured.NestedMap(args, "scoringStrategy")
	if err != nil {
		return params, fmt.Errorf("cannot process field scoringStrategy: %w", err)
	}
	if ok {
		params.ScoringStrategy = &ScoringStrategyParams{}

		scoringType, cacheOk, err := unstructured.NestedString(scoringStratArgs, "type")
		if err != nil {
			return params, fmt.Errorf("cannot process field scoringStrategy.type: %w", err)
		}
		if cacheOk {
			if err := ValidateScoringStrategyType(scoringType); err != nil {
				return params, err
			}
			params.ScoringStrategy.Type = scoringType
		}

		scoringRess, cacheOk, err := unstructured.NestedSlice(scoringStratArgs, "resources")
		if err != nil {
			return params, fmt.Errorf("cannot process field scoringStrategy.resources: %w", err)
		}
		if cacheOk {
			var resources []ResourceSpecParams
			for idx, scRes := range scoringRess {
				res, ok := scRes.(map[string]interface{})
				if !ok {
					return params, fmt.Errorf("unexpected scoringStrategy.resources[%d] data", idx)
				}

				name, ok, err := unstructured.NestedString(res, "name")
				if !ok || err != nil {
					return params, fmt.Errorf("unexpected scoringStrategy.resources[%d].name data (err=%v)", idx, err)
				}

				weight, ok, err := unstructured.NestedFloat64(res, "weight")
				if !ok || err != nil {
					return params, fmt.Errorf("unexpected scoringStrategy.resources[%d].weight data (err=%v)", idx, err)
				}

				resources = append(resources, ResourceSpecParams{
					Name:   name,
					Weight: int64(weight),
				})
			}
			params.ScoringStrategy.Resources = resources
		}
	}

	return params, nil
}

func newInt64(v int64) *int64 {
	return &v
}

func newString(v string) *string {
	return &v
}
