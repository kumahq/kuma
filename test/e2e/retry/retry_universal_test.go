package retry_test

import (
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/retry"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config/core"

	. "github.com/kumahq/kuma/test/framework"
)

var _ = Describe("Test Retry on Universal", func() {
	var cluster Cluster
	var deployOptsFuncs []DeployOptionsFunc

	BeforeEach(func() {
		clusters, err := NewUniversalClusters(
			[]string{Kuma1},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		// Global
		cluster = clusters.GetCluster(Kuma1)
		deployOptsFuncs = []DeployOptionsFunc{}

		err = NewClusterSetup().
			Install(Kuma(core.Standalone, deployOptsFuncs...)).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())

		err = cluster.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		demoClientToken, err := cluster.GetKuma().GenerateDpToken("default", "demo-client")
		Expect(err).ToNot(HaveOccurred())

		echoServerToken, err := cluster.GetKuma().GenerateDpToken("default", "echo-server_kuma-test_svc_8080")
		Expect(err).ToNot(HaveOccurred())

		err = cluster.GetKumactlOptions().RunKumactl("delete", "retry", "retry-all-default")
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(DemoClientUniversal(AppModeDemoClient, "default", demoClientToken)).
			Install(EchoServerUniversal(AppModeEchoServer, "default", "universal", echoServerToken)).
			Setup(cluster)
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

	It("should retry on TCP connection failure", func() {
		echoServerDataplane := `
type: Dataplane
mesh: default
name: fake-echo-server
networking:
  address:  241.0.0.1
  inbound:
  - port: 7777
    servicePort: 7777
    tags:
      kuma.io/service: echo-server_kuma-test_svc_8080
      kuma.io/protocol: http
`
		retryPolicy := `
type: Retry
mesh: default
name: fake-retry-policy
sources:
- match:
    kuma.io/service: demo-client
destinations:
- match:
    kuma.io/service: echo-server_kuma-test_svc_8080
conf:
  http:
    numRetries: 5
`

		retry.DoWithRetry(cluster.GetTesting(), "curl local service",
			DefaultRetries, DefaultTimeout,
			func() (string, error) {
				stdout, _, err := cluster.ExecWithRetries("", "", "demo-client",
					"curl", "-v", "-m", "3", "--fail", "localhost:4001")
				if err != nil {
					return "should retry", err
				}
				if strings.Contains(stdout, "HTTP/1.1 200 OK") {
					return "Accessing service successful", nil
				}
				return "should retry", errors.Errorf("should retry")
			})

		for i := 0; i < 10; i++ {
			// -m 8 to wait for 8 seconds to beat the default 5s connect timeout
			stdout, stderr, err := cluster.Exec("", "", "demo-client", "curl", "-v", "-m", "8", "--fail", "localhost:4001")
			Expect(err).ToNot(HaveOccurred())
			Expect(stderr).To(BeEmpty())
			Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
		}

		err := YamlUniversal(echoServerDataplane)(cluster)
		Expect(err).ToNot(HaveOccurred())

		time.Sleep(5 * time.Second)

		var errs []error

		for i := 0; i < 10; i++ {
			_, _, err := cluster.Exec("", "", "demo-client", "curl", "-v", "-m", "8", "--fail", "localhost:4001")

			if err != nil {
				errs = append(errs, err)
			}
		}

		Expect(errs).ToNot(BeEmpty())

		err = YamlUniversal(retryPolicy)(cluster)
		Expect(err).ToNot(HaveOccurred())

		time.Sleep(5 * time.Second)

		for i := 0; i < 10; i++ {
			stdout, stderr, err := cluster.Exec("", "", "demo-client", "curl", "-v", "-m", "8", "--fail", "localhost:4001")
			Expect(err).ToNot(HaveOccurred())
			Expect(stderr).To(BeEmpty())
			Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
		}
	})
})
