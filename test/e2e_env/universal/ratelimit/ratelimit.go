package ratelimit

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/universal/env"
	. "github.com/kumahq/kuma/test/framework"
)

func Policy() {
	var meshName = "rate-limit"
	var rateLimitPolicy = fmt.Sprintf(`
type: RateLimit
mesh: "%s"
name: rate-limit-all-sources
sources:
- match:
   kuma.io/service: "*"
destinations:
- match:
   kuma.io/service: test-server
conf:
  http:
    requests: 1
    interval: 10s
    onRateLimit:
      status: 429
      headers:
        - key: "x-kuma-rate-limited"
          value: "true"`, meshName)
	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(MeshUniversal(meshName)).
			Install(YamlUniversal(rateLimitPolicy)).
			Install(TestServerUniversal("test-server", meshName, WithArgs([]string{"echo", "--instance", "universal-1"}))).
			Install(DemoClientUniversal("demo-client", meshName, WithTransparentProxy(true))).
			Install(DemoClientUniversal("web", meshName, WithTransparentProxy(true))).
			Install(TestServerExternalServiceUniversal("rate-limit", meshName, 80, false)).
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

	It("should limit per source", func() {
		specificRateLimitPolicy := fmt.Sprintf(`
type: RateLimit
mesh: "%s"
name: rate-limit-demo-client
sources:
- match:
    kuma.io/service: "demo-client"
destinations:
- match:
    kuma.io/service: test-server
conf:
  http:
    onRateLimit:
      status: 400
    requests: 1
    interval: 10s
`, meshName)
		Expect(env.Cluster.Install(YamlUniversal(specificRateLimitPolicy))).To(Succeed())

		By("demo-client specific RateLimit works")
		Eventually(requestRateLimited("demo-client", "test-server", "400"), "10s", "100ms").Should(Succeed())

		By("catch-all RateLimit works")
		Eventually(requestRateLimited("web", "test-server", "429"), "10s", "100ms").Should(Succeed())
	})

	// Added Flake because: https://github.com/kumahq/kuma/issues/4700
	It("should limit echo server as external service", FlakeAttempts(3), func() {
		externalService := fmt.Sprintf(`
type: ExternalService
mesh: "%s"
name: external-service
tags:
  kuma.io/service: external-service
  kuma.io/protocol: http
networking:
  address: "%s_externalservice-rate-limit:88"
`, meshName, env.Cluster.Name())
		specificRateLimitPolicy := fmt.Sprintf(`
type: RateLimit
mesh: "%s"
name: rate-limit-demo-client
sources:
- match:
    kuma.io/service: "demo-client"
destinations:
- match:
    kuma.io/service: "external-service"
conf:
  http:
    onRateLimit:
      status: 429
    requests: 1
    interval: 10s
`, meshName)

		By("Exposing external service and specific rate limit")
		Expect(env.Cluster.Install(YamlUniversal(externalService))).To(Succeed())
		Expect(env.Cluster.Install(YamlUniversal(specificRateLimitPolicy))).To(Succeed())

		By("demo-client specific RateLimit works")
		Eventually(requestRateLimited("demo-client", "external-service", "429"), "10s", "100ms").Should(Succeed())

		By("RateLimit doesn't apply for other clients")
		Consistently(requestRateLimited("web", "external-service", "429"), "10s", "100ms").ShouldNot(Succeed())
	})
}
