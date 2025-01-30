package retry

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func Policy() {
	meshName := "retry"
	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshUniversal(meshName)).
			Install(DemoClientUniversal("demo-client", meshName, WithTransparentProxy(true))).
			Install(TestServerUniversal("test-server", meshName, WithArgs([]string{"echo", "--instance", "universal"}))).
			Install(TrafficRouteUniversal(meshName)).
			Install(TrafficPermissionUniversal(meshName)).
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// Delete the default meshretry policy
		Eventually(func() error {
			return universal.Cluster.GetKumactlOptions().RunKumactl("delete", "meshretry", "--mesh", meshName, "mesh-retry-all-"+meshName)
		}).Should(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, meshName)
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should retry on HTTP connection failure", func() {
		fiPolicy := fmt.Sprintf(`
type: FaultInjection
mesh: "%s"
name: fi-retry
sources:
   - match:
       kuma.io/service: demo-client
destinations:
   - match:
       kuma.io/service: test-server
       kuma.io/protocol: http
conf:
   abort:
     httpStatus: 500
     percentage: 50`, meshName)
		retryPolicy := fmt.Sprintf(`
type: Retry
mesh: "%s"
name: fake-retry-policy
sources:
- match:
    kuma.io/service: demo-client
destinations:
- match:
    kuma.io/service: test-server
conf:
  http:
    numRetries: 5
`, meshName)

		By("Checking requests succeed")
		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client", "test-server.mesh",
			)
			g.Expect(err).ToNot(HaveOccurred())
		}, "30s", "1s").MustPassRepeatedly(3).Should(Succeed())

		By("Adding a fault injection")
		Expect(universal.Cluster.Install(YamlUniversal(fiPolicy))).To(Succeed())

		By("Check some errors happen")
		Eventually(func(g Gomega) {
			response, err := client.CollectFailure(
				universal.Cluster, "demo-client", "test-server.mesh",
				client.NoFail(),
				client.OutputFormat(`{ "received": { "status": %{response_code} } }`),
			)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.ResponseCode).To(Equal(500))
		}, "30s", "100ms").Should(Succeed())

		By("Apply a retry policy")
		Expect(universal.Cluster.Install(YamlUniversal(retryPolicy))).To(Succeed())

		By("Eventually all requests succeed consistently")
		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client", "test-server.mesh",
			)
			g.Expect(err).ToNot(HaveOccurred())
		}, "30s", "1s").MustPassRepeatedly(3).Should(Succeed())
	})
}
