package meshratelimit

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/universal/env"
	. "github.com/kumahq/kuma/test/framework"
)

func Policy() {
	var meshName = "mesh-rate-limit"
	var rateLimitPolicy = fmt.Sprintf(`
type: MeshRateLimit
mesh: "%s"
name: mesh-rate-limit-all-sources
spec:
  targetRef:
    kind: MeshService
    name: test-server
  from:
    - targetRef:
        kind: Mesh
      default:
        local:
          http:
            requests: 1
            interval: 10s
            onRateLimit:
              status: 429
              headers:
                - key: "x-kuma-rate-limited"
                  value: "true"
          tcp:
            connections: 1
            interval: 10s`, meshName)
	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(MeshUniversal(meshName)).
			Install(YamlUniversal(rateLimitPolicy)).
			Install(TestServerUniversal("test-server", meshName, WithArgs([]string{"echo", "--instance", "universal-1"}))).
			Install(DemoClientUniversal("demo-client", meshName, WithTransparentProxy(true))).
			Install(DemoClientUniversal("tcp-client", meshName, WithTransparentProxy(false))).
			Install(DemoClientUniversal("web", meshName, WithTransparentProxy(true))).
			Setup(env.Cluster)).To(Succeed())
	})
	E2EAfterAll(func() {
		Expect(env.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(meshName)).To(Succeed())
	})
	requestRateLimited := func(client string, svc string, status string) func(g Gomega) {
		return func(g Gomega) {
			stdout, _, err := env.Cluster.Exec("", "", client, "curl", "-v", fmt.Sprintf("%s.mesh", svc))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).Should(ContainSubstring(status))
		}
	}
	keepConnectionOpen := func() {
		// Open TCP connections to the test-server
		defer GinkgoRecover()
		ip := env.Cluster.GetApp("test-server").GetIP()
		_, _, _ = env.Cluster.Exec("", "", "tcp-client", "telnet", ip, "80")
	}

	It("should limit all sources", func() {
		By("demo-client to test-server should be ratelimited by catch-all")
		Eventually(requestRateLimited("demo-client", "test-server", "429"), "10s", "100ms").Should(Succeed())

		By("web to test-server should be ratelimited by catch-all")
		Eventually(requestRateLimited("web", "test-server", "429"), "10s", "100ms").Should(Succeed())
	})

	It("should limit tcp connections", func() {
		// open connection
		go keepConnectionOpen()

		// should return 503 when number of connections is exceeded
		Eventually(requestRateLimited("web", "test-server", "503"), "10s", "100ms").Should(Succeed())
	})
}
