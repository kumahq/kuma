package gateway

import (
	"encoding/base64"
	"fmt"
	"net"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/universal/env"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
)

func Gateway() {
	var mesh = "gateway"
	var mtlsMesh = "gateway-mtls"

	const gatewayPort = 8080
	const mtlsGatewayPort = 8081

	ExternalServerUniversal := func(mesh string) InstallFunc {
		return func(cluster Cluster) error {
			return cluster.DeployApp(
				WithArgs([]string{"test-server", "echo", "--port", "8080", "--instance", "external-echo"}),
				WithName("external-echo-"+mesh),
				WithMesh(mesh),
				WithoutDataplane(),
				WithVerbose())
		}
	}

	// DeployCluster creates a universal Kuma cluster using the
	// provided options, installing an echo service as well as a
	// gateway and a client container to send HTTP requests.
	BeforeAll(func() {
		setup := NewClusterSetup().
			Install(MeshUniversal(mesh)).
			Install(MTLSMeshUniversal(mtlsMesh)).
			Install(gatewayClientAppUniversal("gateway-client")).
			Install(echoServerApp(mesh, "echo-server", "echo-service", "universal")).
			Install(echoServerApp(mtlsMesh, "echo-server-mtls", "echo-service", "universal")).
			Install(GatewayProxyUniversal(mesh, "gateway-proxy")).
			Install(GatewayProxyUniversal(mtlsMesh, "gateway-proxy-mtls")).
			Install(ExternalServerUniversal(mesh))

		Expect(setup.Setup(env.Cluster)).To(Succeed())

		Expect(
			env.Cluster.GetKumactlOptions().KumactlApplyFromString(MkGateway("edge-gateway", mesh, false, "example.kuma.io", "echo-service", gatewayPort)),
		).To(Succeed())

		Expect(
			env.Cluster.GetKumactlOptions().KumactlApplyFromString(MkGateway("edge-gateway", mtlsMesh, false, "example-mtls.kuma.io", "echo-service", mtlsGatewayPort)),
		).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(env.Cluster.DeleteMeshApps(mesh)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(mesh)).To(Succeed())
		Expect(env.Cluster.DeleteMeshApps(mtlsMesh)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(mtlsMesh)).To(Succeed())
	})

	GatewayAddress := func(appName string) string {
		return env.Cluster.GetApp(appName).GetIP()
	}

	GatewayAddressPort := func(appName string, port int) string {
		ip := env.Cluster.GetApp(appName).GetIP()
		return net.JoinHostPort(ip, strconv.Itoa(port))
	}

	Context("when mTLS is disabled", func() {
		It("should proxy simple HTTP requests", func() {
			proxySimpleRequests(env.Cluster, "universal",
				GatewayAddressPort("gateway-proxy", gatewayPort), "example.kuma.io")
		})
	})

	Context("when mTLS is enabled", func() {
		It("should proxy simple HTTP requests", func() {
			proxySimpleRequests(env.Cluster, "universal",
				GatewayAddressPort("gateway-proxy-mtls", mtlsGatewayPort), "example-mtls.kuma.io")
		})

		// In mTLS mode, only the presence of TrafficPermission rules allow services to receive
		// traffic, so removing the permission should cause requests to fail. We use this to
		// prove that mTLS is enabled
		It("should fail without TrafficPermission", func() {
			proxyRequestsWithMissingPermission(env.Cluster, mtlsMesh,
				GatewayAddressPort("gateway-proxy-mtls", mtlsGatewayPort), "example-mtls.kuma.io")
		})
	})

	Context("when targeting an external service", func() {
		BeforeAll(func() {
			// The suite-level JustBeforeEach adds a default route to the mesh echo server.
			// Add new route to the external echo server.
			Expect(
				env.Cluster.GetKumactlOptions().KumactlApplyFromString(fmt.Sprintf(`
type: MeshGatewayRoute
mesh: %s
name: external-routes
selectors:
- match:
    kuma.io/service: edge-gateway
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
`, mesh)),
			).To(Succeed())

			Expect(
				env.Cluster.GetKumactlOptions().KumactlApplyFromString(fmt.Sprintf(`
type: ExternalService
mesh: %s
name: external-service
tags:
  kuma.io/service: external-echo
networking:
  address: "%s"
`, mesh, net.JoinHostPort(env.Cluster.GetApp("external-echo-gateway").GetIP(), "8080"))),
			).To(Succeed())
		})

		It("should proxy simple HTTP requests", func() {
			proxySimpleRequests(env.Cluster, "external-echo",
				GatewayAddressPort("gateway-proxy", gatewayPort), "example.kuma.io",
				client.WithPathPrefix("/external"))
		})
	})
	Context("applying ProxyTemplate", func() {
		It("shouldn't error with gateway-proxy import", func() {
			Expect(
				env.Cluster.GetKumactlOptions().KumactlApplyFromString(fmt.Sprintf(`
type: ProxyTemplate
mesh: %s
name: edge-gateway
selectors:
  - match:
      kuma.io/service: edge-gateway
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
`, mesh)),
			).To(Succeed())
		})
	})

	Context("when a rate limit is configured", func() {
		BeforeAll(func() {
			Expect(
				env.Cluster.GetKumactlOptions().KumactlApplyFromString(fmt.Sprintf(`
type: RateLimit
mesh: %s
name: echo-rate-limit
sources:
- match:
    kuma.io/service: edge-gateway
destinations:
- match:
    kuma.io/service: echo-service
conf:
  http:
    requests: 5
    interval: 10s
`, mesh)),
			).To(Succeed())
		})

		It("should be rate limited", func() {
			proxyRequestsWithRateLimit(env.Cluster,
				GatewayAddressPort("gateway-proxy", gatewayPort))
		})
	})

	Context("when targeting a HTTPS gateway", func() {
		BeforeAll(func() {
			// Delete the default gateway that the test fixtures create.
			Expect(
				env.Cluster.GetKumactlOptions().KumactlDelete("meshgateway", "edge-gateway", mesh),
			).To(Succeed())

			// And replace it with a HTTPS gateway.
			Expect(
				env.Cluster.GetKumactlOptions().KumactlApplyFromString(fmt.Sprintf(`
type: MeshGateway
mesh: %s
name: edge-https-gateway
selectors:
- match:
    kuma.io/service: edge-gateway
conf:
  listeners:
  - port: 8080
    protocol: HTTPS
    hostname: example.kuma.io
    tls:
      mode: TERMINATE
      certificates:
      - secret: example-kuma-io-certificate
    tags:
      hostname: example.kuma.io
  - port: 8081
    protocol: HTTPS
    tls:
      mode: TERMINATE
      certificates:
      - secret: example-kuma-io-certificate
`, mesh)),
			).To(Succeed())

			cert, key, err := CreateCertsFor("example.kuma.io")
			Expect(err).To(Succeed())

			payload := base64.StdEncoding.EncodeToString([]byte(strings.Join([]string{key, cert}, "\n")))

			// Create the TLS secret containing the self-signed certificate and corresponding private key.
			Expect(
				env.Cluster.GetKumactlOptions().KumactlApplyFromString(fmt.Sprintf(`
type: Secret
mesh: %s
name: example-kuma-io-certificate
data: %s
	`, mesh, payload)),
			).To(Succeed())

			Expect(
				env.Cluster.GetKumactlOptions().KumactlList("meshgateways", mesh),
			).To(ContainElement("edge-https-gateway"))
		})

		It("should proxy simple HTTPS requests", func() {
			addr := net.JoinHostPort("example.kuma.io", strconv.Itoa(8080))
			proxySecureRequests(
				env.Cluster,
				"universal",
				addr,
				client.Resolve(addr, GatewayAddress("gateway-proxy")),
			)

			proxySecureRequests(
				env.Cluster,
				"universal",
				GatewayAddressPort("gateway-proxy", 8081),
			)
		})
	})
}
