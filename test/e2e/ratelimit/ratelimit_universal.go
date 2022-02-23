package ratelimit

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/externalservice"
)

var cluster Cluster

var rateLimitPolicy = `
type: RateLimit
mesh: default
name: rate-limit-all-sources
sources:
- match:
   kuma.io/service: "*"
destinations:
- match:
   kuma.io/service: test-server
conf:
  http:
    requests: 2
    interval: 10s
    onRateLimit:
      status: 423
      headers:
        - key: "x-kuma-rate-limited"
          value: "true"
`

var _ = E2EBeforeSuite(func() {
	clusters, err := NewUniversalClusters(
		[]string{Kuma3},
		Silent)
	Expect(err).ToNot(HaveOccurred())

	// Global
	cluster = clusters.GetCluster(Kuma3)

	Expect(NewClusterSetup().
		Install(Kuma(core.Standalone)).
		Setup(cluster)).To(Succeed())

	demoClientToken, err := cluster.GetKuma().GenerateDpToken("default", "demo-client")
	Expect(err).ToNot(HaveOccurred())

	testServerToken, err := cluster.GetKuma().GenerateDpToken("default", "test-server")
	Expect(err).ToNot(HaveOccurred())

	webToken, err := cluster.GetKuma().GenerateDpToken("default", "web")
	Expect(err).ToNot(HaveOccurred())

	Expect(YamlUniversal(rateLimitPolicy)(cluster)).To(Succeed())

	Expect(NewClusterSetup().
		Install(TestServerUniversal("test-server", "default", testServerToken, WithArgs([]string{"echo", "--instance", "universal-1"}))).
		Install(DemoClientUniversal(AppModeDemoClient, "default", demoClientToken, WithTransparentProxy(true))).
		Install(DemoClientUniversal("web", "default", webToken, WithTransparentProxy(true))).
		Install(externalservice.Install(externalservice.HttpServer, externalservice.UniversalAppEchoServer81)).
		Setup(cluster)).To(Succeed())

	E2EDeferCleanup(cluster.DismissCluster)
})

func RateLimitOnUniversal() {
	verifyRateLimit := func(client string, total int) func() int {
		return func() int {
			succeeded := 0
			for i := 0; i < total; i++ {
				_, _, err := cluster.Exec("", "", client, "curl", "-v", "--fail", "test-server.mesh")
				if err == nil {
					succeeded++
				}
			}

			return succeeded
		}
	}

	verifyRateLimitExternal := func(client string, total int) func() int {
		return func() int {
			succeeded := 0
			for i := 0; i < total; i++ {
				_, _, err := cluster.Exec("", "", client, "curl", "-v", "--fail", "external-service.mesh")
				if err == nil {
					succeeded++
				}
			}

			return succeeded
		}
	}

	It("should apply limit to client", func() {
		// demo-client specific RateLimit works
		Eventually(verifyRateLimit("demo-client", 5), "60s", "10s").Should(Equal(2))
		// verify determinism by running it once again with shorter timeout
		Eventually(verifyRateLimit("demo-client", 5), "30s", "10s").Should(Equal(2))
	})

	It("should limit per source", func() {
		specificRateLimitPolicy := `
type: RateLimit
mesh: default
name: rate-limit-demo-client
sources:
- match:
    kuma.io/service: "demo-client"
destinations:
- match:
    kuma.io/service: test-server
conf:
  http:
    requests: 4
    interval: 10s
`
		err := YamlUniversal(specificRateLimitPolicy)(cluster)
		Expect(err).ToNot(HaveOccurred())

		// demo-client specific RateLimit works
		Eventually(verifyRateLimit("demo-client", 5), "60s", "10s").Should(Equal(4))
		// verify determinism by running it once again with shorter timeout
		Eventually(verifyRateLimit("demo-client", 5), "30s", "10s").Should(Equal(4))

		// catch-all RateLimit still works
		Eventually(verifyRateLimit("web", 5), "60s", "10s").Should(Equal(2))
		// verify determinism by running it once again with shorter timeout
		Eventually(verifyRateLimit("web", 5), "30s", "10s").Should(Equal(2))
	})

	It("should limit multiple source", func() {
		specificRateLimitPolicy := `
type: RateLimit
mesh: default
name: rate-limit-demo-client
sources:
- match:
    kuma.io/service: "demo-client"
destinations:
- match:
    kuma.io/service: test-server
conf:
  http:
    requests: 4
    interval: 10s
---
type: RateLimit
mesh: default
name: rate-limit-web
sources:
- match:
    kuma.io/service: "web"
destinations:
- match:
    kuma.io/service: test-server
conf:
  http:
    requests: 1
    interval: 10s
`
		err := YamlUniversal(specificRateLimitPolicy)(cluster)
		Expect(err).ToNot(HaveOccurred())

		// demo-client specific RateLimit works
		Eventually(verifyRateLimit("demo-client", 5), "60s", "10s").Should(Equal(4))
		// verify determinism by running it once again with shorter timeout
		Eventually(verifyRateLimit("demo-client", 5), "30s", "10s").Should(Equal(4))

		// web specific RateLimit works
		Eventually(verifyRateLimit("web", 5), "60s", "10s").Should(Equal(1))
		// verify determinism by running it once again with shorter timeout
		Eventually(verifyRateLimit("web", 5), "30s", "10s").Should(Equal(1))
	})

	It("should limit echo server as external service", func() {
		externalService := `
type: ExternalService
mesh: default
name: external-service
tags:
  kuma.io/service: external-service
  kuma.io/protocol: http
networking:
  address: "kuma-3_externalservice-http-server:81"
`
		specificRateLimitPolicy := `
type: RateLimit
mesh: default
name: rate-limit-demo-client
sources:
- match:
    kuma.io/service: "demo-client"
destinations:
- match:
    kuma.io/service: "external-service"
conf:
  http:
    requests: 4
    interval: 10s
`
		err := YamlUniversal(externalService)(cluster)
		Expect(err).ToNot(HaveOccurred())
		err = YamlUniversal(specificRateLimitPolicy)(cluster)
		Expect(err).ToNot(HaveOccurred())

		// demo-client specific RateLimit works
		Eventually(verifyRateLimitExternal("demo-client", 5), "60s", "10s").Should(Equal(4))
		// verify determinism by running it once again with shorter timeout
		Eventually(verifyRateLimitExternal("demo-client", 5), "30s", "10s").Should(Equal(4))
	})
}
