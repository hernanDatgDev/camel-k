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

package traits

import (
	"testing"

	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"

	. "github.com/apache/camel-k/v2/e2e/support"
	v1 "github.com/apache/camel-k/v2/pkg/apis/camel/v1"
)

func TestPodTrait(t *testing.T) {
	t.Parallel()

	WithNewTestNamespace(t, func(ns string) {

		tc := []struct {
			name         string
			templateName string
			assertions   func(t *testing.T, ns string, name string)
		}{
			{
				name:         "pod trait with env vars and volume mounts",
				templateName: "files/template.yaml",
				//nolint: thelper
				assertions: func(t *testing.T, ns string, name string) {
					// check that integrations is working and reading data created by sidecar container
					Eventually(IntegrationLogs(t, ns, name), TestTimeoutShort).Should(ContainSubstring("Content from the sidecar container"))
					// check that env var is injected
					Eventually(IntegrationLogs(t, ns, name), TestTimeoutShort).Should(ContainSubstring("hello from the template"))
					pod := IntegrationPod(t, ns, name)()

					// check if ENV variable is applied
					envValue := getEnvVar("TEST_VARIABLE", pod.Spec)
					Expect(envValue).To(Equal("hello from the template"))
				},
			},
			{
				name:         "pod trait with supplemental groups",
				templateName: "files/template-with-supplemental-groups.yaml",
				//nolint: thelper
				assertions: func(t *testing.T, ns string, name string) {
					Eventually(IntegrationPodHas(t, ns, name, func(pod *corev1.Pod) bool {
						if pod.Spec.SecurityContext == nil {
							return false
						}
						for _, sg := range pod.Spec.SecurityContext.SupplementalGroups {
							if sg == 666 {
								return true
							}
						}
						return false
					}), TestTimeoutShort).Should(BeTrue())
				},
			},
		}

		name := RandomizedSuffixName("pod-template-test")

		for i := range tc {
			test := tc[i]

			t.Run(test.name, func(t *testing.T) {
				Expect(KamelRunWithID(t, operatorID, ns, "files/PodTest.groovy",
					"--name", name,
					"--pod-template", test.templateName,
				).Execute()).To(Succeed())

				// check integration is deployed
				Eventually(IntegrationPodPhase(t, ns, name), TestTimeoutLong).Should(Equal(corev1.PodRunning))
				Eventually(IntegrationConditionStatus(t, ns, name, v1.IntegrationConditionReady), TestTimeoutShort).Should(Equal(corev1.ConditionTrue))

				test.assertions(t, ns, name)

				// Clean up
				Expect(Kamel(t, "delete", "--all", "-n", ns).Execute()).To(Succeed())
			})
		}

		Expect(Kamel(t, "delete", "--all", "-n", ns).Execute()).To(Succeed())
	})
}

func getEnvVar(name string, spec corev1.PodSpec) string {
	for _, i := range spec.Containers[0].Env {
		if i.Name == name {
			return i.Value
		}
	}
	return ""
}
