package universal

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

func PolicyHTTP() {
	healthCheck := func(method, status string) string {
		return fmt.Sprintf(`
type: HealthCheck
name: everything-to-backend
mesh: default
sources:
- match:
    kuma.io/service: '*'
destinations:
- match:
    kuma.io/service: test-server
conf:
  interval: 10s
  timeout: 2s
  unhealthyThreshold: 3
  healthyThreshold: 1
  failTrafficOnPanic: true
  noTrafficInterval: 1s
  healthyPanicThreshold: 0
  reuse_connection: true
  http: 
    path: /%s
    expectedStatuses: 
    - %s`, method, status)
	}

	var cluster Cluster

	BeforeEach(func() {
		cluster = NewUniversalCluster(NewTestingT(), Kuma3, Verbose)

		err := NewClusterSetup().
			Install(Kuma(config_core.Standalone)).
			Install(YamlUniversal(healthCheck("health", "200"))).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())

		demoClientToken, err := cluster.GetKuma().GenerateDpToken("default", "dp-demo-client")
		Expect(err).ToNot(HaveOccurred())
		testServerToken, err := cluster.GetKuma().GenerateDpToken("default", "test-server")
		Expect(err).ToNot(HaveOccurred())

		err = DemoClientUniversal("dp-demo-client", "default", demoClientToken,
			WithTransparentProxy(true))(cluster)
		Expect(err).ToNot(HaveOccurred())
		err = TestServerUniversal("test-server", "default", testServerToken, WithArgs([]string{"health-check", "http"}))(cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}
		Expect(cluster.DeleteKuma()).To(Succeed())
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	It("should mark host as unhealthy if it doesn't reply on health checks", func() {
		// check that test-server is healthy
		cmd := []string{"curl", "--fail", "test-server.mesh/content"}
		stdout, _, err := cluster.ExecWithRetries("", "", "dp-demo-client", cmd...)
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("response"))

		// update HealthCheck policy to check for another status code
		err = YamlUniversal(healthCheck("are-you-healthy", "500"))(cluster)
		Expect(err).ToNot(HaveOccurred())

		// wait cluster 'test-server' to be marked as unhealthy
		Eventually(func() bool {
			cmd := []string{"/bin/bash", "-c", "\"curl localhost:9901/clusters | grep test-server\""}
			stdout, _, err := cluster.ExecWithRetries("", "", "dp-demo-client", cmd...)
			if err != nil {
				return false
			}
			return strings.Contains(stdout, "health_flags::/failed_active_hc")
		}, "30s", "500ms").Should(BeTrue())

		// check that test-server is unhealthy
		cmd = []string{"curl", "test-server.mesh/content"}
		stdout, _, err = cluster.ExecWithRetries("", "", "dp-demo-client", cmd...)
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("no healthy upstream"))
	})
}
