package retry

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
	meshName := "retry"
	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshUniversal(meshName)).
			Install(DemoClientUniversal("demo-client", meshName, WithTransparentProxy(true))).
			Install(TestServerUniversal("test-server", meshName, WithArgs([]string{"echo", "--instance", "universal"}))).
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// Delete the default retry policy
		Eventually(func() error {
			return universal.Cluster.GetKumactlOptions().RunKumactl("delete", "retry", "--mesh", meshName, "retry-all-"+meshName)
		}).Should(Succeed())
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should retry on HTTP connection failure", func() {
		echoServerDataplane := fmt.Sprintf(`
type: Dataplane
mesh: "%s"
name: fake-echo-server
networking:
  address:  241.0.0.1
  inbound:
  - port: 7777
    servicePort: 7777
    tags:
      kuma.io/service: test-server
      kuma.io/protocol: http
`, meshName)
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
			_, err := client.CollectResponse(
				universal.Cluster, "demo-client", "test-server.mesh",
			)
			g.Expect(err).ToNot(HaveOccurred())
		}).Should(Succeed())
		Consistently(func(g Gomega) {
			// -m 8 to wait for 8 seconds to beat the default 5s connect timeout
			_, err := client.CollectResponse(
				universal.Cluster, "demo-client", "test-server.mesh",
				client.WithMaxTime(8),
			)
			g.Expect(err).ToNot(HaveOccurred())
		})

		By("Adding a faulty dataplane")
		Expect(universal.Cluster.Install(YamlUniversal(echoServerDataplane))).To(Succeed())

		By("Check some errors happen")
		var errs []error
		for i := 0; i < 50; i++ {
			time.Sleep(time.Millisecond * 100)
			_, err := client.CollectResponse(
				universal.Cluster, "demo-client", "test-server.mesh",
				client.WithMaxTime(8),
			)
			if err != nil {
				errs = append(errs, err)
			}
		}
		Expect(errs).ToNot(BeEmpty())

		By("Apply a retry policy")
		Expect(universal.Cluster.Install(YamlUniversal(retryPolicy))).To(Succeed())

		By("Eventually all requests succeed consistently")
		Eventually(func(g Gomega) {
			_, err := client.CollectResponse(
				universal.Cluster, "demo-client", "test-server.mesh",
				client.WithMaxTime(8),
			)
			g.Expect(err).ToNot(HaveOccurred())
		}).Should(Succeed())
		Consistently(func(g Gomega) {
			// -m 8 to wait for 8 seconds to beat the default 5s connect timeout
			_, err := client.CollectResponse(
				universal.Cluster, "demo-client", "test-server.mesh",
				client.WithMaxTime(8),
			)
			g.Expect(err).ToNot(HaveOccurred())
		})
	})
}
