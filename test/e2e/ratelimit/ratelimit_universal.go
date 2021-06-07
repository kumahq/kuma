package ratelimit

import (
	"fmt"

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
   connections: 5
   interval: 10s
   onRateLimit:
     status: 423
     headers:
       - key: "x-kuma-rate-limited"
         value: "true"
`

	BeforeEach(func() {
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

	AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}
		err := cluster.DeleteKuma(deployOptsFuncs...)
		Expect(err).ToNot(HaveOccurred())

		err = cluster.DismissCluster()
		Expect(err).ToNot(HaveOccurred())
	})

	It("should limit to 5 requests per 10 sec", func() {
		Eventually(func() bool {
			_, _, err := cluster.Exec("", "", "demo-client", "curl", "-v", "--fail", "echo-server_kuma-test_svc_8080.mesh")
			return err == nil
		}, "30s", "3s").Should(BeTrue())

		for i := 0; i < 4; i++ {
			stdout, stderr, err := cluster.Exec("", "", "demo-client", "curl", "-v", "--fail", "echo-server_kuma-test_svc_8080.mesh")
			Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("failed at %d request, \n-------\n %s  \n-------\n %s", i, stdout, stderr))
			Expect(stderr).To(BeEmpty(), fmt.Sprintf("failed at %d request", i))
			Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"), fmt.Sprintf("failed at %d request", i))
		}

		for i := 0; i < 5; i++ {
			stdout, stderr, err := cluster.Exec("", "", "demo-client", "curl", "-v", "--fail", "echo-server_kuma-test_svc_8080.mesh")
			Expect(err).To(HaveOccurred(), fmt.Sprintf("failed at %d request, \n-------\n %s  \n-------\n %s", i, stdout, stderr))
			Expect(stderr).To(BeEmpty(), fmt.Sprintf("failed at %d request", i))
			Expect(stdout).To(ContainSubstring("423 Locked"), fmt.Sprintf("failed at %d request", i))
		}
	})

	It("should limit to 8 requests per 10 sec", func() {
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
    connections: 8
    interval: 10s
`

		err := YamlUniversal(specificRateLimitPolicy)(cluster)
		Expect(err).ToNot(HaveOccurred())

		Eventually(func() bool {
			_, _, err := cluster.Exec("", "", "demo-client", "curl", "-v", "--fail", "echo-server_kuma-test_svc_8080.mesh")
			return err == nil
		}, "30s", "3s").Should(BeTrue())

		for i := 0; i < 7; i++ {
			stdout, stderr, err := cluster.Exec("", "", "demo-client", "curl", "-v", "--fail", "echo-server_kuma-test_svc_8080.mesh")
			Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("failed at %d request, \n-------\n %s  \n-------\n %s", i, stdout, stderr))
			Expect(stderr).To(BeEmpty(), fmt.Sprintf("failed at %d request", i))
			Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"), fmt.Sprintf("failed at %d request", i))
		}

		for i := 0; i < 2; i++ {
			stdout, stderr, err := cluster.Exec("", "", "demo-client", "curl", "-v", "--fail", "echo-server_kuma-test_svc_8080.mesh")
			Expect(err).To(HaveOccurred(), fmt.Sprintf("failed at %d request, \n-------\n %s  \n-------\n %s", i, stdout, stderr))
			Expect(stderr).To(BeEmpty(), fmt.Sprintf("failed at %d request", i))
			Expect(stdout).To(ContainSubstring("429 Too Many Requests"), fmt.Sprintf("failed at %d request", i))
		}
	})
}
