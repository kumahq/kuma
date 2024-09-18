package meshhealthcheck

import (
	"encoding/base64"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func MeshHealthCheck() {
	Describe("HTTP", func() {
		meshName := "meshhealthcheck-http"
		healthCheck := func(mesh, method, status string) string {
			return fmt.Sprintf(`
type: MeshHealthCheck
mesh: %s
name: everything-to-backend
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshService
        name: test-server
      default:
        interval: 10s
        timeout: 2s
        unhealthyThreshold: 3
        healthyThreshold: 1
        failTrafficOnPanic: true
        noTrafficInterval: 1s
        healthyPanicThreshold: 0
        reuseConnection: true
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
				Install(TestServerUniversal("test-server", meshName, WithArgs([]string{"health-check", "http"}), WithProtocol(mesh.ProtocolHTTP))).
				Setup(universal.Cluster)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEachFailure(func() {
			DebugUniversal(universal.Cluster, meshName)
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
			}, "30s", "500ms").Should(Succeed())

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

	Describe("HTTP to real MeshService", func() {
		meshName := "meshhealthcheck-http-ms"
		healthCheck := func(mesh, method, status string) string {
			return fmt.Sprintf(`
type: MeshHealthCheck
mesh: %s
name: everything-to-backend
spec:
  to:
    - targetRef:
        kind: MeshService
        name: test-server
      default:
        interval: 10s
        timeout: 2s
        unhealthyThreshold: 3
        healthyThreshold: 1
        failTrafficOnPanic: true
        noTrafficInterval: 1s
        healthyPanicThreshold: 0
        reuseConnection: true
        http: 
          path: /%s
          expectedStatuses: 
          - %s`, mesh, method, status)
		}

		uniServiceYAML := fmt.Sprintf(`
type: MeshService
name: test-server
mesh: %s
labels:
  kuma.io/origin: zone
  kuma.io/env: universal
spec:
  selector:
    dataplaneTags:
      kuma.io/service: test-server
  ports:
  - port: 80
    targetPort: 80
    appProtocol: http
`, meshName)
		BeforeAll(func() {
			err := NewClusterSetup().
				Install(MeshUniversal(meshName)).
				Install(YamlUniversal(healthCheck(meshName, "health", "200"))).
				Install(DemoClientUniversal("dp-demo-client", meshName,
					WithTransparentProxy(true)),
				).
				Install(TestServerUniversal("test-server", meshName, WithArgs([]string{"health-check", "http"}), WithProtocol(mesh.ProtocolHTTP))).
				Install(YamlUniversal(uniServiceYAML)).
				Install(YamlUniversal(`
type: HostnameGenerator
name: uni-ms-mhc
spec:
  template: '{{ .DisplayName }}.universal.ms'
  selector:
    meshService:
      matchLabels:
        kuma.io/origin: zone
        kuma.io/env: universal`)).
				Setup(universal.Cluster)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEachFailure(func() {
			DebugUniversal(universal.Cluster, meshName)
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
			}, "30s", "500ms").Should(Succeed())

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
type: MeshHealthCheck
mesh: %s
name: everything-to-backend
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshService
        name: %s
      default:
        interval: 10s
        timeout: 2s
        unhealthyThreshold: 3
        healthyThreshold: 1
        failTrafficOnPanic: true
        noTrafficInterval: 1s
        healthyPanicThreshold: 0
        reuseConnection: true
        tcp: 
          send: %s
          receive:
          - %s`, mesh, serviceName, sendBase64, recvBase64)
		}
		meshName := "meshhealthcheck-tcp"
		BeforeAll(func() {
			err := NewClusterSetup().
				Install(MeshUniversal(meshName)).
				Install(DemoClientUniversal("dp-demo-client", meshName,
					WithTransparentProxy(true)),
				).
				Install(TestServerUniversal("test-server", meshName,
					WithArgs([]string{"health-check", "tcp"}),
					WithProtocol(mesh.ProtocolTCP)),
				).
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
			}, "30s", "500ms").Should(Succeed())

			Consistently(func(g Gomega) {
				cmd := []string{"/bin/bash", "-c", "\"echo request | nc test-server.mesh 80\""}
				stdout, _, _ := universal.Cluster.Exec("", "", "dp-demo-client", cmd...)

				// there is no real attempt to setup a connection with test-server, but Envoy may return either
				// empty response with EXIT_CODE = 0, or  'Ncat: Connection reset by peer.' with EXIT_CODE = 1
				g.Expect(stdout).To(Or(BeEmpty(), ContainSubstring("Ncat: Connection reset by peer.")))
			}).Should(Succeed())
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
type: MeshHealthCheck
mesh: %s
name: gateway-to-backend
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshService
        name: %s
      default:
        interval: 10s
        timeout: 2s
        unhealthyThreshold: 3
        healthyThreshold: 1
        failTrafficOnPanic: true
        noTrafficInterval: 1s
        healthyPanicThreshold: 0
        reuseConnection: true
        tcp: 
          send: %s
          receive:
            - %s`, mesh, serviceName, sendBase64, recvBase64)
		}
		meshName := "meshhealthcheck-mtls-permissive-tcp"
		BeforeAll(func() {
			err := NewClusterSetup().
				Install(mtlsPermissiveMesh(meshName)).
				Install(DemoClientUniversal("dp-demo-client-mtls", meshName,
					WithTransparentProxy(true)),
				).
				Install(TestServerUniversal("test-server-mtls", meshName,
					WithArgs([]string{"health-check", "tcp"}),
					WithProtocol(mesh.ProtocolTCP),
					WithServiceName("test-server-mtls")),
				).
				Install(MeshTrafficPermissionAllowAllUniversal(meshName)).
				Setup(universal.Cluster)
			Expect(err).ToNot(HaveOccurred())
		})
		E2EAfterAll(func() {
			Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
			Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
		})

		It("should mark host as unhealthy if it doesn't reply to health checks when Permissive mTLS enabled", func() {
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
			}, "30s", "500ms").Should(Succeed())

			Consistently(func(g Gomega) {
				cmd := []string{"/bin/bash", "-c", "\"echo request | nc test-server-mtls.mesh 80\""}
				stdout, _, _ := universal.Cluster.Exec("", "", "dp-demo-client-mtls", cmd...)

				// there is no real attempt to setup a connection with test-server, but Envoy may return either
				// empty response with EXIT_CODE = 0, or  'Ncat: Connection reset by peer.' with EXIT_CODE = 1
				g.Expect(stdout).To(Or(BeEmpty(), ContainSubstring("Ncat: Connection reset by peer.")))
			}).Should(Succeed())
		})
	}, Ordered)

	Describe("gRPC", func() {
		meshName := "meshhealthcheck-grpc"
		healthCheck := func(mesh string) string {
			return fmt.Sprintf(`
type: MeshHealthCheck
mesh: %s
name: everything-to-backend
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshService
        name: test-server
      default:
        interval: 5s
        timeout: 1s
        unhealthyThreshold: 2
        healthyThreshold: 1
        failTrafficOnPanic: true
        noTrafficInterval: 1s
        healthyPanicThreshold: 0
        reuseConnection: true
        grpc: {}`, mesh)
		}
		BeforeAll(func() {
			err := NewClusterSetup().
				Install(MeshUniversal(meshName)).
				Install(MeshTrafficPermissionAllowAllUniversal(meshName)).
				Install(YamlUniversal(healthCheck(meshName))).
				Install(TestServerUniversal("test-client", meshName,
					WithServiceName("test-client"),
					WithArgs([]string{"grpc", "client", "--address", "test-server.mesh:80"}),
					WithTransparentProxy(true),
				)).
				Install(TestServerUniversal(
					"test-server",
					meshName,
					WithArgs([]string{"grpc", "server", "--port", "8080"}),
					WithProtocol(mesh.ProtocolGRPC),
					WithTransparentProxy(true),
				)).
				Setup(universal.Cluster)
			Expect(err).ToNot(HaveOccurred())
		})

		E2EAfterAll(func() {
			Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
			Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
		})

		It("should mark host as unhealthy if it doesn't reply to health checks", func() {
			// apply HealthCheck policy
			Expect(YamlUniversal(healthCheck(meshName))(universal.Cluster)).To(Succeed())

			// check that test-server is healthy
			Eventually(func(g Gomega) {
				cmd := []string{"/bin/bash", "-c", "\"curl localhost:9901/clusters | grep test-server\""}
				stdout, _, err := universal.Cluster.Exec("", "", "test-client", cmd...)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(stdout).To(ContainSubstring("health_flags::healthy"))
			}, "30s", "500ms").Should(Succeed())

			// stop grpc server
			cmd := []string{"/bin/bash", "-c", "\"kill -SIGSTOP $(pidof test-server)\""}
			_, _, err := universal.Cluster.Exec("", "", "test-server", cmd...)
			Expect(err).ToNot(HaveOccurred())

			// wait cluster 'test-server' to be marked as unhealthy
			Eventually(func(g Gomega) {
				cmd := []string{"/bin/bash", "-c", "\"curl localhost:9901/clusters | grep test-server\""}
				stdout, _, err := universal.Cluster.Exec("", "", "test-client", cmd...)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(stdout).To(ContainSubstring("health_flags::/failed_active_hc"))
			}, "30s", "500ms").Should(Succeed())

			// resume grpc server
			cmd = []string{"/bin/bash", "-c", "\"kill -SIGCONT $(pidof test-server)\""}
			_, _, err = universal.Cluster.Exec("", "", "test-server", cmd...)
			Expect(err).ToNot(HaveOccurred())

			// wait cluster 'test-server' to be marked as healthy
			Eventually(func(g Gomega) {
				cmd := []string{"/bin/bash", "-c", "\"curl localhost:9901/clusters | grep test-server\""}
				stdout, _, err := universal.Cluster.Exec("", "", "test-client", cmd...)
				Expect(err).ToNot(HaveOccurred())
				Expect(stdout).ToNot(ContainSubstring("health_flags::healthy"))
			}, "30s", "500ms").Should(Succeed())
		})
	}, Ordered)

	Describe("HTTP with MeshHTTPRoute", func() {
		meshName := "meshhealthcheck-http-and-meshhttproute"
		healthCheck := func(mesh, method, status string) string {
			return fmt.Sprintf(`
type: MeshHealthCheck
mesh: %s
name: everything-to-backend
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshService
        name: test-server
      default:
        interval: 10s
        timeout: 2s
        unhealthyThreshold: 3
        healthyThreshold: 1
        failTrafficOnPanic: true
        noTrafficInterval: 1s
        healthyPanicThreshold: 0
        reuseConnection: true
        http: 
          path: /%s
          expectedStatuses: 
          - %s`, mesh, method, status)
		}

		meshHttpRoute := fmt.Sprintf(`
type: MeshHTTPRoute
mesh: %s
name: http-route-1
spec: 
  targetRef: 
    kind: MeshService
    name: dp-demo-client
  to: 
    - targetRef: 
        kind: MeshService
        name: test-server
      rules: 
        - matches: 
            - path: 
                value: /
                type: PathPrefix
          default: 
            backendRefs: 
              - kind: MeshServiceSubset
                name: test-server
                tags: 
                  version: v1
                weight: 50
              - kind: MeshServiceSubset
                name: test-server
                tags: 
                  version: v2
                weight: 50`, meshName)

		BeforeAll(func() {
			err := NewClusterSetup().
				Install(MeshUniversal(meshName)).
				Install(YamlUniversal(healthCheck(meshName, "health", "200"))).
				Install(DemoClientUniversal("dp-demo-client", meshName,
					WithTransparentProxy(true)),
				).
				Install(TestServerUniversal("test-server-1", meshName, WithArgs([]string{"health-check", "http"}), WithProtocol(mesh.ProtocolHTTP), WithServiceVersion("v1"))).
				Install(TestServerUniversal("test-server-2", meshName, WithArgs([]string{"health-check", "http"}), WithProtocol(mesh.ProtocolHTTP), WithServiceVersion("v2"))).
				Setup(universal.Cluster)
			Expect(err).ToNot(HaveOccurred())
			Expect(universal.Cluster.Install(YamlUniversal(meshHttpRoute))).To(Succeed())
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

			// wait for both split clusters to be marked as unhealthy
			Eventually(func(g Gomega) {
				cmd := []string{"/bin/bash", "-c", `"curl --silent localhost:9901/clusters | grep -c 'test-server-.*health_flags::/failed_active_hc'"`}
				stdout, _, err := universal.Cluster.Exec("", "", "dp-demo-client", cmd...)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(strings.TrimSpace(stdout)).To(Equal("2"))
			}, "30s", "500ms").Should(Succeed())

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
}
