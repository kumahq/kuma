package reachableservices

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func ReachableServices() {
	meshName := "reachable-svc"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshUniversal(meshName)).
			Install(TestServerUniversal("first-test-server", meshName, WithArgs([]string{"echo"}), WithServiceName("first-test-server"))).
			Install(TestServerUniversal("second-test-server", meshName, WithArgs([]string{"echo"}), WithServiceName("second-test-server"))).
			Install(DemoClientUniversal(AppModeDemoClient, meshName, WithTransparentProxy(true), WithReachableServices("first-test-server"))).
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should be able to connect to reachable services", func() {
		Eventually(func(g Gomega) {
			// when
			_, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client", "first-test-server.mesh",
			)
			// then
			g.Expect(err).ToNot(HaveOccurred())
		}).Should(Succeed())
	})

	// Disabled because of flakiness: https://github.com/kumahq/kuma/issues/9349
	XIt("should not be able to non reachable services", func() {
		Consistently(func(g Gomega) {
			// when
			response, err := client.CollectFailure(
				universal.Cluster, "demo-client", "second-test-server.mesh",
			)
			// then
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Exitcode).To(Equal(6))
		}).Should(Succeed())
	})
}
