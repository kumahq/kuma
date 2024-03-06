package timeout

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func Policy() {
	meshName := "timeout"
	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshUniversal(meshName)).
			Install(TimeoutUniversal(meshName)).
			Install(TrafficRouteUniversal(meshName)).
			Install(DemoClientUniversal("demo-client", meshName,
				WithTransparentProxy(true)),
			).
			Install(TestServerUniversal("test-server", meshName,
				WithArgs([]string{"echo", "--instance", "universal-1"})),
			).
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// Delete the default meshtimeout policy
		Eventually(func() error {
			return universal.Cluster.GetKumactlOptions().RunKumactl("delete", "meshtimeout", "--mesh", meshName, "mesh-timeout-all-"+meshName)
		}).Should(Succeed())
	})
	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should reset the connection by timeout", func() {
		By("check requests take over 2s")
		Eventually(func(g Gomega) {
			start := time.Now()
			_, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client", "test-server.mesh",
				client.WithHeader("x-set-response-delay-ms", "2000"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(time.Since(start)).To(BeNumerically(">", time.Second*2))
		}).Should(Succeed())

		By("apply a new policy")
		err := universal.Cluster.Install(YamlUniversal(fmt.Sprintf(`
type: Timeout
mesh: "%s"
name: echo-service-timeouts
sources:
- match:
    kuma.io/service: '*'
destinations:
- match:
    kuma.io/service: test-server
conf:
  connectTimeout: 10s
  http:
    requestTimeout: 1s
`, meshName)))
		Expect(err).ToNot(HaveOccurred())

		By("eventually requests timeout consistently")
		Eventually(func(g Gomega) {
			response, err := client.CollectFailure(
				universal.Cluster, "demo-client", "test-server.mesh",
				client.WithHeader("x-set-response-delay-ms", "2000"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.ResponseCode).To(Equal(504))
		}, "30s", "1s").MustPassRepeatedly(3).Should(Succeed())
	})
}
