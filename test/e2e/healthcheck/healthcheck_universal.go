package healthcheck

import (
	"encoding/base64"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

func ServiceProbes() {
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

		err = EchoServerUniversal("dp-echo-server", "default", "universal", echoServerToken, ProxyOnly(), ServiceProbe())(cluster)
		Expect(err).ToNot(HaveOccurred())
		err = DemoClientUniversal("dp-demo-client", "default", demoClientToken, ServiceProbe())(cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}
		Expect(cluster.DeleteKuma(deployOptsFuncs...)).To(Succeed())
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	It("should update dataplane.inbound.health", func() {
		Eventually(func() (string, error) {
			output, err := cluster.GetKumactlOptions().RunKumactlAndGetOutputV(Verbose, "get", "dataplane", "dp-echo-server", "-oyaml")
			if err != nil {
				return "", err
			}
			return output, nil
		}, "30s", "500ms").Should(ContainSubstring("health: {}"))

		Eventually(func() (string, error) {
			output, err := cluster.GetKumactlOptions().RunKumactlAndGetOutputV(Verbose, "get", "dataplane", "dp-demo-client", "-oyaml")
			if err != nil {
				return "", err
			}
			return output, nil
		}, "30s", "500ms").Should(ContainSubstring("ready: true"))
	})
}

func Policy() {
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
	var deployOptsFuncs []DeployOptionsFunc

	BeforeEach(func() {
		cluster = NewUniversalCluster(NewTestingT(), Kuma3, Verbose)
		deployOptsFuncs = []DeployOptionsFunc{}

		err := NewClusterSetup().
			Install(Kuma(core.Standalone, deployOptsFuncs...)).
			Install(YamlUniversal(healthCheck("foo", "bar"))).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())
		err = cluster.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		demoClientToken, err := cluster.GetKuma().GenerateDpToken("default", "demo-client")
		Expect(err).ToNot(HaveOccurred())
		testServerToken, err := cluster.GetKuma().GenerateDpToken("default", "test-server")
		Expect(err).ToNot(HaveOccurred())

		err = DemoClientUniversal("dp-demo-client", "default", demoClientToken,
			WithTransparentProxy(true), WithBuiltinDNS(true))(cluster)
		Expect(err).ToNot(HaveOccurred())
		err = TestServerUniversal("test-server", "default", testServerToken, WithTransparentProxy(true))(cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}
		Expect(cluster.DeleteKuma(deployOptsFuncs...)).To(Succeed())
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	FIt("should mark host as unhealthy if it doesn't reply on health checks", func() {
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
			cmd := []string{"/bin/bash", "-c", "\"curl localhost:30001/clusters | grep test-server\""}
			stdout, _, err := cluster.ExecWithRetries("", "", "dp-demo-client", cmd...)
			if err != nil {
				return false
			}
			return strings.Contains(stdout, "health_flags::/failed_active_hc")
		}, "30s", "500ms").Should(BeTrue())
	})
}
