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
 * Copyright 2022 Red Hat, Inc.
 */

package updaters

import (
	"fmt"

	"github.com/go-logr/logr"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests"
	nfdmanifests "github.com/k8stopologyawareschedwg/deployer/pkg/manifests/nfd"
	rtemanifests "github.com/k8stopologyawareschedwg/deployer/pkg/manifests/rte"
)

func GetObjects(opts Options, updaterType, namespace string, mcOpts manifests.MachineConfigOptions) ([]client.Object, error) {

	if updaterType == RTE {
		mf, err := rtemanifests.GetManifests(opts.Platform, opts.PlatformVersion, namespace, mcOpts)
		if err != nil {
			return nil, err
		}
		ret, err := mf.Render(rteOptionsFrom(opts, namespace))
		if err != nil {
			return nil, err
		}
		return ret.ToObjects(), nil
	}
	if updaterType == NFD {
		mf, err := nfdmanifests.GetManifests(opts.Platform, namespace)
		if err != nil {
			return nil, err
		}
		ret, err := mf.Render(nfdOptionsFrom(opts, namespace))
		if err != nil {
			return nil, err
		}
		return ret.ToObjects(), nil
	}
	return nil, fmt.Errorf("unsupported updater: %q", updaterType)
}

func getCreatableObjects(opts Options, cli client.Client, log logr.Logger, updaterType, namespace string, mcOpts manifests.MachineConfigOptions) ([]deployer.WaitableObject, error) {
	if updaterType == RTE {
		mf, err := rtemanifests.GetManifests(opts.Platform, opts.PlatformVersion, namespace, mcOpts)
		if err != nil {
			return nil, err
		}
		ret, err := mf.Render(rteOptionsFrom(opts, namespace))
		if err != nil {
			return nil, err
		}
		return ret.ToCreatableObjects(cli, log), nil
	}
	if updaterType == NFD {
		mf, err := nfdmanifests.GetManifests(opts.Platform, namespace)
		if err != nil {
			return nil, err
		}
		ret, err := mf.Render(nfdOptionsFrom(opts, namespace))
		if err != nil {
			return nil, err
		}
		return ret.ToCreatableObjects(cli, log), nil
	}
	return nil, fmt.Errorf("unsupported updater: %q", updaterType)
}

func getDeletableObjects(opts Options, cli client.Client, log logr.Logger, updaterType, namespace string, mcOpts manifests.MachineConfigOptions) ([]deployer.WaitableObject, error) {
	if updaterType == RTE {
		mf, err := rtemanifests.GetManifests(opts.Platform, opts.PlatformVersion, namespace, mcOpts)
		if err != nil {
			return nil, err
		}
		ret, err := mf.Render(rteOptionsFrom(opts, namespace))
		if err != nil {
			return nil, err
		}
		return ret.ToDeletableObjects(cli, log), nil
	}
	if updaterType == NFD {
		mf, err := nfdmanifests.GetManifests(opts.Platform, namespace)
		if err != nil {
			return nil, err
		}
		ret, err := mf.Render(nfdOptionsFrom(opts, namespace))
		if err != nil {
			return nil, err
		}
		return ret.ToDeletableObjects(cli, log), nil
	}
	return nil, fmt.Errorf("unsupported updater: %q", updaterType)
}

func rteOptionsFrom(opts Options, namespace string) rtemanifests.RenderOptions {
	return rtemanifests.RenderOptions{
		ConfigData:       opts.RTEConfigData,
		PullIfNotPresent: opts.PullIfNotPresent,
		Namespace:        namespace,
	}
}

func nfdOptionsFrom(opts Options, namespace string) nfdmanifests.RenderOptions {
	return nfdmanifests.RenderOptions{
		PullIfNotPresent: opts.PullIfNotPresent,
		Namespace:        namespace,
	}
}
