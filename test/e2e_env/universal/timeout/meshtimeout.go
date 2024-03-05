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

func PluginTest() {
	meshName := "meshtimeout"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshUniversal(meshName)).
			Install(DemoClientUniversal("demo-client", meshName,
				WithTransparentProxy(true)),
			).
			Install(TestServerUniversal("test-server", meshName,
				WithArgs([]string{"echo", "--instance", "universal-1"})),
			).
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// Delete the default meshretry policy
		Eventually(func() error {
			return universal.Cluster.GetKumactlOptions().RunKumactl("delete", "meshretry", "--mesh", meshName, "mesh-retry-all-"+meshName)
		}).Should(Succeed())
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})
	E2EAfterEach(func() {
		Expect(universal.Cluster.GetKumactlOptions().KumactlDelete("meshtimeout", "default", meshName)).To(Succeed())
	})

	DescribeTable("should reset the connection by timeout", func(timeoutConfig string) {
		By("check requests take over 5s")
		Eventually(func(g Gomega) {
			start := time.Now()
			_, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client", "test-server.mesh",
				client.WithHeader("x-set-response-delay-ms", "3000"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(time.Since(start)).To(BeNumerically(">", time.Second*3))
		}).Should(Succeed())

		By("apply a new policy")
		Expect(universal.Cluster.Install(YamlUniversal(timeoutConfig))).To(Succeed())

		By("eventually requests timeout consistently")
		Eventually(func(g Gomega) {
			response, err := client.CollectFailure(
				universal.Cluster, "demo-client", "test-server.mesh",
				client.WithHeader("x-set-response-delay-ms", "3000"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.ResponseCode).To(Equal(504))
		}).WithTimeout(15 * time.Second).Should(Succeed())
	},
		Entry("outbound timeout", fmt.Sprintf(`
type: MeshTimeout
name: default
mesh: %s
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshService
        name: test-server
      default:
        connectionTimeout: 20s
        http:
          requestTimeout: 1s`, meshName)),
		Entry("inbound timeout", fmt.Sprintf(`
type: MeshTimeout
name: default
mesh: %s
spec:
  targetRef:
    kind: Mesh
  from:
    - targetRef:
        kind: Mesh
      default:
        connectionTimeout: 20s
        http:
          requestTimeout: 1s`, meshName)),
	)
}
