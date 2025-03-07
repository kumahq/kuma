package meshratelimit

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envoy_admin"
	"github.com/kumahq/kuma/test/framework/envoy_admin/stats"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func Policy() {
	meshName := "mesh-rate-limit"
	rateLimitPolicy := fmt.Sprintf(`
type: MeshRateLimit
mesh: "%s"
name: mesh-rate-limit-all-sources
spec:
  targetRef:
    kind: MeshService
    name: test-server
  rules:
    - default:
        local:
          http:
            requestRate:
              num: 1
              interval: 10s
            onRateLimit:
              status: 429
              headers:
                add:
                - name: "x-kuma-rate-limited"
                  value: "true"`, meshName)
	rateLimitPolicyTcp := fmt.Sprintf(`
type: MeshRateLimit
mesh: "%s"
name: mesh-rate-limit-tcp
spec:
  targetRef:
    kind: MeshService
    name: test-server-tcp
  rules:
    - default:
        local:
          tcp:
            connectionRate: 
              num: 1
              interval: 10s`, meshName)
	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(MeshUniversal(meshName)).
			Install(YamlUniversal(rateLimitPolicy)).
			Install(YamlUniversal(rateLimitPolicyTcp)).
			Install(TestServerUniversal("test-server", meshName, WithArgs([]string{"echo", "--instance", "universal-1"}))).
			Install(TestServerUniversal("test-server-tcp", meshName,
				WithArgs([]string{"echo", "--instance", "universal-2"}),
				WithServiceName("test-server-tcp"))).
			Install(DemoClientUniversal("demo-client", meshName, WithTransparentProxy(true))).
			Install(DemoClientUniversal("tcp-client", meshName, WithTransparentProxy(false))).
			Install(DemoClientUniversal("web", meshName, WithTransparentProxy(true))).
			Setup(universal.Cluster)).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, meshName)
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})
	requestRateLimited := func(container string, svc string, responseCode int) func(g Gomega) {
		return func(g Gomega) {
			response, err := client.CollectFailure(
				universal.Cluster, container, fmt.Sprintf("%s.mesh", svc),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.ResponseCode).Should(Equal(responseCode))
		}
	}
	keepConnectionOpen := func() {
		// Open TCP connections to the test-server
		defer GinkgoRecover()
		ip := universal.Cluster.GetApp("test-server-tcp").GetIP()
		_, _, _ = universal.Cluster.Exec("", "", "tcp-client", "telnet", ip, "80")
	}

	tcpRateLimitStats := func(admin envoy_admin.Tunnel) *stats.Stats {
		s, err := admin.GetStats("local_rate_limit.tcp_rate_limit.rate_limited")
		Expect(err).ToNot(HaveOccurred())
		return s
	}

	It("should limit all sources", func() {
		By("demo-client to test-server should be rate limited by mesh-rate-limit-all-sources")
		Eventually(requestRateLimited("demo-client", "test-server", 429), "10s", "100ms").Should(Succeed())

		By("web to test-server should be rate limited by mesh-rate-limit-all-sources")
		Eventually(requestRateLimited("web", "test-server", 429), "10s", "100ms").Should(Succeed())
	})

	It("should limit tcp connections", func() {
		admin := universal.Cluster.GetApp("test-server-tcp").GetEnvoyAdminTunnel()
		// should have no ratelimited connections
		Expect(tcpRateLimitStats(admin)).To(stats.BeEqualZero())

		// open connection
		go keepConnectionOpen()

		// should return 503 when number of connections is exceeded
		Eventually(requestRateLimited("web", "test-server-tcp", 503), "10s", "100ms").Should(Succeed())
		// and stats should increase
		Expect(tcpRateLimitStats(admin)).To(stats.BeGreaterThanZero())
	})
}
