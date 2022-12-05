package timeout

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/universal/env"
	. "github.com/kumahq/kuma/test/framework"
)

func PluginTest() {
	meshName := "meshtimeout"
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
			Setup(env.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})
	E2EAfterAll(func() {
		Expect(env.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should reset the connection by timeout", func() {
		By("Checking requests succeed")
		Eventually(func(g Gomega) {
			stdout, _, err := env.Cluster.Exec("", "", "demo-client",
				"curl", "-v", "--fail", "test-server.mesh")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
		}).Should(Succeed())

		By("check requests take over 5s")
		Eventually(func(g Gomega) {
			start := time.Now()
			stdout, _, err := env.Cluster.Exec("", "", "demo-client",
				"curl", "-v", "--fail", "test-server.mesh")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
			g.Expect(time.Since(start)).To(BeNumerically(">", time.Second*5))
		}).Should(Succeed())

		By("apply a new policy")
		Expect(env.Cluster.Install(YamlUniversal(timeout))).To(Succeed())

		By("eventually requests timeout consistently")
		Eventually(func(g Gomega) {
			stdout, _, err := env.Cluster.Exec("", "", "demo-client",
				"curl", "-v", "test-server.mesh")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring("upstream request timeout"))
		}).Should(Succeed())
	})
}
