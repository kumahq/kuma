package matching

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func Matching() {
	const mesh = "matching"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshUniversal(mesh)).
			Install(DemoClientUniversal("demo-client-1", mesh, WithTransparentProxy(true))).
			Install(DemoClientUniversal("demo-client-2", mesh, WithTransparentProxy(true))).
			Install(TestServerUniversal("test-server", mesh,
				WithTransparentProxy(true),
				WithArgs([]string{"echo", "--instance", "echo-v1"})),
			).
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(mesh)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(mesh)).To(Succeed())
	})

	// Added Flake because: https://github.com/kumahq/kuma/issues/4700
	It("should both fault injections with the same destination proxy", FlakeAttempts(3), func() {
		Expect(YamlUniversal(fmt.Sprintf(`
type: FaultInjection
mesh: %s
name: fi1
sources:
   - match:
       kuma.io/service: demo-client-1
destinations:
   - match:
       kuma.io/service: test-server
       kuma.io/protocol: http
conf:
   abort:
     httpStatus: 401
     percentage: 100`, mesh))(universal.Cluster)).To(Succeed())

		Expect(YamlUniversal(fmt.Sprintf(`
type: FaultInjection
mesh: %s
name: fi2
sources:
   - match:
       kuma.io/service: demo-client-2
destinations:
   - match:
       kuma.io/service: test-server
       kuma.io/protocol: http
conf:
   abort:
     httpStatus: 402
     percentage: 100`, mesh))(universal.Cluster)).To(Succeed())

		Eventually(func(g Gomega) {
			response, err := client.CollectFailure(
				universal.Cluster, "demo-client-1", "test-server.mesh",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.ResponseCode).To(Equal(401))
		}, "60s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			response, err := client.CollectFailure(
				universal.Cluster, "demo-client-2", "test-server.mesh",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.ResponseCode).To(Equal(402))
		}, "60s", "1s").Should(Succeed())
	})
}
