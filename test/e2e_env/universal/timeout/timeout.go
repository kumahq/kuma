package timeout

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
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
		By("Checking requests succeed")
		Eventually(func(g Gomega) {
			stdout, _, err := universal.Cluster.Exec("", "", "demo-client",
				"curl", "-v", "--fail", "test-server.mesh")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
		}).Should(Succeed())

		By("check requests take over 5s")
		start := time.Now()
		_, _, err := universal.Cluster.Exec("", "", "demo-client",
			"curl", "-v", "--fail", "test-server.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(time.Since(start)).To(BeNumerically(">", time.Second*5))

		By("apply a new policy")
		Expect(universal.Cluster.Install(YamlUniversal(timeout))).To(Succeed())

		By("eventually requests timeout consistently")
		Eventually(func() string {
			stdout, _, _ := universal.Cluster.Exec("", "", "demo-client",
				"curl", "-v", "test-server.mesh")
			return stdout
		}).Should(ContainSubstring("upstream request timeout"))
		Consistently(func() string {
			stdout, _, _ := universal.Cluster.Exec("", "", "demo-client",
				"curl", "-v", "test-server.mesh")
			return stdout
		}).Should(ContainSubstring("upstream request timeout"))
	})
}
