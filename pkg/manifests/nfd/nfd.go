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

package nfd

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests"
	nfdupdate "github.com/k8stopologyawareschedwg/deployer/pkg/objectupdate/nfd"
	rbacupdate "github.com/k8stopologyawareschedwg/deployer/pkg/objectupdate/rbac"
	"github.com/k8stopologyawareschedwg/deployer/pkg/options"
)

type Manifests struct {
	Namespace          *corev1.Namespace
	SATopologyUpdater  *corev1.ServiceAccount
	CRTopologyUpdater  *rbacv1.ClusterRole
	CRBTopologyUpdater *rbacv1.ClusterRoleBinding
	DSTopologyUpdater  *appsv1.DaemonSet

	plat platform.Platform
}

func (mf Manifests) Clone() Manifests {
	ret := Manifests{
		plat:               mf.plat,
		Namespace:          mf.Namespace.DeepCopy(),
		CRTopologyUpdater:  mf.CRTopologyUpdater.DeepCopy(),
		CRBTopologyUpdater: mf.CRBTopologyUpdater.DeepCopy(),
		DSTopologyUpdater:  mf.DSTopologyUpdater.DeepCopy(),
		SATopologyUpdater:  mf.SATopologyUpdater.DeepCopy(),
	}

	return ret
}

func (mf Manifests) Render(opts options.UpdaterDaemon) (Manifests, error) {
	ret := mf.Clone()

	if opts.Namespace != "" {
		ret.Namespace.Name = opts.Namespace
	}

	rbacupdate.ClusterRoleBinding(ret.CRBTopologyUpdater, mf.SATopologyUpdater.Name, ret.Namespace.Name)

	ret.DSTopologyUpdater.Spec.Template.Spec.ServiceAccountName = mf.SATopologyUpdater.Name

	nfdupdate.UpdaterDaemonSet(ret.DSTopologyUpdater, opts.DaemonSet)

	return ret, nil
}

func (mf Manifests) ToObjects() []client.Object {
	return []client.Object{
		mf.Namespace,
		// topology-updater objects
		mf.SATopologyUpdater,
		mf.CRTopologyUpdater,
		mf.CRBTopologyUpdater,
		mf.DSTopologyUpdater,
	}
}

func New(plat platform.Platform) Manifests {
	mf := Manifests{
		plat: plat,
	}

	return mf
}

func NewWithOptions(opts options.Render) (Manifests, error) {
	var err error
	mf := New(opts.Platform)

	mf.Namespace, err = manifests.Namespace(manifests.ComponentNodeFeatureDiscovery)
	if err != nil {
		return mf, err
	}

	mf.SATopologyUpdater, err = manifests.ServiceAccount(manifests.ComponentNodeFeatureDiscovery, manifests.SubComponentNodeFeatureDiscoveryTopologyUpdater, opts.Namespace)
	if err != nil {
		return mf, err
	}
	mf.CRTopologyUpdater, err = manifests.ClusterRole(manifests.ComponentNodeFeatureDiscovery, manifests.SubComponentNodeFeatureDiscoveryTopologyUpdater)
	if err != nil {
		return mf, err
	}
	mf.CRBTopologyUpdater, err = manifests.ClusterRoleBinding(manifests.ComponentNodeFeatureDiscovery, manifests.SubComponentNodeFeatureDiscoveryTopologyUpdater)
	if err != nil {
		return mf, err
	}
	mf.DSTopologyUpdater, err = manifests.DaemonSet(manifests.ComponentNodeFeatureDiscovery, manifests.SubComponentNodeFeatureDiscoveryTopologyUpdater, opts.Namespace)
	if err != nil {
		return mf, err
	}

	return mf, nil
}

// GetManifests is deprecated, use NewWithOptions in new code
func GetManifests(plat platform.Platform, namespace string) (Manifests, error) {
	return NewWithOptions(options.Render{
		Platform:  plat,
		Namespace: namespace,
	})
}
