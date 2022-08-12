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
		BeforeAll(func() {
			err := NewClusterSetup().
				Install(MeshUniversal(meshName)).
				Install(YamlUniversal(healthCheck(meshName, "health", "200"))).
				Install(DemoClientUniversal("dp-demo-client", meshName,
					WithTransparentProxy(true)),
				).
				Install(TestServerUniversal("test-server", meshName, WithArgs([]string{"health-check", "http"}))).
				Setup(env.Cluster)
			Expect(err).ToNot(HaveOccurred())
		})

		E2EAfterAll(func() {
			Expect(env.Cluster.DeleteMeshApps(meshName)).To(Succeed())
			Expect(env.Cluster.DeleteMesh(meshName)).To(Succeed())
		})

		It("should mark host as unhealthy if it doesn't reply on health checks", func() {
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
	}, Ordered)

	Describe("TCP", func() {
		healthCheck := func(mesh, serviceName, send, recv string) string {
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
    kuma.io/service: %s
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
    - %s`, mesh, serviceName, sendBase64, recvBase64)
		}
		meshName := "healthcheck-tcp"
		BeforeAll(func() {
			err := NewClusterSetup().
				Install(MeshUniversal(meshName)).
				Install(DemoClientUniversal("dp-demo-client", meshName,
					WithTransparentProxy(true)),
				).
				Install(TestServerUniversal("test-server", meshName,
					WithArgs([]string{"health-check", "tcp"}),
					WithProtocol("tcp")),
				).
				Setup(env.Cluster)
			Expect(err).ToNot(HaveOccurred())
		})
		E2EAfterAll(func() {
			Expect(env.Cluster.DeleteMeshApps(meshName)).To(Succeed())
			Expect(env.Cluster.DeleteMesh(meshName)).To(Succeed())
		})

		It("should mark host as unhealthy if it doesn't reply on health checks", func() {
			// check that test-server is healthy
			cmd := []string{"/bin/bash", "-c", "\"echo request | nc test-server.mesh 80\""}
			stdout, _, err := env.Cluster.ExecWithRetries("", "", "dp-demo-client", cmd...)
			Expect(err).ToNot(HaveOccurred())
			Expect(stdout).To(ContainSubstring("response"))

			// update HealthCheck policy to check for another 'recv' line
			Expect(YamlUniversal(healthCheck(meshName, "test-server", "foo", "baz"))(env.Cluster)).To(Succeed())

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
	}, Ordered)

	Context("TCP with permissive mTLS", func() {
		mtlsPermissiveMesh := func(mesh string) InstallFunc {
			return YamlUniversal(fmt.Sprintf(`
type: Mesh
name: %s
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
    mode: PERMISSIVE
`, mesh))
		}
		healthCheck := func(mesh, serviceName, send, recv string) string {
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
    kuma.io/service: %s
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
    - %s`, mesh, serviceName, sendBase64, recvBase64)
		}
		meshName := "healthcheck-mtls-permissive-tcp"
		BeforeAll(func() {
			err := NewClusterSetup().
				Install(mtlsPermissiveMesh(meshName)).
				Install(DemoClientUniversal("dp-demo-client-mtls", meshName,
					WithTransparentProxy(true)),
				).
				Install(TestServerUniversal("test-server-mtls", meshName,
					WithArgs([]string{"health-check", "tcp"}),
					WithProtocol("tcp"),
					WithServiceName("test-server-mtls")),
				).
				Setup(env.Cluster)
			Expect(err).ToNot(HaveOccurred())
		})
		E2EAfterAll(func() {
			Expect(env.Cluster.DeleteMeshApps(meshName)).To(Succeed())
			Expect(env.Cluster.DeleteMesh(meshName)).To(Succeed())
		})

		It("should mark host as unhealthy if it doesn't reply on health checks when Permissive mTLS enabled", func() {
			// check that test-server-mtls is healthy
			cmd := []string{"/bin/bash", "-c", "\"echo request | nc test-server-mtls.mesh 80\""}
			stdout, _, err := env.Cluster.ExecWithRetries("", "", "dp-demo-client-mtls", cmd...)
			Expect(err).ToNot(HaveOccurred())
			Expect(stdout).To(ContainSubstring("response"))

			// update HealthCheck policy to check for another 'recv' line
			Expect(YamlUniversal(healthCheck(meshName, "test-server-mtls", "foo", "baz"))(env.Cluster)).To(Succeed())

			// wait cluster 'test-server-mtls' to be marked as unhealthy
			Eventually(func() bool {
				cmd := []string{"/bin/bash", "-c", "\"curl localhost:9901/clusters | grep test-server-mtls\""}
				stdout, _, err := env.Cluster.ExecWithRetries("", "", "dp-demo-client-mtls", cmd...)
				if err != nil {
					return false
				}
				return strings.Contains(stdout, "health_flags::/failed_active_hc")
			}, "30s", "500ms").Should(BeTrue())

			cmd = []string{"/bin/bash", "-c", "\"echo request | nc test-server-mtls.mesh 80\""}
			stdout, _, _ = env.Cluster.ExecWithRetries("", "", "dp-demo-client-mtls", cmd...)

			// there is no real attempt to setup a connection with test-server, but Envoy may return either
			// empty response with EXIT_CODE = 0, or  'Ncat: Connection reset by peer.' with EXIT_CODE = 1
			Expect(stdout).To(Or(BeEmpty(), ContainSubstring("Ncat: Connection reset by peer.")))
		})
	}, Ordered)
}
