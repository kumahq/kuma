package gateway

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/url"
	"path"
	"strings"

	"github.com/gruntwork-io/terratest/modules/shell"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/test/e2e/trafficroute/testutil"
	. "github.com/kumahq/kuma/test/framework"
)

func GatewayOnUniversal() {
	var cluster *UniversalCluster

	EchoServerUniversal := func(name string) InstallFunc {
		return func(cluster Cluster) error {
			const service = "echo-service"
			token, err := cluster.GetKuma().GenerateDpToken("default", service)
			if err != nil {
				return err
			}

			return TestServerUniversal(
				name,
				"default",
				token,
				WithArgs([]string{"echo", "--instance", "universal"}),
				WithServiceName(service),
			)(cluster)
		}
	}

	// GatewayClientUniversal runs an empty container that will
	// function as a client for a gateway.
	GatewayClientUniversal := func(name string) InstallFunc {
		return func(cluster Cluster) error {
			return cluster.DeployApp(WithName(name), WithoutDataplane(), WithVerbose())
		}
	}

	ExternalServerUniversal := func(name string) InstallFunc {
		return func(cluster Cluster) error {
			return cluster.DeployApp(
				WithArgs([]string{"test-server", "echo", "--port", "8080", "--instance", name}),
				WithName(name),
				WithoutDataplane(),
				WithVerbose())
		}
	}

	GatewayProxyUniversal := func(name string) InstallFunc {
		return func(cluster Cluster) error {
			token, err := cluster.GetKuma().GenerateDpToken("default", "edge-gateway")
			if err != nil {
				return err
			}

			dataplaneYaml := `
type: Dataplane
mesh: default
name: {{ name }}
networking:
  address:  {{ address }}
  gateway:
    type: BUILTIN
    tags:
      kuma.io/service: edge-gateway
`
			return cluster.DeployApp(
				WithKumactlFlow(),
				WithName(name),
				WithToken(token),
				WithVerbose(),
				WithYaml(dataplaneYaml),
			)
		}
	}

	SetupCluster := func(setup *ClusterSetup) {
		cluster = NewUniversalCluster(NewTestingT(), Kuma1, Silent)
		Expect(cluster).ToNot(BeNil())

		err := setup.Setup(cluster)

		// The makefile rule that builds the kuma-universal:latest image
		// that is used for e2e tests by default rebuilds Kuma with Gateway
		// disabled. This means that unless BUILD_WITH_EXPERIMENTAL_GATEWAY=Y is
		// set persistently in the environment, Gateway will not be supported.
		// We use the `WithKumactlFlow` option to detect the unsupported gateway
		// type early (when kumactl creates the dataplane resource), and skip
		// the remaining tests.
		var shellErr *shell.ErrWithCmdOutput
		if errors.As(err, &shellErr) {
			if strings.Contains(shellErr.Output.Combined(), "unsupported gateway type") {
				Skip("kuma-cp builtin Gateway support is not enabled")
			}
		}

		// Otherwise, we expect the cluster build to succeed.
		Expect(err).To(Succeed())
	}

	// DeployCluster creates a universal Kuma cluster using the
	// provided options, installing an echo service as well as a
	// gateway and a client container to send HTTP requests.
	DeployCluster := func(opt ...KumaDeploymentOption) {
		opt = append(opt, WithVerbose())

		SetupCluster(NewClusterSetup().
			Install(Kuma(config_core.Standalone, opt...)).
			Install(GatewayClientUniversal("gateway-client")).
			Install(EchoServerUniversal("echo-server")).
			Install(GatewayProxyUniversal("gateway-proxy")),
		)
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
type: Gateway
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
type: GatewayRoute
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
			cluster.GetKumactlOptions().KumactlList("gateways", "default"),
		).To(ContainElement("edge-gateway"))
	})

	E2EAfterEach(func() {
		Expect(cluster.DismissCluster()).ToNot(HaveOccurred())
	})

	// ProxySimpleRequests tests that basic HTTP requests are proxied to a service.
	ProxySimpleRequests := func(prefix string, instance string) func() {
		return func() {
			Eventually(func(g Gomega) {
				target := fmt.Sprintf("http://%s/%s",
					net.JoinHostPort(cluster.GetApp("gateway-proxy").GetIP(), "8080"),
					path.Join(prefix, "test", url.PathEscape(GinkgoT().Name())),
				)

				response, err := testutil.CollectResponse(
					cluster, "gateway-client", target,
					testutil.WithHeader("Host", "example.kuma.io"),
				)

				g.Expect(err).To(Succeed())
				g.Expect(response.Instance).To(Equal(instance))
				g.Expect(response.Received.Headers["Host"]).To(ContainElement("example.kuma.io"))
			}, "30s", "1s").Should(Succeed())
		}
	}

	Context("when mTLS is disabled", func() {
		BeforeEach(func() {
			DeployCluster(KumaUniversalDeployOpts...)
		})

		It("should proxy simple HTTP requests", ProxySimpleRequests("/", "universal"))
	})

	Context("when mTLS is enabled", func() {
		BeforeEach(func() {
			mtls := WithMeshUpdate("default", func(mesh *mesh_proto.Mesh) *mesh_proto.Mesh {
				mesh.Mtls = &mesh_proto.Mesh_Mtls{
					EnabledBackend: "builtin",
					Backends: []*mesh_proto.CertificateAuthorityBackend{
						{Name: "builtin", Type: "builtin"},
					},
				}
				return mesh
			})

			DeployCluster(append(KumaUniversalDeployOpts, mtls)...)
		})

		It("should proxy simple HTTP requests", ProxySimpleRequests("/", "universal"))

		// In mTLS mode, only the presence of TrafficPermission rules allow services to receive
		// traffic, so removing the permission should cause requests to fail. We use this to
		// prove that mTLS is enabled
		It("should fail without TrafficPermission", func() {
			Expect(
				cluster.GetKumactlOptions().KumactlDelete(
					"traffic-permission", "allow-all-default", "default"),
			).To(Succeed())

			Eventually(func(g Gomega) {
				target := fmt.Sprintf("http://%s/%s",
					net.JoinHostPort(cluster.GetApp("gateway-proxy").GetIP(), "8080"),
					path.Join("test", url.PathEscape(GinkgoT().Name())),
				)

				status, err := testutil.CollectFailure(
					cluster, "gateway-client", target,
					testutil.WithHeader("Host", "example.kuma.io"),
				)

				g.Expect(err).To(Succeed())
				g.Expect(status.ResponseCode).To(Equal(503))
			}, "30s", "1s").Should(Succeed())
		})
	})

	Context("when targeting an external service", func() {
		BeforeEach(func() {
			opt := append(KumaUniversalDeployOpts, WithVerbose())
			SetupCluster(NewClusterSetup().
				Install(Kuma(config_core.Standalone, opt...)).
				Install(ExternalServerUniversal("external-echo")).
				Install(GatewayClientUniversal("gateway-client")).
				Install(GatewayProxyUniversal("gateway-proxy")).
				Install(EchoServerUniversal("echo-server")),
			)
		})

		JustBeforeEach(func() {
			// The suite-level JustBeforeEach adds a default route to the mesh echo server.
			// Add new route to the external echo server.
			Expect(
				cluster.GetKumactlOptions().KumactlApplyFromString(`
type: GatewayRoute
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

		It("should proxy simple HTTP requests",
			ProxySimpleRequests("/external", "external-echo"))
	})

	Context("when targeting a HTTPS gateway", func() {
		BeforeEach(func() {
			DeployCluster(KumaUniversalDeployOpts...)
		})

		JustBeforeEach(func() {
			// Delete the default gateway that the test fixtures create.
			Expect(
				cluster.GetKumactlOptions().KumactlDelete("gateway", "edge-gateway", "default"),
			).To(Succeed())

			// And replace it with a HTTPS gateway.
			Expect(
				cluster.GetKumactlOptions().KumactlApplyFromString(`
type: Gateway
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
				cluster.GetKumactlOptions().KumactlList("gateways", "default"),
			).To(ContainElement("edge-https-gateway"))
		})

		It("should proxy simple HTTPS requests", func() {
			Eventually(func(g Gomega) {
				target := fmt.Sprintf("https://%s/%s",
					net.JoinHostPort("example.kuma.io", "8080"),
					path.Join("https", "test", url.PathEscape(GinkgoT().Name())),
				)

				response, err := testutil.CollectResponse(
					cluster, "gateway-client", target,
					testutil.Insecure(),
					testutil.Resolve("example.kuma.io", 8080, cluster.GetApp("gateway-proxy").GetIP()),
					testutil.WithHeader("Host", "example.kuma.io"),
				)

				g.Expect(err).To(Succeed())
				g.Expect(response.Instance).To(Equal("universal"))
				g.Expect(response.Received.Headers["Host"]).To(ContainElement("example.kuma.io"))
			}, "30s", "1s").Should(Succeed())
		})
	})
}
