package ratelimit

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/externalservice"
)

const kumaClusterId = Kuma3

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
    requests: 1
    interval: 10s
    onRateLimit:
      status: 429
      headers:
        - key: "x-kuma-rate-limited"
          value: "true"`

var _ = E2EBeforeSuite(func() {
	clusters, err := NewUniversalClusters(
		[]string{kumaClusterId},
		Silent)
	Expect(err).ToNot(HaveOccurred())

	// Global
	cluster = clusters.GetCluster(kumaClusterId)

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
	requestRateLimited := func(client string, svc string, status string) bool {
		stdout, _, err := cluster.Exec("", "", client, "curl", "-v", fmt.Sprintf("%s.mesh", svc))
		return err == nil && strings.Contains(stdout, status)
	}

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
    onRateLimit:
      status: 400
    requests: 1
    interval: 10s
`
		err := YamlUniversal(specificRateLimitPolicy)(cluster)
		Expect(err).ToNot(HaveOccurred())

		// demo-client specific RateLimit works
		Eventually(func() bool {
			return requestRateLimited("demo-client", "test-server", "400")
		}, "30s", "100ms").Should(BeTrue())

		// catch-all RateLimit still works
		Eventually(func() bool {
			return requestRateLimited("web", "test-server", "429")
		}, "30s", "100ms").Should(BeTrue())
	})

	It("should limit echo server as external service", func() {
		externalService := fmt.Sprintf(`
type: ExternalService
mesh: default
name: external-service
tags:
  kuma.io/service: external-service
  kuma.io/protocol: http
networking:
  address: "%s_externalservice-http-server:81"
`, kumaClusterId)
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
    onRateLimit:
      status: 429
    requests: 1
    interval: 10s
`
		err := YamlUniversal(externalService)(cluster)
		Expect(err).ToNot(HaveOccurred())
		err = YamlUniversal(specificRateLimitPolicy)(cluster)
		Expect(err).ToNot(HaveOccurred())

		// demo-client specific RateLimit works
		Eventually(func() bool {
			return requestRateLimited("demo-client", "external-service", "429")
		}, "30s", "100ms").Should(BeTrue())
	})
}
