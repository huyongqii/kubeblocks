/*
Copyright ApeCloud Inc.

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

package helm

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"helm.sh/helm/v3/pkg/repo"

	"github.com/apecloud/kubeblocks/internal/cli/testing"
	"github.com/apecloud/kubeblocks/internal/cli/types"
	"github.com/apecloud/kubeblocks/version"
)

var _ = Describe("helm util", func() {
	It("repo", func() {
		r := repo.Entry{
			Name: "test-repo",
			URL:  "https://test-kubebllcks.com/test-repo",
		}
		Expect(AddRepo(&r)).Should(HaveOccurred())
		Expect(RemoveRepo(&r)).Should(Succeed())
	})

	It("Action Config", func() {
		cfg, err := NewActionConfig("test", "config")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(cfg).ShouldNot(BeNil())
	})

	It("Install", func() {
		o := &InstallOpts{
			Name:      testing.KubeBlocksChartName,
			Chart:     testing.KubeBlocksChartURL,
			Namespace: "default",
			Version:   version.DefaultKubeBlocksVersion,
		}
		cfg := FakeActionConfig()
		Expect(cfg).ShouldNot(BeNil())
		_, err := o.Install(cfg)
		Expect(err).Should(HaveOccurred())
		Expect(o.UnInstall(cfg)).Should(HaveOccurred())
	})

	It("Upgrade", func() {
		o := &InstallOpts{
			Name:      types.KubeBlocksChartName,
			Chart:     "kubeblocks-test-chart",
			Namespace: "default",
			Version:   version.DefaultKubeBlocksVersion,
		}
		cfg := FakeActionConfig()
		Expect(cfg).ShouldNot(BeNil())
		_, err := o.Upgrade(cfg)
		Expect(err).Should(HaveOccurred())
		Expect(o.UnInstall(cfg)).Should(HaveOccurred())
	})
})