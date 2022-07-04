package virtualoutbound

import (
	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func VirtualOutboundOnK8s() {
	var k8sCluster Cluster

	BeforeEach(func() {
		t := NewTestingT()
		k8sCluster = NewK8sCluster(t, Kuma1, Silent)

		err := NewClusterSetup().
			Install(Kuma(config_core.Standalone,
				WithEnv("KUMA_DNS_SERVER_SERVICE_VIP_ENABLED", "false"),
			)).
			Install(NamespaceWithSidecarInjection(TestNamespace)).
			Install(DemoClientK8s("default", TestNamespace)).
			Install(testserver.Install(testserver.WithStatefulSet(true), testserver.WithReplicas(2))).
			Setup(k8sCluster)
		if err != nil {
			lines := int64(50)
			k8s.GetPodLogs(t, k8sCluster.GetKubectlOptions(TestNamespace), "kuma-control-plane", &v1.PodLogOptions{TailLines: &lines})
		}
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterEach(func() {
		Expect(k8sCluster.DeleteKuma()).To(Succeed())
		Expect(k8sCluster.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(k8sCluster.DismissCluster()).To(Succeed())
	})

	It("doesn't support default vips", func() {
		virtualOutboundAll := `
apiVersion: kuma.io/v1alpha1
kind: VirtualOutbound
mesh: default
metadata:
  name: instance
spec:
  selectors:
  - match:
      kuma.io/service: "*"
  conf:
    host: "{{.svc}}.foo"
    port: "8080"
    parameters:
    - name: "svc"
      tagKey: "kuma.io/service"
`
		err := YamlK8s(virtualOutboundAll)(k8sCluster)
		Expect(err).ToNot(HaveOccurred())
		// when client sends requests to server
		clientPodName, err := PodNameOfApp(k8sCluster, "demo-client", TestNamespace)
		Expect(err).ToNot(HaveOccurred())

		// Succeed with virtual-outbound
		stdout, stderr, err := k8sCluster.ExecWithRetries(TestNamespace, clientPodName, "demo-client",
			"curl", "-v", "-m", "3", "--fail", "test-server_kuma-test_svc_80.foo:8080")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))
		Expect(stdout).To(ContainSubstring(`"instance":"test-server`))

		// Fails with built in vip (it's disabled in conf)
		_, _, err = k8sCluster.Exec(TestNamespace, clientPodName, "demo-client",
			"curl", "-v", "-m", "3", "--fail", "test-server_kuma-test_svc_80.mesh:80")
		Expect(err).To(HaveOccurred())
	})

	It("virtual outbounds on statefulSet", func() {
		virtualOutboundAll := `
apiVersion: kuma.io/v1alpha1
kind: VirtualOutbound
mesh: default
metadata:
  name: instance
spec:
  selectors:
  - match:
      kuma.io/service: "*"
      statefulset.kubernetes.io/pod-name: "*"
  conf:
    host: "{{.svc}}.{{.inst}}"
    port: "8080"
    parameters:
    - name: "svc"
      tagKey: "kuma.io/service"
    - name: "inst"
      tagKey: "statefulset.kubernetes.io/pod-name"
`
		err := YamlK8s(virtualOutboundAll)(k8sCluster)
		Expect(err).ToNot(HaveOccurred())
		// when client sends requests to server
		clientPodName, err := PodNameOfApp(k8sCluster, "demo-client", TestNamespace)
		Expect(err).ToNot(HaveOccurred())

		stdout, stderr, err := k8sCluster.ExecWithRetries(TestNamespace, clientPodName, "demo-client",
			"curl", "-v", "-m", "3", "--fail", "test-server_kuma-test_svc_80.test-server-0:8080")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))
		Expect(stdout).To(ContainSubstring(`"instance":"test-server-0"`))

		stdout, stderr, err = k8sCluster.ExecWithRetries(TestNamespace, clientPodName, "demo-client",
			"curl", "-v", "-m", "3", "--fail", "test-server_kuma-test_svc_80.test-server-1:8080")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))
		Expect(stdout).To(ContainSubstring(`"instance":"test-server-1"`))
	})
}
