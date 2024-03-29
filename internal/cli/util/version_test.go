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

package util

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/client-go/kubernetes"

	"github.com/apecloud/kubeblocks/internal/cli/testing"
)

const kbVersion = "0.3.0"

var _ = Describe("version util", func() {
	It("get version info when client is nil", func() {
		info, err := GetVersionInfo(nil)
		Expect(err).Should(Succeed())
		Expect(info).ShouldNot(BeEmpty())
		Expect(info[KubeBlocksApp]).Should(BeEmpty())
		Expect(info[KubernetesApp]).Should(BeEmpty())
		Expect(info[KBCLIApp]).ShouldNot(BeEmpty())
	})

	It("get version info when client variable is a nil pointer", func() {
		var client *kubernetes.Clientset
		info, err := GetVersionInfo(client)
		Expect(err).Should(Succeed())
		Expect(info).ShouldNot(BeEmpty())
		Expect(info[KubeBlocksApp]).Should(BeEmpty())
		Expect(info[KubernetesApp]).Should(BeEmpty())
		Expect(info[KBCLIApp]).ShouldNot(BeEmpty())
	})

	It("get version info when KubeBlocks is deployed", func() {
		client := testing.FakeClientSet(testing.FakeKBDeploy(kbVersion))
		info, err := GetVersionInfo(client)
		Expect(err).Should(Succeed())
		Expect(info).ShouldNot(BeEmpty())
		Expect(info[KubeBlocksApp]).Should(Equal(kbVersion))
		Expect(info[KubernetesApp]).ShouldNot(BeEmpty())
		Expect(info[KBCLIApp]).ShouldNot(BeEmpty())
	})

	It("get version info when KubeBlocks is not deployed", func() {
		client := testing.FakeClientSet()
		info, err := GetVersionInfo(client)
		Expect(err).Should(Succeed())
		Expect(info).ShouldNot(BeEmpty())
		Expect(info[KubeBlocksApp]).Should(BeEmpty())
		Expect(info[KubernetesApp]).ShouldNot(BeEmpty())
		Expect(info[KBCLIApp]).ShouldNot(BeEmpty())
	})

	It("getKubeBlocksVersion", func() {
		client := testing.FakeClientSet(testing.FakeKBDeploy(""))
		v, err := getKubeBlocksVersion(client)
		Expect(v).Should(BeEmpty())
		Expect(err).Should(HaveOccurred())

		client = testing.FakeClientSet(testing.FakeKBDeploy(kbVersion))
		v, err = getKubeBlocksVersion(client)
		Expect(v).Should(Equal(kbVersion))
		Expect(err).Should(Succeed())
	})

	It("getK8sVersion", func() {
		client := testing.FakeClientSet()
		v, err := getK8sVersion(client.Discovery())
		Expect(v).ShouldNot(BeEmpty())
		Expect(err).Should(Succeed())
	})
})
