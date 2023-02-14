/*
Copyright ApeCloud, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package component

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"

	dbaasv1alpha1 "github.com/apecloud/kubeblocks/apis/dbaas/v1alpha1"
)

var _ = Describe("probe_utils", func() {

	BeforeEach(func() {
		// Add any steup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	Context("buildProbeContainers", func() {
		var container *corev1.Container
		var component *Component
		var probeServiceHTTPPort, probeServiceGrpcPort int
		var clusterDefProbe *dbaasv1alpha1.ClusterDefinitionProbe

		BeforeEach(func() {
			var err error
			container, err = buildProbeContainer()
			Expect(err).NotTo(HaveOccurred())
			probeServiceHTTPPort, probeServiceGrpcPort = 3501, 50001

			clusterDefProbe = &dbaasv1alpha1.ClusterDefinitionProbe{}
			clusterDefProbe.PeriodSeconds = 1
			clusterDefProbe.TimeoutSeconds = 1
			clusterDefProbe.FailureThreshold = 1
			component = &Component{}
			component.CharacterType = "mysql"
		})

		It("Build role changed probe container", func() {
			buildRoleChangedProbeContainer("wesql", container, clusterDefProbe, probeServiceHTTPPort)
			Expect(len(container.ReadinessProbe.Exec.Command)).ShouldNot(BeZero())
		})

		It("Build role service container", func() {
			buildProbeServiceContainer(component, container, probeServiceHTTPPort, probeServiceGrpcPort)
			Expect(len(container.Command)).ShouldNot(BeZero())
		})
	})
})