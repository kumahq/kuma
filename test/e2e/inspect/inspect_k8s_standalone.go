package inspect

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

func KubernetesStandalone() {
	var cluster *K8sCluster
	var demoClientName string

	BeforeEach(func() {
		cluster = NewK8sCluster(NewTestingT(), Kuma1, Silent)

		err := NewClusterSetup().
			Install(Kuma(config_core.Standalone, WithVerbose())).
			Install(NamespaceWithSidecarInjection(TestNamespace)).
			Install(DemoClientK8s("default", TestNamespace)).
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

		podName, err := PodNameOfApp(cluster, "demo-client", TestNamespace)
		Expect(err).ToNot(HaveOccurred())
		demoClientName = podName
	})

	E2EAfterEach(func() {
		Expect(cluster.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(cluster.DeleteKuma()).To(Succeed())
		Expect(cluster.DismissCluster()).ToNot(HaveOccurred())
	})

	It("should return envoy config_dump", func() {
		dataplaneName := fmt.Sprintf("%s.%s", demoClientName, TestNamespace)
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
