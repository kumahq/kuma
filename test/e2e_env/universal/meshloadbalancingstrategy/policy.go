package meshloadbalancingstrategy

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func Policy() {
	meshName := "mesh-load-balancing-strategy"

	BeforeEach(func() {
		Expect(NewClusterSetup().
			Install(MeshUniversal(meshName)).
			Install(TestServerUniversal("test-server-1", meshName,
				WithArgs([]string{"echo", "--instance", "universal-1"}))).
			Install(TestServerUniversal("test-server-2", meshName,
				WithArgs([]string{"echo", "--instance", "universal-2"}))).
			Install(TestServerUniversal("test-server-3", meshName,
				WithArgs([]string{"echo", "--instance", "universal-3"}))).
			Install(DemoClientUniversal("demo-client", meshName,
				WithTransparentProxy(true))).
			Setup(universal.Cluster)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should use ring hash load balancing strategy", func() {
		Eventually(func(g Gomega) {
			responses, err := client.CollectResponsesByInstance(
				universal.Cluster, "demo-client", "test-server.mesh")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses).To(HaveLen(3))
		}, "30s", "500ms").Should(Succeed())

		Expect(YamlUniversal(`
type: MeshLoadBalancingStrategy
name: ring-hash
mesh: mesh-load-balancing-strategy
spec:
  targetRef:
    kind: MeshService
    name: demo-client
  to:
    - targetRef:
        kind: MeshService
        name: test-server
      default:
        loadBalancer:
          type: RingHash
          ringHash:
            hashPolicies:
              - type: Header
                header:
                  name: x-header
`)(universal.Cluster)).To(Succeed())

		Eventually(func(g Gomega) {
			responses, err := client.CollectResponsesByInstance(
				universal.Cluster, "demo-client", "test-server.mesh",
				client.WithHeader("x-header", "value"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses).To(HaveLen(1))
		}, "30s", "500ms").Should(Succeed())
	})
}
