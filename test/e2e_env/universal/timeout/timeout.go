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
	faultInjection := fmt.Sprintf(`
type: FaultInjection
mesh: "%s"
name: fi1
sources:
   - match:
       kuma.io/service: demo-client
destinations:
   - match:
       kuma.io/service: test-server
       kuma.io/protocol: http
conf:
   delay:
     percentage: 100
     value: 5s
`, meshName)

	timeout := fmt.Sprintf(`
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
    requestTimeout: 2s
`, meshName)

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshUniversal(meshName)).
			Install(YamlUniversal(faultInjection)).
			Install(DemoClientUniversal("demo-client", meshName,
				WithTransparentProxy(true)),
			).
			Install(TestServerUniversal("test-server", meshName,
				WithArgs([]string{"echo", "--instance", "universal-1"})),
			).
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})
	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should reset the connection by timeout", func() {
		By("check requests take over 5s")
		Eventually(func(g Gomega) {
			start := time.Now()
			_, err := client.CollectResponse(
				universal.Cluster, "demo-client", "test-server.mesh",
				client.WithMaxTime(10),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(time.Since(start)).To(BeNumerically(">", time.Second*5))
		}).Should(Succeed())

		By("apply a new policy")
		Expect(universal.Cluster.Install(YamlUniversal(timeout))).To(Succeed())

		By("eventually requests timeout consistently")
		expectation := func(g Gomega) {
			response, err := client.CollectFailure(
				universal.Cluster, "demo-client", "test-server.mesh",
				client.WithMaxTime(10),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.ResponseCode).To(Equal(504))
		}
		Eventually(expectation).Should(Succeed())
		Consistently(expectation).Should(Succeed())
	})
}
