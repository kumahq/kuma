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

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshUniversal(meshName)).
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
	E2EAfterEach(func() {
		Expect(env.Cluster.GetKumactlOptions().KumactlDelete("meshtimeout", "default", meshName)).To(Succeed())
	})

	DescribeTable("should reset the connection by timeout", func(timeoutConfig string) {
		By("check requests take over 5s")
		Eventually(func(g Gomega) {
			start := time.Now()
			stdout, _, err := env.Cluster.Exec("", "", "demo-client",
				"curl", "-v", "-H", "\"x-set-response-delay-ms: 3000\"", "--fail", "test-server.mesh")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
			g.Expect(time.Since(start)).To(BeNumerically(">", time.Second*3))
		}).Should(Succeed())

		By("apply a new policy")
		Expect(env.Cluster.Install(YamlUniversal(timeoutConfig))).To(Succeed())

		By("eventually requests timeout consistently")
		Eventually(func(g Gomega) {
			stdout, _, err := env.Cluster.Exec("", "", "demo-client",
				"curl", "-v", "-H", "\"x-set-response-delay-ms: 3000\"", "test-server.mesh")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring("upstream request timeout"))
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
