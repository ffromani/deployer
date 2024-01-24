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
 * Copyright 2024 Red Hat, Inc.
 */

package options

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
)

type Options struct {
	UserPlatform           platform.Platform
	UserPlatformVersion    platform.Version
	Replicas               int
	RTEConfigData          string
	PullIfNotPresent       bool
	UpdaterType            string
	UpdaterPFPEnable       bool
	UpdaterNotifEnable     bool
	UpdaterCRIHooksEnable  bool
	UpdaterSyncPeriod      time.Duration
	UpdaterVerbose         int
	SchedProfileName       string
	SchedResyncPeriod      time.Duration
	SchedVerbose           int
	SchedCtrlPlaneAffinity bool
	WaitInterval           time.Duration
	WaitTimeout            time.Duration
	ClusterPlatform        platform.Platform
	ClusterVersion         platform.Version
	WaitCompletion         bool
}

type API struct {
	Platform platform.Platform
}

type Scheduler struct {
	Platform          platform.Platform
	WaitCompletion    bool
	Replicas          int32
	ProfileName       string
	PullIfNotPresent  bool
	CacheResyncPeriod time.Duration
	CtrlPlaneAffinity bool
	Verbose           int
}

type DaemonSet struct {
	Verbose            int
	PullIfNotPresent   bool
	PFPEnable          bool
	NotificationEnable bool
	NodeSelector       *metav1.LabelSelector
	UpdateInterval     time.Duration
}

type UpdaterDaemon struct {
	DaemonSet                 DaemonSet
	MachineConfigPoolSelector *metav1.LabelSelector
	ConfigData                string
	Namespace                 string
	Name                      string
}

type Updater struct {
	Platform        platform.Platform
	PlatformVersion platform.Version
	WaitCompletion  bool
	RTEConfigData   string
	DaemonSet       DaemonSet
	EnableCRIHooks  bool
}

func ForDaemonSet(commonOpts *Options) DaemonSet {
	return DaemonSet{
		PullIfNotPresent:   commonOpts.PullIfNotPresent,
		PFPEnable:          commonOpts.UpdaterPFPEnable,
		NotificationEnable: commonOpts.UpdaterNotifEnable,
		UpdateInterval:     commonOpts.UpdaterSyncPeriod,
		Verbose:            commonOpts.UpdaterVerbose,
	}
}
