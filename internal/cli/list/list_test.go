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

package list

import (
	"bytes"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest/fake"
	cmdtesting "k8s.io/kubectl/pkg/cmd/testing"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	"github.com/apecloud/kubeblocks/internal/cli/types"
	"github.com/apecloud/kubeblocks/internal/cli/util"
)

var _ = Describe("List", func() {
	var cmd *cobra.Command
	var streams genericclioptions.IOStreams
	buf := new(bytes.Buffer)

	buildTestCmd := func(f cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
		o := NewListOptions(f, streams, schema.GroupVersionResource{Group: "", Resource: "pods", Version: types.K8sCoreAPIVersion})
		cmd := &cobra.Command{
			Use:   "ls-test",
			Short: "List test.",
			Run: func(cmd *cobra.Command, args []string) {
				_, err := o.Run()
				util.CheckErr(err)
			},
		}
		o.AddFlags(cmd)
		return cmd
	}

	mockClient := func(data runtime.Object) *cmdtesting.TestFactory {
		tf := cmdtesting.NewTestFactory().WithNamespace("test")
		defer tf.Cleanup()

		codec := scheme.Codecs.LegacyCodec(scheme.Scheme.PrioritizedVersionsAllGroups()...)
		tf.UnstructuredClient = &fake.RESTClient{
			NegotiatedSerializer: resource.UnstructuredPlusDefaultContentConfig().NegotiatedSerializer,
			Resp:                 &http.Response{StatusCode: http.StatusOK, Header: cmdtesting.DefaultHeader(), Body: cmdtesting.ObjBody(codec, data)},
		}
		return tf
	}

	BeforeEach(func() {
		pods, _, _ := cmdtesting.TestData()
		tf := mockClient(pods)
		streams, _, buf, _ = genericclioptions.NewTestIOStreams()
		cmd = buildTestCmd(tf, streams)
		cmd.SetOut(buf)
	})

	It("run", func() {
		pods, _, _ := cmdtesting.TestData()
		tf := mockClient(pods)
		streams, _, buf, _ := genericclioptions.NewTestIOStreams()
		cmd := buildTestCmd(tf, streams)
		cmd.Run(cmd, []string{})
		Expect(len(buf.String()) > 0).Should(BeTrue())
	})

	Context("List Objects", func() {
		It("Without any flags", func() {
			expected := `NAME   AGE
foo    <unknown>
bar    <unknown>
`
			cmd.Run(cmd, []string{})
			Expect(buf.String()).To(Equal(expected))
		})

		It("With -A flag", func() {
			expected := `NAMESPACE   NAME   AGE
test        foo    <unknown>
test        bar    <unknown>
`
			_ = cmd.Flags().Set("all-namespace", "true")
			cmd.Run(cmd, []string{})
			Expect(buf.String()).To(Equal(expected))
		})

		It("With -o wide flag", func() {
			expected := `NAME   AGE
foo    <unknown>
bar    <unknown>
`
			_ = cmd.Flags().Set("output", "wide")
			cmd.Run(cmd, []string{})
			Expect(buf.String()).To(Equal(expected))
		})

		It("With -o yaml flag", func() {
			expected := `apiVersion: v1
items:
- apiVersion: v1
  kind: Pod
  metadata:
    creationTimestamp: null
    name: foo
    namespace: test
    resourceVersion: "10"
  spec:
    containers: null
    dnsPolicy: ClusterFirst
    enableServiceLinks: true
    restartPolicy: Always
    securityContext: {}
    terminationGracePeriodSeconds: 30
  status: {}
- apiVersion: v1
  kind: Pod
  metadata:
    creationTimestamp: null
    name: bar
    namespace: test
    resourceVersion: "11"
  spec:
    containers: null
    dnsPolicy: ClusterFirst
    enableServiceLinks: true
    restartPolicy: Always
    securityContext: {}
    terminationGracePeriodSeconds: 30
  status: {}
kind: List
metadata:
  resourceVersion: ""
`
			_ = cmd.Flags().Set("output", "yaml")
			cmd.Run(cmd, []string{})
			Expect(buf.String()).To(Equal(expected))
		})

		It("With -o json flag", func() {
			expected := `{
    "apiVersion": "v1",
    "items": [
        {
            "apiVersion": "v1",
            "kind": "Pod",
            "metadata": {
                "creationTimestamp": null,
                "name": "foo",
                "namespace": "test",
                "resourceVersion": "10"
            },
            "spec": {
                "containers": null,
                "dnsPolicy": "ClusterFirst",
                "enableServiceLinks": true,
                "restartPolicy": "Always",
                "securityContext": {},
                "terminationGracePeriodSeconds": 30
            },
            "status": {}
        },
        {
            "apiVersion": "v1",
            "kind": "Pod",
            "metadata": {
                "creationTimestamp": null,
                "name": "bar",
                "namespace": "test",
                "resourceVersion": "11"
            },
            "spec": {
                "containers": null,
                "dnsPolicy": "ClusterFirst",
                "enableServiceLinks": true,
                "restartPolicy": "Always",
                "securityContext": {},
                "terminationGracePeriodSeconds": 30
            },
            "status": {}
        }
    ],
    "kind": "List",
    "metadata": {
        "resourceVersion": ""
    }
}
`
			_ = cmd.Flags().Set("output", "json")
			cmd.Run(cmd, []string{})
			Expect(buf.String()).To(Equal(expected))
		})

		It("No resources found", func() {
			tf := mockClient(&corev1.PodList{})
			streams, _, buf, errbuf := genericclioptions.NewTestIOStreams()
			cmd = buildTestCmd(tf, streams)
			cmd.SetOut(buf)
			cmd.Run(cmd, []string{})

			Expect(buf.String()).To(Equal(""))
			Expect(errbuf.String()).To(Equal("No pods found in test namespace.\n"))
		})
	})
})
