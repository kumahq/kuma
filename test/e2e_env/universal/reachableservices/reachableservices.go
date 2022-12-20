package reachableservices

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/universal/env"
	. "github.com/kumahq/kuma/test/framework"
)

func ReachableServices() {
	meshName := "reachable-svc"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshUniversal(meshName)).
			Install(TestServerUniversal("first-test-server", meshName, WithArgs([]string{"echo"}), WithServiceName("first-test-server"))).
			Install(TestServerUniversal("second-test-server", meshName, WithArgs([]string{"echo"}), WithServiceName("second-test-server"))).
			Install(DemoClientUniversal(AppModeDemoClient, meshName, WithTransparentProxy(true), WithReachableServices("first-test-server"))).
			Setup(env.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterAll(func() {
		Expect(env.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should be able to connect to reachable services", func() {
		Eventually(func(g Gomega) {
			// when
			stdout, _, err := env.Cluster.Exec("", "", "demo-client",
				"curl", "-v", "--fail", "first-test-server.mesh")

			// then
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
		}).Should(Succeed())
	})

	It("should not be able to non reachable services", func() {
		Consistently(func(g Gomega) {
			// when
			_, _, err := env.Cluster.Exec("", "", "demo-client",
				"curl", "-v", "--fail", "second-test-server.mesh")

			// then
			g.Expect(err).To(HaveOccurred())
		}).Should(Succeed())
	})
}
