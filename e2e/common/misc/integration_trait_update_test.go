//go:build integration
// +build integration

// To enable compilation of this file in Goland, go to "Settings -> Go -> Vendoring & Build Tags -> Custom Tags" and add "integration"

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

package misc

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"

	. "github.com/apache/camel-k/v2/e2e/support"
	v1 "github.com/apache/camel-k/v2/pkg/apis/camel/v1"
)

func TestTraitUpdates(t *testing.T) {
	t.Parallel()

	WithNewTestNamespace(t, func(ns string) {

		t.Run("run and update trait", func(t *testing.T) {
			name := RandomizedSuffixName("yaml-route")
			Expect(KamelRunWithID(t, operatorID, ns, "files/yaml.yaml", "--name", name).Execute()).To(Succeed())
			Eventually(IntegrationPodPhase(t, ns, name), TestTimeoutLong).Should(Equal(corev1.PodRunning))
			Eventually(IntegrationConditionStatus(t, ns, name, v1.IntegrationConditionReady), TestTimeoutShort).Should(Equal(corev1.ConditionTrue))
			var numberOfPods = func(pods *int32) bool {
				return *pods >= 1 && *pods <= 2
			}
			// Adding a property will change the camel trait
			Expect(KamelRunWithID(t, operatorID, ns, "files/yaml.yaml", "--name", name, "-p", "hello=world").Execute()).To(Succeed())
			Consistently(IntegrationPodsNumbers(t, ns, name), TestTimeoutShort, 1*time.Second).Should(Satisfy(numberOfPods))
		})

		Expect(Kamel(t, "delete", "--all", "-n", ns).Execute()).To(Succeed())
	})
}
