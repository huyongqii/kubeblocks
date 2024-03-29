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

package playground

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cp "github.com/apecloud/kubeblocks/internal/cli/cloudprovider"
	clitesting "github.com/apecloud/kubeblocks/internal/cli/testing"
	"github.com/apecloud/kubeblocks/internal/cli/types"
)

func TestPlayground(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PlayGround Suite")
}

var _ = BeforeSuite(func() {
	// set fake image info
	cp.K3sImage = "fake-k3s-image"
	cp.K3dToolsImage = "fake-k3s-tools-image"
	cp.K3dProxyImage = "fake-k3d-proxy-image"

	// set default cluster name to test
	types.K3dClusterName = "kb-playground-test"
	kbClusterName = "kb-playground-test-cluster"

	// use a fake URL to test
	types.KubeBlocksRepoName = clitesting.KubeBlocksRepoName
	types.KubeBlocksChartName = clitesting.KubeBlocksChartName
	types.KubeBlocksChartURL = clitesting.KubeBlocksChartURL
})
