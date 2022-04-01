package reachableservices

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

func ReachableServicesOnUniversal() {
	var cluster Cluster

	BeforeEach(func() {
		cluster = NewUniversalCluster(NewTestingT(), Kuma1, Silent)

		err := NewClusterSetup().
			Install(Kuma(config_core.Standalone)).
			Install(TestServerUniversal("first-test-server", "default", WithArgs([]string{"echo"}), WithServiceName("first-test-server"))).
			Install(TestServerUniversal("second-test-server", "default", WithArgs([]string{"echo"}), WithServiceName("second-test-server"))).
			Install(DemoClientUniversal(AppModeDemoClient, "default", WithTransparentProxy(true), WithReachableServices("first-test-server"))).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterEach(func() {
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	It("should be able to connect only to reachable services", func() {
		// when tries to connect to a reachable service
		stdout, _, err := cluster.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "--fail", "first-test-server.mesh")

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))

		// when trying to connect to non-reachable services via Kuma DNS
		_, _, err = cluster.Exec("", "", "demo-client",
			"curl", "-v", "--fail", "second-test-server.mesh")

		// then it fails because Kuma DP has no such VIP
		Expect(err).To(HaveOccurred())
	})
}
