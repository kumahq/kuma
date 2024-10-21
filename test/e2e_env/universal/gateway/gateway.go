package gateway

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/url"
	"path"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshfaultinjection_api "github.com/kumahq/kuma/pkg/plugins/policies/meshfaultinjection/api/v1alpha1"
	meshratelimit_api "github.com/kumahq/kuma/pkg/plugins/policies/meshratelimit/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func Gateway() {
	mesh := "gateway"

	const gatewayPort = 8080
	const gateway2Port = 8088

	BeforeAll(func() {
		setup := NewClusterSetup().
			Install(MeshUniversal(mesh)).
			Install(GatewayClientAppUniversal("gateway-client")).
			Install(echoServerApp(mesh, "echo-server", "echo-service", "universal")).
			Install(GatewayProxyUniversal(mesh, "gateway-proxy")).
			Install(YamlUniversal(MkGateway("gateway-proxy", mesh, "gateway-proxy", false, "example.kuma.io", "echo-service", gatewayPort))).
			Install(GatewayProxyUniversal(mesh, "second-gateway-proxy")).
			Install(YamlUniversal(MkGateway("second-gateway-proxy", mesh, "second-gateway-proxy", false, "test.kuma.io", "echo-service", gateway2Port)))

		Expect(setup.Setup(universal.Cluster)).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, mesh)
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteApp("gateway-client")).To(Succeed())
		Expect(universal.Cluster.DeleteMeshApps(mesh)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(mesh)).To(Succeed())
	})

	GatewayAddress := func(appName string) string {
		return universal.Cluster.GetApp(appName).GetIP()
	}

	GatewayAddressPort := func(appName string, port int) string {
		ip := universal.Cluster.GetApp(appName).GetIP()
		return net.JoinHostPort(ip, strconv.Itoa(port))
	}

	Context("when mTLS is disabled", func() {
		It("should proxy simple HTTP requests", func() {
			Eventually(ProxySimpleRequests(universal.Cluster, "universal",
				GatewayAddressPort("gateway-proxy", gatewayPort), "example.kuma.io"), "60s", "1s").Should(Succeed())
		})
	})

	Context("when mTLS is enabled", func() {
		It("should proxy simple HTTP requests", func() {
			Expect(universal.Cluster.Install(MTLSMeshUniversal(mesh))).To(Succeed())
			Expect(universal.Cluster.Install(MeshTrafficPermissionAllowAllUniversal(mesh))).To(Succeed())

			Eventually(ProxySimpleRequests(universal.Cluster, "universal",
				GatewayAddressPort("gateway-proxy", gatewayPort), "example.kuma.io"), "60s", "1s").Should(Succeed())
		})

		AfterAll(func() {
			Expect(DeleteMeshResources(universal.Cluster, mesh, v1alpha1.MeshTrafficPermissionResourceTypeDescriptor)).To(Succeed())
		})
	})

	Context("when targeting an external service", func() {
		BeforeAll(func() {
			Expect(universal.Cluster.Install(TestServerExternalServiceUniversal("gateway-ext-service", 8080, false))).To(Succeed())
			Expect(universal.Cluster.Install(YamlUniversal(fmt.Sprintf(`
type: ExternalService
mesh: %s
name: external-service
tags:
  kuma.io/service: external-echo
networking:
  address: "%s"
`, mesh, net.JoinHostPort(universal.Cluster.GetApp("gateway-ext-service").GetIP(), "8080")))),
			).To(Succeed())
			Expect(universal.Cluster.Install(TrafficRouteUniversal(mesh))).To(Succeed())
			Expect(universal.Cluster.Install(TrafficPermissionUniversal(mesh))).To(Succeed())
			Expect(
				universal.Cluster.Install(YamlUniversal(fmt.Sprintf(`
type: MeshGatewayRoute
mesh: %s
name: external-routes
selectors:
- match:
    kuma.io/service: gateway-proxy
conf:
  http:
    rules:
    - matches:
      - path:
          match: PREFIX
          value: /external
      backends:
      - destination:
          kuma.io/service: external-echo
`, mesh))),
			).To(Succeed())
		})

		AfterAll(func() {
			Expect(DeleteMeshResources(universal.Cluster, mesh, core_mesh.TrafficPermissionResourceTypeDescriptor)).To(Succeed())
			Expect(DeleteMeshResources(universal.Cluster, mesh, core_mesh.TrafficRouteResourceTypeDescriptor)).To(Succeed())
			Expect(DeleteMeshResources(universal.Cluster, mesh, core_mesh.ExternalServiceResourceTypeDescriptor)).To(Succeed())
			Expect(universal.Cluster.DeleteApp("gateway-ext-service")).To(Succeed())
		})

		It("should proxy simple HTTP requests", func() {
			Eventually(ProxySimpleRequests(universal.Cluster, "gateway-ext-service",
				GatewayAddressPort("gateway-proxy", gatewayPort), "example.kuma.io",
				client.WithPathPrefix("/external")), "60s", "1s").Should(Succeed())
		})
	})

	Context("applying ProxyTemplate", func() {
		It("shouldn't error with gateway-proxy import", func() {
			Expect(
				universal.Cluster.Install(YamlUniversal(fmt.Sprintf(`
type: ProxyTemplate
mesh: %s
name: gateway-proxy
selectors:
  - match:
      kuma.io/service: gateway-proxy
conf:
  imports:
    - gateway-proxy
  modifications:
    - cluster:
        operation: add
        value: |
          name: test-cluster
          connectTimeout: 5s
          type: STATIC
`, mesh))),
			).To(Succeed())
		})
	})

	Context("when a rate limit is configured", func() {
		BeforeAll(func() {
			Expect(
				universal.Cluster.Install(YamlUniversal(fmt.Sprintf(`
type: RateLimit
mesh: %s
name: echo-rate-limit
sources:
- match:
    kuma.io/service: gateway-proxy
destinations:
- match:
    kuma.io/service: echo-service
conf:
  http:
    requests: 5
    interval: 10s
`, mesh))),
			).To(Succeed())
			Expect(universal.Cluster.Install(TrafficRouteUniversal(mesh))).To(Succeed())
		})
		AfterAll(func() {
			Expect(DeleteMeshResources(universal.Cluster, mesh, core_mesh.RateLimitResourceTypeDescriptor)).To(Succeed())
			Expect(DeleteMeshResources(universal.Cluster, mesh, core_mesh.TrafficRouteResourceTypeDescriptor)).To(Succeed())
		})

		It("should be rate limited", func() {
			gatewayAddr := GatewayAddressPort("gateway-proxy", gatewayPort)
			Logf("expecting 429 response from %q", gatewayAddr)
			Eventually(func(g Gomega) {
				target := fmt.Sprintf("http://%s/%s",
					gatewayAddr, path.Join("test", url.PathEscape(GinkgoT().Name())),
				)

				response, err := client.CollectFailure(universal.Cluster, "gateway-client", target, client.WithHeader("Host", "example.kuma.io"))

				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.ResponseCode).To(Equal(429))
			}, "30s", "1s").Should(Succeed())
		})
	})

	Context("when a MeshRateLimit is configured", func() {
		BeforeAll(func() {
			Expect(
				universal.Cluster.Install(YamlUniversal(fmt.Sprintf(`
type: MeshRateLimit
mesh: "%s"
name: mesh-rate-limit-all-sources
spec:
  targetRef:
    kind: MeshGateway
    name: gateway-proxy
  to:
    - targetRef:
        kind: Mesh
      default:
        local:
          http:
            requestRate:
              num: 1
              interval: 10s
            onRateLimit:
              status: 428
              headers:
                add:
                - name: "x-kuma-rate-limited"
                  value: "true"`, mesh))),
			).To(Succeed())
		})
		AfterAll(func() {
			Expect(DeleteMeshResources(universal.Cluster, mesh, meshratelimit_api.MeshRateLimitResourceTypeDescriptor)).To(Succeed())
		})

		It("should be rate limited", func() {
			gatewayAddr := GatewayAddressPort("gateway-proxy", gatewayPort)
			Logf("expecting 428 response from %q", gatewayAddr)
			Eventually(func(g Gomega) {
				target := fmt.Sprintf("http://%s/%s",
					gatewayAddr, path.Join("test", url.PathEscape(GinkgoT().Name())),
				)

				response, err := client.CollectFailure(universal.Cluster, "gateway-client", target, client.WithHeader("Host", "example.kuma.io"))

				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.ResponseCode).To(Equal(428))
			}, "30s", "1s").Should(Succeed())
		})
	})

	Context("when a MeshFaultInjection is configured", func() {
		E2EAfterEach(func() {
			Expect(DeleteMeshResources(universal.Cluster, mesh, meshfaultinjection_api.MeshFaultInjectionResourceTypeDescriptor)).To(Succeed())
		})

		It("should return custom error code for all requests", func() {
			// give
			Expect(
				universal.Cluster.Install(YamlUniversal(fmt.Sprintf(`
type: MeshFaultInjection
mesh: "%s"
name: mesh-fault-injection-all-sources
spec:
  targetRef:
    kind: MeshGateway
    name: gateway-proxy
  to:
    - targetRef:
        kind: Mesh
      default:
        http:
          - abort:
              httpStatus: 421
              percentage: 100`, mesh))),
			).To(Succeed())

			gatewayAddr := GatewayAddressPort("gateway-proxy", gatewayPort)
			Logf("expecting 421 response from %q", gatewayAddr)
			Eventually(func(g Gomega) {
				target := fmt.Sprintf("http://%s/%s",
					gatewayAddr, path.Join("test", url.PathEscape(GinkgoT().Name())),
				)

				response, err := client.CollectFailure(universal.Cluster, "gateway-client", target, client.WithHeader("Host", "example.kuma.io"))

				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.ResponseCode).To(Equal(421))
			}, "30s", "1s").Should(Succeed())
		})

		It("should return custom error code for all requests from both gateways", func() {
			// given
			Expect(
				universal.Cluster.Install(YamlUniversal(fmt.Sprintf(`
type: MeshFaultInjection
mesh: "%s"
name: mesh-fault-injection-all-sources-all-gateways
spec:
  targetRef:
    kind: Mesh
    proxyTypes: ["Gateway"]
  to:
    - targetRef:
        kind: Mesh
      default:
        http:
          - abort:
              httpStatus: 422
              percentage: 100`, mesh))),
			).To(Succeed())

			// then
			gatewayAddr := GatewayAddressPort("gateway-proxy", gatewayPort)
			Logf("expecting 422 response from %q", gatewayAddr)
			Eventually(func(g Gomega) {
				target := fmt.Sprintf("http://%s/%s",
					gatewayAddr, path.Join("test", url.PathEscape(GinkgoT().Name())),
				)

				response, err := client.CollectFailure(universal.Cluster, "gateway-client", target, client.WithHeader("Host", "example.kuma.io"))

				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.ResponseCode).To(Equal(422))
			}, "30s", "1s").Should(Succeed())

			// and second gateway returns the same error
			gatewayAddr = GatewayAddressPort("second-gateway-proxy", gateway2Port)
			Logf("expecting 422 response from %q", gatewayAddr)
			Eventually(func(g Gomega) {
				target := fmt.Sprintf("http://%s/%s",
					gatewayAddr, path.Join("test", url.PathEscape(GinkgoT().Name())),
				)

				response, err := client.CollectFailure(universal.Cluster, "gateway-client", target, client.WithHeader("Host", "test.kuma.io"))

				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.ResponseCode).To(Equal(422))
			}, "30s", "1s").Should(Succeed())
		})
	})

	Context("when targeting a HTTPS gateway listener", func() {
		BeforeAll(func() {
			cert, key, err := CreateCertsFor("example.kuma.io")
			Expect(err).ToNot(HaveOccurred())

			payload := base64.StdEncoding.EncodeToString([]byte(strings.Join([]string{key, cert}, "\n")))

			// Create the TLS secret containing the self-signed certificate and corresponding private key.
			Expect(
				universal.Cluster.Install(YamlUniversal(fmt.Sprintf(`
type: Secret
mesh: %s
name: example-kuma-io-certificate
data: %s
	`, mesh, payload))),
			).To(Succeed())

			// Add HTTPS listeners
			Expect(
				universal.Cluster.Install(YamlUniversal(fmt.Sprintf(`
type: MeshGateway
mesh: %s
name: gateway-proxy
selectors:
- match:
    kuma.io/service: gateway-proxy
conf:
  listeners:
  - port: 8080
    protocol: HTTP
    hostname: example.kuma.io
  - port: 9080
    protocol: HTTPS
    hostname: example.kuma.io
    tls:
      mode: TERMINATE
      certificates:
      - secret: example-kuma-io-certificate
    tags:
      hostname: example.kuma.io
  - port: 9081
    protocol: HTTPS
    tls:
      mode: TERMINATE
      certificates:
      - secret: example-kuma-io-certificate
`, mesh))),
			).To(Succeed())
			Expect(universal.Cluster.Install(MeshTrafficPermissionAllowAllUniversal(mesh))).To(Succeed())
		})

		AfterAll(func() {
			Expect(DeleteMeshResources(universal.Cluster, mesh, v1alpha1.MeshTrafficPermissionResourceTypeDescriptor)).To(Succeed())
		})

		It("should proxy simple HTTPS requests with Host header", func() {
			addr := net.JoinHostPort("example.kuma.io", strconv.Itoa(9080))
			Eventually(proxySecureRequests(
				universal.Cluster,
				"universal",
				addr,
				client.Resolve(addr, GatewayAddress("gateway-proxy")),
			), "60s", "1s").Should(Succeed())
		})

		It("should proxy simple HTTPS requests without hostname", func() {
			Eventually(proxySecureRequests(
				universal.Cluster, "universal",
				GatewayAddressPort("gateway-proxy", 9081)), "1m", "1s").Should(Succeed())
		})
	})

	It("really uses mTLS", func() {
		gatewayAddr := GatewayAddressPort("gateway-proxy", gatewayPort)
		host := "example.kuma.io"
		Logf("expecting 503 response from %q", gatewayAddr)
		Eventually(func(g Gomega) {
			target := fmt.Sprintf("http://%s/%s",
				gatewayAddr, path.Join("test", url.PathEscape(GinkgoT().Name())),
			)

			status, err := client.CollectFailure(universal.Cluster, "gateway-client", target, client.WithHeader("Host", host))

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(status.ResponseCode).To(Equal(403))
		}, "30s", "1s").Should(Succeed())
	})
}
