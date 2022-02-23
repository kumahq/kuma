package gateway

import (
	"encoding/base64"
	"fmt"
	"net"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
)

func GatewayOnUniversal() {
	var cluster *UniversalCluster

	const GatewayPort = "8080"

	ExternalServerUniversal := func(name string) InstallFunc {
		return func(cluster Cluster) error {
			return cluster.DeployApp(
				WithArgs([]string{"test-server", "echo", "--port", "8080", "--instance", name}),
				WithName(name),
				WithoutDataplane(),
				WithVerbose())
		}
	}

	// DeployCluster creates a universal Kuma cluster using the
	// provided options, installing an echo service as well as a
	// gateway and a client container to send HTTP requests.
	DeployCluster := func(opt ...KumaDeploymentOption) {
		opt = append(opt, WithVerbose(), WithEnv("KUMA_EXPERIMENTAL_MESHGATEWAY", "true"))
		cluster = NewUniversalCluster(NewTestingT(), Kuma1, Silent)

		err := NewClusterSetup().
			Install(Kuma(config_core.Standalone, opt...)).
			Install(GatewayClientAppUniversal("gateway-client")).
			Install(EchoServerApp("echo-server", "echo-service", "universal")).
			Install(GatewayProxyUniversal("gateway-proxy")).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())
	}

	GatewayAddress := func(appName string) string {
		return cluster.GetApp(appName).GetIP()
	}

	// Before each test, verify the cluster is up and stable.
	JustBeforeEach(func() {
		Expect(cluster.VerifyKuma()).To(Succeed())

		// Synchronize on the dataplanes coming up.
		Eventually(func(g Gomega) {
			dataplanes, err := cluster.GetKumactlOptions().KumactlList("dataplanes", "default")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(dataplanes).Should(ContainElements("echo-server", "gateway-proxy"))
		}, "60s", "1s").Should(Succeed())
	})

	// Before each test, install the gateway and routes.
	JustBeforeEach(func() {
		Expect(
			cluster.GetKumactlOptions().KumactlApplyFromString(`
type: MeshGateway
mesh: default
name: edge-gateway
selectors:
- match:
    kuma.io/service: edge-gateway
conf:
  listeners:
  - port: 8080
    protocol: HTTP
    hostname: example.kuma.io
    tags:
      hostname: example.kuma.io
`),
		).To(Succeed())

		Expect(
			cluster.GetKumactlOptions().KumactlApplyFromString(`
type: MeshGatewayRoute
mesh: default
name: edge-gateway
selectors:
- match:
    kuma.io/service: edge-gateway
conf:
  http:
    rules:
    - matches:
      - path:
          match: PREFIX
          value: /
      backends:
      - destination:
          kuma.io/service: echo-service
`),
		).To(Succeed())

		Expect(
			cluster.GetKumactlOptions().KumactlList("meshgateways", "default"),
		).To(ContainElement("edge-gateway"))
	})

	E2EAfterEach(func() {
		Expect(cluster.DismissCluster()).ToNot(HaveOccurred())
	})

	Context("when mTLS is disabled", func() {
		BeforeEach(func() {
			DeployCluster()
		})

		It("should proxy simple HTTP requests", func() {
			ProxySimpleRequests(cluster, "universal",
				net.JoinHostPort(GatewayAddress("gateway-proxy"), GatewayPort))
		})
	})

	Context("when mTLS is enabled", func() {
		BeforeEach(func() {
			DeployCluster(OptEnableMeshMTLS)
		})

		It("should proxy simple HTTP requests", func() {
			ProxySimpleRequests(cluster, "universal",
				net.JoinHostPort(GatewayAddress("gateway-proxy"), GatewayPort))
		})

		// In mTLS mode, only the presence of TrafficPermission rules allow services to receive
		// traffic, so removing the permission should cause requests to fail. We use this to
		// prove that mTLS is enabled
		It("should fail without TrafficPermission", func() {
			ProxyRequestsWithMissingPermission(cluster,
				net.JoinHostPort(GatewayAddress("gateway-proxy"), GatewayPort))
		})
	})

	Context("when targeting an external service", func() {
		BeforeEach(func() {
			cluster = NewUniversalCluster(NewTestingT(), Kuma1, Silent)
			err := NewClusterSetup().
				Install(Kuma(config_core.Standalone, WithVerbose(), WithEnv("KUMA_EXPERIMENTAL_MESHGATEWAY", "true"))).
				Install(ExternalServerUniversal("external-echo")).
				Install(GatewayClientAppUniversal("gateway-client")).
				Install(GatewayProxyUniversal("gateway-proxy")).
				Install(EchoServerApp("echo-server", "echo-service", "universal")).
				Setup(cluster)
			Expect(err).ToNot(HaveOccurred())
		})

		JustBeforeEach(func() {
			// The suite-level JustBeforeEach adds a default route to the mesh echo server.
			// Add new route to the external echo server.
			Expect(
				cluster.GetKumactlOptions().KumactlApplyFromString(`
type: MeshGatewayRoute
mesh: default
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
`),
			).To(Succeed())

			Expect(
				cluster.GetKumactlOptions().KumactlApplyFromString(fmt.Sprintf(`
type: ExternalService
mesh: default
name: external-service
tags:
  kuma.io/service: external-echo
networking:
  address: "%s"
`, net.JoinHostPort(cluster.GetApp("external-echo").GetIP(), "8080"))),
			).To(Succeed())
		})

		It("should proxy simple HTTP requests", func() {
			ProxySimpleRequests(cluster, "external-echo",
				net.JoinHostPort(GatewayAddress("gateway-proxy"), GatewayPort),
				client.WithPathPrefix("/external"))
		})
	})

	Context("when targeting a HTTPS gateway", func() {
		BeforeEach(func() {
			DeployCluster()
		})

		JustBeforeEach(func() {
			// Delete the default gateway that the test fixtures create.
			Expect(
				cluster.GetKumactlOptions().KumactlDelete("meshgateway", "edge-gateway", "default"),
			).To(Succeed())

			// And replace it with a HTTPS gateway.
			Expect(
				cluster.GetKumactlOptions().KumactlApplyFromString(`
type: MeshGateway
mesh: default
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
`),
			).To(Succeed())

			cert, key, err := CreateCertsFor("example.kuma.io")
			Expect(err).To(Succeed())

			payload := base64.StdEncoding.EncodeToString([]byte(strings.Join([]string{key, cert}, "\n")))

			// Create the TLS secret containing the self-signed certificate and corresponding private key.
			Expect(
				cluster.GetKumactlOptions().KumactlApplyFromString(fmt.Sprintf(`
type: Secret
mesh: default
name: example-kuma-io-certificate
data: %s
`, payload)),
			).To(Succeed())

			Expect(
				cluster.GetKumactlOptions().KumactlList("meshgateways", "default"),
			).To(ContainElement("edge-https-gateway"))
		})

		It("should proxy simple HTTPS requests", func() {
			ProxySecureRequests(cluster, "universal",
				net.JoinHostPort("example.kuma.io", GatewayPort),
				client.Resolve("example.kuma.io", 8080, GatewayAddress("gateway-proxy")))
		})
	})

	Context("when a rate limit is configured", func() {
		BeforeEach(func() {
			DeployCluster()
		})

		JustBeforeEach(func() {
			Expect(
				cluster.GetKumactlOptions().KumactlApplyFromString(`
type: RateLimit
mesh: default
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
`),
			).To(Succeed())
		})

		It("should be rate limited", func() {
			ProxyRequestsWithRateLimit(cluster,
				net.JoinHostPort(GatewayAddress("gateway-proxy"), GatewayPort))
		})
	})
}
