package inspect

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

func KubernetesStandalone() {
	var cluster *K8sCluster
	var demoClient *kube_core.Pod

	GetPod := func(namespace, app string) *kube_core.Pod {
		pods, err := k8s.ListPodsE(
			cluster.GetTesting(),
			cluster.GetKubectlOptions(namespace),
			metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", app),
			},
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(pods).To(HaveLen(1))

		return &pods[0]
	}

	BeforeEach(func() {
		cluster = NewK8sCluster(NewTestingT(), Kuma1, Silent)

		err := NewClusterSetup().
			Install(Kuma(config_core.Standalone, WithVerbose())).
			Install(NamespaceWithSidecarInjection(TestNamespace)).
			Install(DemoClientK8s("default")).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	// Before each test, verify that we have the Dataplanes that we expect to need.
	JustBeforeEach(func() {
		Expect(cluster.VerifyKuma()).To(Succeed())

		// Synchronize on the dataplanes coming up.
		Eventually(func(g Gomega) {
			dataplanes, err := cluster.GetKumactlOptions().KumactlList("dataplanes", "default")
			g.Expect(err).ToNot(HaveOccurred())
			// Dataplane names are generated, so we check for a partial match.
			g.Expect(dataplanes).Should(ContainElement(ContainSubstring("demo-client")))
		}, "60s", "1s").Should(Succeed())

		demoClient = GetPod(TestNamespace, "demo-client")
	})

	E2EAfterEach(func() {
		Expect(cluster.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(cluster.DeleteKuma()).To(Succeed())
		Expect(cluster.DismissCluster()).ToNot(HaveOccurred())
	})

	It("should return envoy config_dump", func() {
		dataplaneName := fmt.Sprintf("%s.%s", demoClient.GetName(), TestNamespace)
		stdout, err := cluster.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "dataplane", dataplaneName, "--config-dump")
		Expect(err).ToNot(HaveOccurred())

		Expect(stdout).To(ContainSubstring(`"name": "demo-client_kuma-test_svc"`))
		Expect(stdout).To(ContainSubstring(`"name": "inbound:passthrough:ipv4"`))
		Expect(stdout).To(ContainSubstring(`"name": "inbound:passthrough:ipv6"`))
		Expect(stdout).To(ContainSubstring(`"name": "kuma:envoy:admin"`))
		Expect(stdout).To(ContainSubstring(`"name": "outbound:passthrough:ipv4"`))
		Expect(stdout).To(ContainSubstring(`"name": "outbound:passthrough:ipv6"`))
	})
}
