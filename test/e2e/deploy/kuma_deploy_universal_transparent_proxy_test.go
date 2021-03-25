package deploy_test

import (
	"strings"

	"github.com/gruntwork-io/terratest/modules/retry"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

var _ = Describe("Test Universal Transparent Proxy deployment", func() {

	const iterations = 10

	var cluster Cluster
	var deployOptsFuncs []DeployOptionsFunc

	BeforeEach(func() {
		cluster = NewUniversalCluster(NewTestingT(), Kuma3, Silent)
		deployOptsFuncs = []DeployOptionsFunc{}

		err := NewClusterSetup().
			Install(Kuma(core.Standalone, deployOptsFuncs...)).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())
		err = cluster.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		echoServerToken, err := cluster.GetKuma().GenerateDpToken("default", "echo-server_kuma-test_svc_8080")
		Expect(err).ToNot(HaveOccurred())
		demoClientToken, err := cluster.GetKuma().GenerateDpToken("default", "demo-client")
		Expect(err).ToNot(HaveOccurred())

		err = EchoServerUniversal(AppModeEchoServer, "default", "universal", echoServerToken, WithTransparentProxy(false))(cluster)
		Expect(err).ToNot(HaveOccurred())
		err = DemoClientUniversal(AppModeDemoClient, "default", demoClientToken, WithTransparentProxy(true))(cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}
		Expect(cluster.DeleteKuma(deployOptsFuncs...)).To(Succeed())
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	It("should access the service using .mesh", func() {

		for i := 0; i < iterations; i++ {
			retry.DoWithRetry(cluster.GetTesting(), "curl remote service",
				DefaultRetries, DefaultTimeout,
				func() (string, error) {
					stdout, _, err := cluster.ExecWithRetries("", "", "demo-client",
						"curl", "-v", "-m", "3", "echo-server_kuma-test_svc_8080.mesh")
					if err != nil {
						return "should retry", err
					}
					if strings.Contains(stdout, "HTTP/1.1 200 OK") {
						return "Accessing service successful", nil
					}
					return "should retry", errors.Errorf("should retry")
				})
			retry.DoWithRetry(cluster.GetTesting(), "curl remote service with dots",
				DefaultRetries, DefaultTimeout,
				func() (string, error) {
					stdout, _, err := cluster.ExecWithRetries("", "", "demo-client",
						"curl", "-v", "-m", "3", "echo-server.kuma-test.svc.8080.mesh")
					if err != nil {
						return "should retry", err
					}
					if strings.Contains(stdout, "HTTP/1.1 200 OK") {
						return "Accessing service successful", nil
					}
					return "should retry", errors.Errorf("should retry")
				})
		}
	})
})
