package healthcheck

import (
	"encoding/base64"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/universal"
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
				Install(TrafficRouteUniversal(meshName)).
				Install(TrafficPermissionUniversal(meshName)).
				Setup(universal.Cluster)
			Expect(err).ToNot(HaveOccurred())
		})
		E2EAfterAll(func() {
			Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
			Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
		})

		It("should mark host as unhealthy if it doesn't reply on health checks", func() {
			// check that test-server is healthy
			Eventually(func(g Gomega) {
				stdout, _, err := client.CollectResponse(
					universal.Cluster, "dp-demo-client", "test-server.mesh/content",
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(stdout).To(ContainSubstring("response"))
			}).Should(Succeed())

			// update HealthCheck policy to check for another status code
			Expect(YamlUniversal(healthCheck(meshName, "are-you-healthy", "500"))(universal.Cluster)).To(Succeed())

			// wait cluster 'test-server' to be marked as unhealthy
			Eventually(func(g Gomega) {
				cmd := []string{"/bin/bash", "-c", "\"curl localhost:9901/clusters | grep test-server\""}
				stdout, _, err := universal.Cluster.Exec("", "", "dp-demo-client", cmd...)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(stdout).To(ContainSubstring("health_flags::/failed_active_hc"))
			}).Should(Succeed())

			// check that test-server is unhealthy
			Consistently(func(g Gomega) {
				response, err := client.CollectFailure(
					universal.Cluster, "dp-demo-client", "test-server.mesh/content",
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.ResponseCode).To(Equal(503))
			}).Should(Succeed())
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
				Install(TimeoutUniversal(meshName)).
				Install(RetryUniversal(meshName)).
				Install(TrafficRouteUniversal(meshName)).
				Install(TrafficPermissionUniversal(meshName)).
				Install(CircuitBreakerUniversal(meshName)).
				Setup(universal.Cluster)
			Expect(err).ToNot(HaveOccurred())
		})
		E2EAfterAll(func() {
			Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
			Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
		})

		It("should mark host as unhealthy if it doesn't reply on health checks", func() {
			// check that test-server is healthy
			Eventually(func(g Gomega) {
				cmd := []string{"/bin/bash", "-c", "\"echo request | nc test-server.mesh 80\""}
				stdout, _, err := universal.Cluster.Exec("", "", "dp-demo-client", cmd...)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(stdout).To(ContainSubstring("response"))
			}).Should(Succeed())

			// update HealthCheck policy to check for another 'recv' line
			Expect(YamlUniversal(healthCheck(meshName, "test-server", "foo", "baz"))(universal.Cluster)).To(Succeed())

			// wait cluster 'test-server' to be marked as unhealthy
			Eventually(func(g Gomega) {
				cmd := []string{"/bin/bash", "-c", "\"curl localhost:9901/clusters | grep test-server\""}
				stdout, _, err := universal.Cluster.Exec("", "", "dp-demo-client", cmd...)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(stdout).To(ContainSubstring("health_flags::/failed_active_hc"))
			}, "30s", "1s").Should(Succeed())

			Consistently(func(g Gomega) {
				cmd := []string{"/bin/bash", "-c", "\"echo request | nc test-server.mesh 80\""}
				stdout, _, _ := universal.Cluster.Exec("", "", "dp-demo-client", cmd...)

				// there is no real attempt to setup a connection with test-server, but Envoy may return either
				// empty response with EXIT_CODE = 0, or  'Ncat: Connection reset by peer.' with EXIT_CODE = 1
				g.Expect(stdout).To(Or(BeEmpty(), ContainSubstring("Ncat: Connection reset by peer.")))
			}).Should(Succeed())
		})
	}, Ordered)

	Describe("TCP with permissive mTLS", func() {
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
				Install(TimeoutUniversal(meshName)).
				Install(RetryUniversal(meshName)).
				Install(TrafficRouteUniversal(meshName)).
				Install(TrafficPermissionUniversal(meshName)).
				Install(CircuitBreakerUniversal(meshName)).
				Setup(universal.Cluster)
			Expect(err).ToNot(HaveOccurred())
		})
		E2EAfterAll(func() {
			Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
			Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
		})

		It("should mark host as unhealthy if it doesn't reply on health checks", func() {
			// check that test-server-mtls is healthy
			Eventually(func(g Gomega) {
				cmd := []string{"/bin/bash", "-c", "\"echo request | nc test-server-mtls.mesh 80\""}
				stdout, _, err := universal.Cluster.Exec("", "", "dp-demo-client-mtls", cmd...)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(stdout).To(ContainSubstring("response"))
			}).Should(Succeed())

			// update HealthCheck policy to check for another 'recv' line
			Expect(YamlUniversal(healthCheck(meshName, "test-server-mtls", "foo", "baz"))(universal.Cluster)).To(Succeed())

			// wait cluster 'test-server-mtls' to be marked as unhealthy
			Eventually(func(g Gomega) {
				cmd := []string{"/bin/bash", "-c", "\"curl localhost:9901/clusters | grep test-server-mtls\""}
				stdout, _, err := universal.Cluster.Exec("", "", "dp-demo-client-mtls", cmd...)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(stdout).To(ContainSubstring("health_flags::/failed_active_hc"))
			}, "30s", "1s").Should(Succeed())

			Consistently(func(g Gomega) {
				cmd := []string{"/bin/bash", "-c", "\"echo request | nc test-server-mtls.mesh 80\""}
				stdout, _, _ := universal.Cluster.Exec("", "", "dp-demo-client-mtls", cmd...)
				// there is no real attempt to setup a connection with test-server, but Envoy may return either
				// empty response with EXIT_CODE = 0, or  'Ncat: Connection reset by peer.' with EXIT_CODE = 1
				g.Expect(stdout).To(Or(BeEmpty(), ContainSubstring("Ncat: Connection reset by peer.")))
			}).Should(Succeed())
		})
	}, Ordered)
}
