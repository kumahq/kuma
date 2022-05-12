package healthcheck

import (
	"encoding/base64"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/universal/env"
	. "github.com/kumahq/kuma/test/framework"
)

func Policy() {
	Describe("HTTP", func() {
		meshName := "healthcheck-http"
		healthCheck := func(mesh, method, status string) string {
			return fmt.Sprintf(`
type: HealthCheck
name: everything-to-backend
mesh: %s
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
    - %s`, mesh, method, status)
		}
		It("should mark host as unhealthy if it doesn't reply on health checks", func() {
			DeferCleanup(func() {
				Expect(env.Cluster.DeleteMeshApps(meshName)).To(Succeed())
				Expect(env.Cluster.DeleteMesh(meshName)).To(Succeed())
			})
			err := NewClusterSetup().
				Install(MeshUniversal(meshName)).
				Install(YamlUniversal(healthCheck(meshName, "health", "200"))).
				Install(DemoClientUniversal("dp-demo-client", meshName,
					WithTransparentProxy(true)),
				).
				Install(TestServerUniversal("test-server", meshName, WithArgs([]string{"health-check", "http"}))).
				Setup(env.Cluster)
			Expect(err).ToNot(HaveOccurred())

			// check that test-server is healthy
			cmd := []string{"curl", "--fail", "test-server.mesh/content"}
			stdout, _, err := env.Cluster.ExecWithRetries("", "", "dp-demo-client", cmd...)
			Expect(err).ToNot(HaveOccurred())
			Expect(stdout).To(ContainSubstring("response"))

			// update HealthCheck policy to check for another status code
			Expect(YamlUniversal(healthCheck(meshName, "are-you-healthy", "500"))(env.Cluster)).To(Succeed())

			// wait cluster 'test-server' to be marked as unhealthy
			Eventually(func() bool {
				cmd := []string{"/bin/bash", "-c", "\"curl localhost:9901/clusters | grep test-server\""}
				stdout, _, err := env.Cluster.ExecWithRetries("", "", "dp-demo-client", cmd...)
				if err != nil {
					return false
				}
				return strings.Contains(stdout, "health_flags::/failed_active_hc")
			}, "30s", "500ms").Should(BeTrue())

			// check that test-server is unhealthy
			cmd = []string{"curl", "test-server.mesh/content"}
			stdout, _, err = env.Cluster.ExecWithRetries("", "", "dp-demo-client", cmd...)
			Expect(err).ToNot(HaveOccurred())
			Expect(stdout).To(ContainSubstring("no healthy upstream"))
		})
	})

	Describe("TCP", func() {
		healthCheck := func(mesh, send, recv string) string {
			sendBase64 := base64.StdEncoding.EncodeToString([]byte(send))
			recvBase64 := base64.StdEncoding.EncodeToString([]byte(recv))

			return fmt.Sprintf(`
type: HealthCheck
name: gateway-to-backend
mesh: %s
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
    - %s`, mesh, sendBase64, recvBase64)
		}

		It("should mark host as unhealthy if it doesn't reply on health checks", func() {
			meshName := "healthcheck-tcp"
			DeferCleanup(func() {
				Expect(env.Cluster.DeleteMeshApps(meshName)).To(Succeed())
				Expect(env.Cluster.DeleteMesh(meshName)).To(Succeed())
			})
			err := NewClusterSetup().
				Install(MeshUniversal(meshName)).
				Install(YamlUniversal(healthCheck(meshName, "foo", "bar"))).
				Install(DemoClientUniversal("dp-demo-client", meshName,
					WithTransparentProxy(true)),
				).
				Install(TestServerUniversal("test-server", meshName,
					WithArgs([]string{"health-check", "tcp"}),
					WithProtocol("tcp")),
				).
				Setup(env.Cluster)
			Expect(err).ToNot(HaveOccurred())

			// check that test-server is healthy
			cmd := []string{"/bin/bash", "-c", "\"echo request | nc test-server.mesh 80\""}
			stdout, _, err := env.Cluster.ExecWithRetries("", "", "dp-demo-client", cmd...)
			Expect(err).ToNot(HaveOccurred())
			Expect(stdout).To(ContainSubstring("response"))

			// update HealthCheck policy to check for another 'recv' line
			Expect(YamlUniversal(healthCheck(meshName, "foo", "baz"))(env.Cluster)).To(Succeed())

			// wait cluster 'test-server' to be marked as unhealthy
			Eventually(func() bool {
				cmd := []string{"/bin/bash", "-c", "\"curl localhost:9901/clusters | grep test-server\""}
				stdout, _, err := env.Cluster.ExecWithRetries("", "", "dp-demo-client", cmd...)
				if err != nil {
					return false
				}
				return strings.Contains(stdout, "health_flags::/failed_active_hc")
			}, "30s", "500ms").Should(BeTrue())

			cmd = []string{"/bin/bash", "-c", "\"echo request | nc test-server.mesh 80\""}
			stdout, _, _ = env.Cluster.ExecWithRetries("", "", "dp-demo-client", cmd...)

			// there is no real attempt to setup a connection with test-server, but Envoy may return either
			// empty response with EXIT_CODE = 0, or  'Ncat: Connection reset by peer.' with EXIT_CODE = 1
			Expect(stdout).To(Or(BeEmpty(), ContainSubstring("Ncat: Connection reset by peer.")))
		})
	})
}
