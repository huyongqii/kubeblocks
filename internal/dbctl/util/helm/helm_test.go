/*
Copyright 2022 The KubeBlocks Authors

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

	"github.com/apecloud/kubeblocks/internal/dbctl/types"
)

var _ = Describe("helm util", func() {

	It("add Repo", func() {
		r := repo.Entry{
			Name: "mysql-operator",
			URL:  "https://mysql.github.io/mysql-operator/",
		}
		Expect(AddRepo(&r)).Should(Succeed())
		Expect(RemoveRepo(&r)).Should(Succeed())
	})

	It("Action Config", func() {
		cfg, err := NewActionConfig("test", "config")
		Expect(err == nil).To(BeTrue())
		Expect(cfg != nil).To(BeTrue())
	})

	It("Install", func() {
		o := &InstallOpts{
			Name:      types.DbaasHelmName,
			Chart:     types.DbaasHelmChart,
			Namespace: "default",
			Version:   types.DbaasDefaultVersion,
		}
		cfg := FakeActionConfig()
		Expect(cfg != nil).Should(BeTrue())
		Expect(o.Install(cfg)).Should(Succeed())
		Expect(o.UnInstall(cfg)).Should(Succeed())
	})
})