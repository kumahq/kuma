package ratelimit

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/externalservice"
)

const kumaClusterId = Kuma3

const limitInterval = 10 * time.Second

var cluster Cluster
var rateLimitPolicy = fmt.Sprintf(`
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
    interval: %s
    onRateLimit:
      status: 429
      headers:
        - key: "x-kuma-rate-limited"
          value: "true"
`, limitInterval)

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
	// In the best case, we start sending requests and can send all successfully
	// and consecutively before the bucket is empty.
	// Otherwise the bucket empties before we've sent all which will lead to the
	// number of successes being more than the limit.
	// So we send requests until the rate limit is first hit, send failing
	// requests until the first one succeeds, then we send as
	// many requests as we can until we are rate limited again.
	// We compare that number of successes to the expected number.
	// The assumption here is that we can send enough requests to hit the limit
	// within the interval.
	verifyRateLimit := func(client string, svc string, expected int) {
		// This is wrapped in Eventually. Shortly after new
		// RateLimits are applied, there appears to be some nondeterminism around how
		// many tokens are put in the bucket at one time.
		verify := func() error {
			verifyErrorMsg := fmt.Sprintf("couldn't verify rate limit for %s to %s to equal %d", client, svc, expected)

			begin := time.Now()
			succeeded := 0
			var haveFailedRequest bool
			for {
				_, _, err := cluster.Exec("", "", client, "curl", "-v", "--fail", fmt.Sprintf("%s.mesh", svc))
				if err != nil {
					haveFailedRequest = true
					if succeeded > 0 {
						// we've failed, then succeeded and are now failing again
						if succeeded == expected {
							return nil
						}
						return fmt.Errorf("%s: sent %d successful requests", verifyErrorMsg, succeeded)
					}
				} else if haveFailedRequest {
					// We've failed at least once and have n >= 0 consecutive,
					// successful requests before this one
					succeeded++
				}
				if time.Now().After(begin.Add(2 * limitInterval)) {
					return fmt.Errorf("%s: couldn't send a terminated string of successful requests", verifyErrorMsg)
				}
			}
		}

		const verifyTimeout = 30 * time.Second
		verifyPollingInterval := limitInterval + 1*time.Second

		Eventually(verify, verifyTimeout, verifyPollingInterval).Should(Succeed())
		Consistently(verify, verifyTimeout, 1*time.Second).Should(Succeed())
	}

	It("should apply limit to client", func() {
		// demo-client specific RateLimit works
		verifyRateLimit("demo-client", "test-server", 2)
	})

	It("should limit per source", func() {
		specificRateLimitPolicy := fmt.Sprintf(`
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
    interval: %s
`, limitInterval)
		err := YamlUniversal(specificRateLimitPolicy)(cluster)
		Expect(err).ToNot(HaveOccurred())

		// demo-client specific RateLimit works
		verifyRateLimit("demo-client", "test-server", 4)

		// catch-all RateLimit still works
		verifyRateLimit("web", "test-server", 2)
	})

	It("should limit multiple source", func() {
		specificRateLimitPolicy := fmt.Sprintf(`
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
    interval: %[1]s
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
    interval: %[1]s
`, limitInterval)
		err := YamlUniversal(specificRateLimitPolicy)(cluster)
		Expect(err).ToNot(HaveOccurred())

		// demo-client specific RateLimit works
		verifyRateLimit("demo-client", "test-server", 4)

		// demo-client specific RateLimit works
		verifyRateLimit("web", "test-server", 1)
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
		specificRateLimitPolicy := fmt.Sprintf(`
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
    interval: %s
`, limitInterval)
		err := YamlUniversal(externalService)(cluster)
		Expect(err).ToNot(HaveOccurred())
		err = YamlUniversal(specificRateLimitPolicy)(cluster)
		Expect(err).ToNot(HaveOccurred())

		// demo-client specific RateLimit works
		verifyRateLimit("demo-client", "external-service", 4)
	})
}
