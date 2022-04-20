package deploy

import (
	"strings"

	"github.com/gruntwork-io/terratest/modules/retry"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

func UniversalTransparentProxyDeployment() {
	var cluster Cluster

	BeforeEach(func() {
		cluster = NewUniversalCluster(NewTestingT(), Kuma3, Silent)

		err := NewClusterSetup().
			Install(Kuma(core.Standalone)).
			Install(TestServerUniversal("test-server", "default",
				WithArgs([]string{"echo", "--instance", "universal"}),
				WithServiceName("echo-server_kuma-test_svc_8080")),
			).
			Install(DemoClientUniversal(AppModeDemoClient, "default", WithTransparentProxy(true))).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterEach(func() {
		Expect(cluster.DeleteKuma()).To(Succeed())
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	It("should access the service using .mesh", func() {
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
		retry.DoWithRetry(cluster.GetTesting(), "curl service with dots",
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
	})

	It("should re-install transparent proxy", func() {
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

		stdout, _, err := cluster.ExecWithRetries("", "", "demo-client",
			"/usr/bin/kumactl", "install", "transparent-proxy",
			"--kuma-dp-user", "kuma-dp", "--skip-resolv-conf", "--verbose")
		Expect(stdout).To(ContainSubstring("Transparent proxy set up successfully"))
		Expect(err).ToNot(HaveOccurred())

		retry.DoWithRetry(cluster.GetTesting(), "curl service with dots",
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
	})
}
