package universal

import (
	"encoding/base64"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

func PolicyTCP() {
	healthCheck := func(send, recv string) string {
		sendBase64 := base64.StdEncoding.EncodeToString([]byte(send))
		recvBase64 := base64.StdEncoding.EncodeToString([]byte(recv))

		return fmt.Sprintf(`
type: HealthCheck
name: gateway-to-backend
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
  tcp: 
    send: %s
    receive:
    - %s`, sendBase64, recvBase64)
	}

	var cluster Cluster

	BeforeEach(func() {
		cluster = NewUniversalCluster(NewTestingT(), Kuma3, Verbose)

		err := NewClusterSetup().
			Install(Kuma(config_core.Standalone)).
			Install(YamlUniversal(healthCheck("foo", "bar"))).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())

		demoClientToken, err := cluster.GetKuma().GenerateDpToken("default", "dp-demo-client")
		Expect(err).ToNot(HaveOccurred())
		testServerToken, err := cluster.GetKuma().GenerateDpToken("default", "test-server")
		Expect(err).ToNot(HaveOccurred())

		err = DemoClientUniversal("dp-demo-client", "default", demoClientToken,
			WithTransparentProxy(true))(cluster)
		Expect(err).ToNot(HaveOccurred())
		err = TestServerUniversal("test-server", "default", testServerToken,
			WithArgs([]string{"health-check", "tcp"}),
			WithProtocol("tcp"))(cluster)
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
		cmd := []string{"/bin/bash", "-c", "\"echo request | nc test-server.mesh 80\""}
		stdout, _, err := cluster.ExecWithRetries("", "", "dp-demo-client", cmd...)
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("response"))

		// update HealthCheck policy to check for another 'recv' line
		err = YamlUniversal(healthCheck("foo", "baz"))(cluster)
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
		cmd = []string{"/bin/bash", "-c", "\"echo request | nc test-server.mesh 80\""}
		stdout, _, _ = cluster.ExecWithRetries("", "", "dp-demo-client", cmd...)

		// there is no real attempt to setup a connection with test-server, but Envoy may return either
		// empty response with EXIT_CODE = 0, or  'Ncat: Connection reset by peer.' with EXIT_CODE = 1
		Expect(stdout).To(Or(BeEmpty(), ContainSubstring("Ncat: Connection reset by peer.")))
	})
}
