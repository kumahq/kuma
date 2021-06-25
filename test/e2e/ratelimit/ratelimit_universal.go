package ratelimit

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

func RateLimitOnUniversal() {
	var cluster Cluster
	var deployOptsFuncs []DeployOptionsFunc
	rateLimitPolicy := `
type: RateLimit
mesh: default
name: rate-limit-all-sources
sources:
- match:
   kuma.io/service: "*"
destinations:
- match:
   kuma.io/service: echo-server_kuma-test_svc_8080
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

	E2EBeforeSuite(func() {
		clusters, err := NewUniversalClusters(
			[]string{Kuma3},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		// Global
		cluster = clusters.GetCluster(Kuma3)
		deployOptsFuncs = KumaUniversalDeployOpts

		err = NewClusterSetup().
			Install(Kuma(core.Standalone, deployOptsFuncs...)).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())

		demoClientToken, err := cluster.GetKuma().GenerateDpToken("default", "demo-client")
		Expect(err).ToNot(HaveOccurred())

		echoServerToken, err := cluster.GetKuma().GenerateDpToken("default", "echo-server_kuma-test_svc_8080")
		Expect(err).ToNot(HaveOccurred())

		webToken, err := cluster.GetKuma().GenerateDpToken("default", "web")
		Expect(err).ToNot(HaveOccurred())

		err = YamlUniversal(rateLimitPolicy)(cluster)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(EchoServerUniversal(AppModeEchoServer, "default", "universal", echoServerToken, WithTransparentProxy(true))).
			Install(DemoClientUniversal(AppModeDemoClient, "default", demoClientToken, WithTransparentProxy(true))).
			Install(DemoClientUniversal("web", "default", webToken, WithTransparentProxy(true))).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())

		err = cluster.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterSuite(func() {
		if ShouldSkipCleanup() {
			return
		}
		err := cluster.DeleteKuma(deployOptsFuncs...)
		Expect(err).ToNot(HaveOccurred())

		err = cluster.DismissCluster()
		Expect(err).ToNot(HaveOccurred())
	})

	verifyRateLimit := func(client string, total int) func() int {
		return func() int {
			succeeded := 0
			for i := 0; i < total; i++ {
				_, _, err := cluster.Exec("", "", client, "curl", "-v", "--fail", "echo-server_kuma-test_svc_8080.mesh")
				if err == nil {
					succeeded++
				}
			}

			return succeeded
		}
	}

	It("should limit to 2 requests per 5 sec", func() {
		// demo-client specific RateLimit works
		Eventually(verifyRateLimit("demo-client", 5), "60s", "1s").Should(Equal(2))
		// verify determinism by running it once again with shorter timeout
		Eventually(verifyRateLimit("demo-client", 5), "30s", "1s").Should(Equal(2))
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
    kuma.io/service: echo-server_kuma-test_svc_8080
conf:
  http:
    requests: 4
    interval: 10s
`
		err := YamlUniversal(specificRateLimitPolicy)(cluster)
		Expect(err).ToNot(HaveOccurred())

		// demo-client specific RateLimit works
		Eventually(verifyRateLimit("demo-client", 5), "60s", "1s").Should(Equal(4))
		// verify determinism by running it once again with shorter timeout
		Eventually(verifyRateLimit("demo-client", 5), "30s", "1s").Should(Equal(4))

		// catch-all RateLimit still works
		Eventually(verifyRateLimit("web", 5), "60s", "1s").Should(Equal(2))
		// verify determinism by running it once again with shorter timeout
		Eventually(verifyRateLimit("web", 5), "30s", "1s").Should(Equal(2))
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
    kuma.io/service: echo-server_kuma-test_svc_8080
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
    kuma.io/service: echo-server_kuma-test_svc_8080
conf:
  http:
    requests: 1
    interval: 10s
`
		err := YamlUniversal(specificRateLimitPolicy)(cluster)
		Expect(err).ToNot(HaveOccurred())

		// demo-client specific RateLimit works
		Eventually(verifyRateLimit("demo-client", 5), "60s", "1s").Should(Equal(4))
		// verify determinism by running it once again with shorter timeout
		Eventually(verifyRateLimit("demo-client", 5), "30s", "1s").Should(Equal(4))

		// web specific RateLimit works
		Eventually(verifyRateLimit("web", 5), "60s", "1s").Should(Equal(1))
		// verify determinism by running it once again with shorter timeout
		Eventually(verifyRateLimit("web", 5), "30s", "1s").Should(Equal(1))
	})
}
