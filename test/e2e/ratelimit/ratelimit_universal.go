package ratelimit

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"

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
    interval: 5s
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

		err = YamlUniversal(rateLimitPolicy)(cluster)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(EchoServerUniversal(AppModeEchoServer, "default", "universal", echoServerToken, WithTransparentProxy(true))).
			Install(DemoClientUniversal(AppModeDemoClient, "default", demoClientToken, WithTransparentProxy(true))).
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

	verifyRateLimit := func(client string, succeed, fail int) func() bool {
		return func() bool {
			for i := 0; i < succeed; i++ {
				_, _, err := cluster.Exec("", "", client, "curl", "-v", "--fail", "echo-server_kuma-test_svc_8080.mesh")

				if err != nil {
					return false
				}
			}

			for i := 0; i < fail; i++ {
				_, _, err := cluster.Exec("", "", client, "curl", "-v", "--fail", "echo-server_kuma-test_svc_8080.mesh")
				if err == nil {
					return false
				}
			}
			return true
		}
	}

	It("should limit to 2 requests per 5 sec", func() {
		Eventually(verifyRateLimit("demo-client", 2, 3), "60s", "5s").Should(BeTrue())
		time.Sleep(6 * time.Second)
		Expect(verifyRateLimit("demo-client", 2, 3)()).To(BeTrue())
	})

	It("should limit to 4 requests per 5 sec", func() {
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
    interval: 5s
`
		err := YamlUniversal(specificRateLimitPolicy)(cluster)
		Expect(err).ToNot(HaveOccurred())

		Eventually(verifyRateLimit("demo-client", 4, 1), "60s", "5s").Should(BeTrue())
		time.Sleep(6 * time.Second)
		Expect(verifyRateLimit("demo-client", 4, 1)()).To(BeTrue())
	})
}
