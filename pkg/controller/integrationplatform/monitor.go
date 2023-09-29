/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package integrationplatform

import (
	"context"
	"fmt"

	v1 "github.com/apache/camel-k/v2/pkg/apis/camel/v1"
	platformutil "github.com/apache/camel-k/v2/pkg/platform"
	"github.com/apache/camel-k/v2/pkg/util/defaults"
	"github.com/apache/camel-k/v2/pkg/util/openshift"
	corev1 "k8s.io/api/core/v1"
)

// NewMonitorAction returns an action that monitors the integration platform after it's fully initialized.
func NewMonitorAction() Action {
	return &monitorAction{}
}

type monitorAction struct {
	baseAction
}

func (action *monitorAction) Name() string {
	return "monitor"
}

func (action *monitorAction) CanHandle(platform *v1.IntegrationPlatform) bool {
	return platform.Status.Phase == v1.IntegrationPlatformPhaseReady || platform.Status.Phase == v1.IntegrationPlatformPhaseError
}

func (action *monitorAction) Handle(ctx context.Context, platform *v1.IntegrationPlatform) (*v1.IntegrationPlatform, error) {
	// Just track the version of the operator in the platform resource
	if platform.Status.Version != defaults.Version {
		platform.Status.Version = defaults.Version
		action.L.Info("IntegrationPlatform version updated", "version", platform.Status.Version)
	}

	// Refresh applied configuration
	if err := platformutil.ConfigureDefaults(ctx, action.client, platform, false); err != nil {
		return nil, err
	}

	// Registry condition
	isOpenshift, err := openshift.IsOpenShift(action.client)
	if err != nil {
		return platform, err
	}
	if isOpenshift {
		platform.Status.SetCondition(
			v1.IntegrationPlatformConditionTypeRegistryAvailable,
			corev1.ConditionFalse,
			"IntegrationPlatformRegistryAvailable",
			"registry not available because provided by Openshift")
	} else {
		if &platform.Status.Build.Registry == nil || platform.Status.Build.Registry.Address == "" {
			// error, we need a registry if we're not on Openshift
			platform.Status.Phase = v1.IntegrationPlatformPhaseError
			platform.Status.SetCondition(
				v1.IntegrationPlatformConditionTypeRegistryAvailable,
				corev1.ConditionFalse,
				"IntegrationPlatformRegistryAvailable",
				"registry not available")
		} else {
			platform.Status.Phase = v1.IntegrationPlatformPhaseReady
			platform.Status.SetCondition(
				v1.IntegrationPlatformConditionTypeRegistryAvailable,
				corev1.ConditionTrue,
				"IntegrationPlatformRegistryAvailable",
				fmt.Sprintf("registry available at %s", platform.Status.Build.Registry.Address))
		}
	}

	return platform, nil
}
