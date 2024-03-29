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

package cloudprovider

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("aws cloud provider", func() {
	const (
		tfPath              = "./testdata/aws/eks"
		expectedClusterName = "kb-playground-test"
	)

	It("get cluster name from state file", func() {
		By("get cluster name from state file")
		vals, err := getOutputValues(tfPath, clusterNameKey)
		Expect(err).Should(Succeed())
		Expect(vals).Should(HaveLen(1))
		Expect(vals).Should(ContainElement(expectedClusterName))

		By("get unknown key from state file")
		vals, err = getOutputValues(tfPath, "unknownKey")
		Expect(err).Should(Succeed())
		Expect(vals).ShouldNot(BeEmpty())
		Expect(vals).Should(ContainElement(""))
	})
})
