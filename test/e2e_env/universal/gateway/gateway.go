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

	"github.com/kumahq/kuma/test/e2e_env/universal/env"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
)

func Gateway() {
	var mesh = "gateway"

	const gatewayPort = 8080

	BeforeAll(func() {
		setup := NewClusterSetup().
			Install(MeshUniversal(mesh)).
			Install(gatewayClientAppUniversal("gateway-client")).
			Install(echoServerApp(mesh, "echo-server", "echo-service", "universal")).
			Install(GatewayProxyUniversal(mesh, "gateway-proxy")).
			Install(YamlUniversal(MkGateway("edge-gateway", mesh, false, "example.kuma.io", "echo-service", gatewayPort)))

		Expect(setup.Setup(env.Cluster)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(env.Cluster.DeleteMeshApps(mesh)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(mesh)).To(Succeed())
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
			Expect(env.Cluster.Install(MTLSMeshUniversal(mesh))).To(Succeed())

			proxySimpleRequests(env.Cluster, "universal",
				GatewayAddressPort("gateway-proxy", gatewayPort), "example.kuma.io")
		})
	})

	Context("when targeting an external service", func() {
		BeforeAll(func() {
			Expect(
				env.Cluster.DeployApp(
					WithArgs([]string{"test-server", "echo", "--port", "8080", "--instance", "external-echo"}),
					WithName("external-echo-"+mesh),
					WithMesh(mesh),
					WithoutDataplane(),
					WithVerbose()),
			).To(Succeed())

			Expect(
				env.Cluster.Install(YamlUniversal(fmt.Sprintf(`
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
`, mesh))),
			).To(Succeed())

			Expect(
				env.Cluster.Install(YamlUniversal(fmt.Sprintf(`
type: ExternalService
mesh: %s
name: external-service
tags:
  kuma.io/service: external-echo
networking:
  address: "%s"
`, mesh, net.JoinHostPort(env.Cluster.GetApp("external-echo-gateway").GetIP(), "8080")))),
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
				env.Cluster.Install(YamlUniversal(fmt.Sprintf(`
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
`, mesh))),
			).To(Succeed())
		})
	})

	Context("when a rate limit is configured", func() {
		BeforeAll(func() {
			Expect(
				env.Cluster.Install(YamlUniversal(fmt.Sprintf(`
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
`, mesh))),
			).To(Succeed())
		})

		It("should be rate limited", func() {
			gatewayAddr := GatewayAddressPort("gateway-proxy", gatewayPort)
			Logf("expecting 429 response from %q", gatewayAddr)
			Eventually(func(g Gomega) {
				target := fmt.Sprintf("http://%s/%s",
					gatewayAddr, path.Join("test", url.PathEscape(GinkgoT().Name())),
				)

				response, err := client.CollectFailure(env.Cluster, "gateway-client", target, client.WithHeader("Host", "example.kuma.io"))

				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.ResponseCode).To(Equal(429))
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
				env.Cluster.Install(YamlUniversal(fmt.Sprintf(`
type: Secret
mesh: %s
name: example-kuma-io-certificate
data: %s
	`, mesh, payload))),
			).To(Succeed())

			// Add HTTPS listeners
			Expect(
				env.Cluster.Install(YamlUniversal(fmt.Sprintf(`
type: MeshGateway
mesh: %s
name: edge-gateway
selectors:
- match:
    kuma.io/service: edge-gateway
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
		})

		It("should proxy simple HTTPS requests with Host header", func() {
			addr := net.JoinHostPort("example.kuma.io", strconv.Itoa(9080))
			proxySecureRequests(
				env.Cluster,
				"universal",
				addr,
				client.Resolve(addr, GatewayAddress("gateway-proxy")),
			)
		})

		It("should proxy simple HTTPS requests without hostname", func() {
			proxySecureRequests(
				env.Cluster,
				"universal",
				GatewayAddressPort("gateway-proxy", 9081),
			)
		})
	})

	It("really uses mTLS", func() {
		// In mTLS mode, only the presence of TrafficPermission rules allow services to receive
		// traffic, so removing the permission should cause requests to fail. We use this to
		// prove that mTLS is enabled
		PermissionName := "allow-all-" + mesh

		Logf("deleting TrafficPermission %q", PermissionName)
		Expect(env.Cluster.GetKumactlOptions().KumactlDelete(
			"traffic-permission", PermissionName, mesh),
		).To(Succeed())

		gatewayAddr := GatewayAddressPort("gateway-proxy", gatewayPort)
		host := "example.kuma.io"
		Logf("expecting 503 response from %q", gatewayAddr)
		Eventually(func(g Gomega) {
			target := fmt.Sprintf("http://%s/%s",
				gatewayAddr, path.Join("test", url.PathEscape(GinkgoT().Name())),
			)

			status, err := client.CollectFailure(env.Cluster, "gateway-client", target, client.WithHeader("Host", host))

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(status.ResponseCode).To(Equal(503))
		}, "30s", "1s").Should(Succeed())
	})
}
