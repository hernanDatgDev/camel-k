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

package trait

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"

	traitv1 "github.com/apache/camel-k/v2/pkg/apis/camel/v1/trait"
	"github.com/apache/camel-k/v2/pkg/util/kubernetes"
)

const (
	tolerationTraitID    = "toleration"
	tolerationTraitOrder = 1200
)

type tolerationTrait struct {
	BaseTrait
	traitv1.TolerationTrait `property:",squash"`
}

func newTolerationTrait() Trait {
	return &tolerationTrait{
		BaseTrait: NewBaseTrait(tolerationTraitID, tolerationTraitOrder),
	}
}

func (t *tolerationTrait) Configure(e *Environment) (bool, *TraitCondition, error) {
	if e.Integration == nil || !ptr.Deref(t.Enabled, false) {
		return false, nil, nil
	}

	if len(t.Taints) == 0 {
		return false, nil, fmt.Errorf("no taint was provided")
	}

	return e.IntegrationInRunningPhases(), nil, nil
}

func (t *tolerationTrait) Apply(e *Environment) error {
	tolerations, err := kubernetes.NewTolerations(t.Taints)
	if err != nil {
		return err
	}
	podSpec := e.GetIntegrationPodSpec()

	if podSpec == nil {
		return fmt.Errorf("could not find any integration deployment for %v", e.Integration.Name)
	}
	if podSpec.Tolerations == nil {
		podSpec.Tolerations = make([]corev1.Toleration, 0)
	}
	podSpec.Tolerations = append(podSpec.Tolerations, tolerations...)
	return nil
}
