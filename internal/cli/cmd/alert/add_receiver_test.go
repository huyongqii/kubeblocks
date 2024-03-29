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

package alert

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	clientfake "k8s.io/client-go/rest/fake"
	cmdtesting "k8s.io/kubectl/pkg/cmd/testing"

	"github.com/apecloud/kubeblocks/internal/cli/testing"
)

const (
	testNamespace = "test"
)

var mockConfigmap = func(name string, key string, value string) *corev1.ConfigMap {
	cm := &corev1.ConfigMap{}
	cm.Name = name
	cm.Namespace = testNamespace
	cm.Data = map[string]string{key: value}
	return cm
}

var mockBaseOptions = func(s genericclioptions.IOStreams) baseOptions {
	o := baseOptions{IOStreams: s}
	alertManagerConfig := `
    global:
      smtp_from: alert-test@apecloud.com
      smtp_smarthost: smtp.feishu.cn:587
      smtp_auth_username: alert-test@apecloud.com
      smtp_auth_password: 123456abc
      smtp_auth_identity: alert-test@apecloud.com
    receivers:
    - name: default-receiver
    - name: receiver-7pb52
      webhook_configs:
      - max_alerts: 10
        url: http://kubeblocks-webhook-adaptor-config.default:5001/api/v1/notify/receiver-7pb52
    route:
      group_interval: 30s
      group_wait: 5s
      receiver: default-receiver
      repeat_interval: 10m
      routes:
      - continue: true
        matchers:
        - app_kubernetes_io_instance=~a|b|c
        - severity=~info|warning
        receiver: receiver-7pb52`
	webhookAdaptorConfig := `
    receivers:
    - name: receiver-7pb52
      params:
        url: https://oapi.dingtalk.com/robot/send?access_token=123456
      type: dingtalk-webhook`
	alertCM := mockConfigmap(alertConfigmapName, alertConfigFileName, alertManagerConfig)
	webhookAdaptorCM := mockConfigmap(webhookAdaptorConfigmapName, webhookAdaptorFileName, webhookAdaptorConfig)
	o.alterConfigMap = alertCM
	o.webhookConfigMap = webhookAdaptorCM
	return o
}

var _ = Describe("add receiver", func() {
	var f *cmdtesting.TestFactory
	var s genericclioptions.IOStreams

	BeforeEach(func() {
		f = cmdtesting.NewTestFactory()
		f.Client = &clientfake.RESTClient{}
		s, _, _, _ = genericclioptions.NewTestIOStreams()
	})

	AfterEach(func() {
		f.Cleanup()
	})

	It("create new add receiver cmd", func() {
		cmd := newAddReceiverCmd(f, s)
		Expect(cmd).NotTo(BeNil())
		Expect(cmd.HasSubCommands()).Should(BeFalse())
	})

	It("complete", func() {
		o := baseOptions{IOStreams: s}
		Expect(o.complete(f)).Should(HaveOccurred())
	})

	It("validate", func() {
		By("nothing to be input, should fail")
		o := addReceiverOptions{baseOptions: baseOptions{IOStreams: s}}
		Expect(o.validate([]string{})).Should(HaveOccurred())

		By("set email, do not specify the name")
		o.emails = []string{"foo@bar.com"}
		o.alterConfigMap = mockConfigmap(alertConfigmapName, alertConfigFileName, "")
		Expect(o.validate([]string{})).Should(HaveOccurred())
		Expect(o.name).ShouldNot(BeEmpty())

		By("set email, specify the name")
		Expect(o.validate([]string{"test"})).Should(HaveOccurred())
		Expect(o.name).Should(Equal("test"))

		By("set email, set smtp config in configmap")
		baseOptions := mockBaseOptions(s)
		o.alterConfigMap = baseOptions.alterConfigMap
		Expect(o.validate([]string{})).Should(Succeed())
	})

	It("build receiver", func() {
		o := addReceiverOptions{baseOptions: baseOptions{IOStreams: s}}
		o.emails = []string{"foo@bar.com", "foo1@bar.com,foo2@bar.com"}
		o.webhooks = []string{"url=https://oapi.dingtalk.com/robot/send", "url=https://oapi.dingtalk.com/robot/send,url=https://oapi.dingtalk.com/robot/send?"}
		o.slacks = []string{"api_url=https://foo.com,channel=foo,username=test"}
		o.webhookConfigMap = mockConfigmap(webhookAdaptorConfigmapName, webhookAdaptorFileName, "")
		Expect(o.buildReceiver()).Should(Succeed())
		Expect(o.receiver).ShouldNot(BeNil())
		Expect(o.receiver.EmailConfigs).Should(HaveLen(3))
		Expect(o.receiver.WebhookConfigs).Should(HaveLen(2))
		Expect(o.receiver.SlackConfigs).Should(HaveLen(1))
	})

	It("build routes", func() {
		o := addReceiverOptions{baseOptions: baseOptions{IOStreams: s}}
		o.name = "receiver-test"
		o.clusters = []string{"cluster1", "cluster2"}
		o.severities = []string{"critical", "warning"}
		o.buildRoute()
		Expect(o.route).ShouldNot(BeNil())
		Expect(o.route.Receiver).Should(Equal(o.name))
		Expect(o.route.Matchers).Should(HaveLen(2))
		Expect(o.route.Matchers[0]).Should(ContainSubstring(routeMatcherClusterKey))
		Expect(o.route.Matchers[1]).Should(ContainSubstring(routeMatcherSeverityKey))
	})

	It("run", func() {
		o := addReceiverOptions{baseOptions: baseOptions{IOStreams: s}}
		alertCM := mockConfigmap(alertConfigmapName, alertConfigFileName, "")
		webhookAdaptorCM := mockConfigmap(webhookAdaptorConfigmapName, webhookAdaptorFileName, "")
		o.baseOptions.alterConfigMap = alertCM
		o.baseOptions.webhookConfigMap = webhookAdaptorCM
		o.client = testing.FakeClientSet(alertCM, webhookAdaptorCM)
		o.name = "receiver-test"
		Expect(o.addReceiver()).Should(Succeed())
		Expect(o.addWebhookReceivers()).Should(Succeed())
	})
})
