package gateway

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/gruntwork-io/terratest/modules/shell"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

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

	BeforeEach(func() {
		cluster = NewUniversalCluster(NewTestingT(), Kuma1, Silent)
		Expect(cluster).ToNot(BeNil())

		deployOpts := append(KumaUniversalDeployOpts, WithVerbose())

		err := NewClusterSetup().
			Install(Kuma(config_core.Standalone, deployOpts...)).
			Install(GatewayClientUniversal("gateway-client")).
			Install(EchoServerUniversal("echo-server")).
			Install(GatewayProxyUniversal("gateway-proxy")).
			Setup(cluster)

		// OK, this is messed up. The makefile rule that builds the
		// kuma-universal:latest image that is used for e2e tests
		// always rebuilds Kuma with Gateway disabled. This means
		// that by default, no Gateway tests will work, so we use the
		// `WithKumactlFlow` option to detect the unsupported gateway
		// type early (when kumactl creates the dataplane resource),
		// and skip the remaining tests.
		var shellErr *shell.ErrWithCmdOutput
		if errors.As(err, &shellErr) {
			if strings.Contains(shellErr.Output.Combined(), "unsupported gateway type") {
				Skip("kuma-cp builtin Gateway support is not enabled")
			}
		}

		// Otherwise, we expect the cluster build to succeed.
		Expect(err).To(Succeed())

		Expect(cluster.VerifyKuma()).To(Succeed())

		// TODO(jpeach) For how the default traffic route make
		// the gateway generate invalid clusters. Remove this when
		// we yank TrafficRoute support.
		Expect(
			cluster.GetKumactlOptions().KumactlDelete("traffic-route", "route-all-default", "default"),
		).To(Succeed())

		// Synchronize on the dataplanes coming up.
		Eventually(func(g Gomega) {
			dataplanes, err := cluster.GetKumactlOptions().KumactlList("dataplanes", "default")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(dataplanes).Should(ContainElements("echo-server", "gateway-proxy"))
		}, "60s", "1s").Should(Succeed())
	})

	E2EAfterEach(func() {
		Expect(cluster.DismissCluster()).ToNot(HaveOccurred())
	})

	It("should proxy simple HTTP requests", func() {
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

		gateways, err := cluster.GetKumactlOptions().KumactlList("gateways", "default")
		Expect(err).To(Succeed())
		Expect(gateways).To(ContainElement("edge-gateway"))

		Eventually(func(g Gomega) {
			p := path.Join("test", url.PathEscape(GinkgoT().Name()))
			target := fmt.Sprintf("http://%s:8080/%s",
				cluster.GetApp("gateway-proxy").GetIP(), p)

			response, err := testutil.CollectResponse(
				cluster, "gateway-client", target,
				testutil.WithHeader("Host", "example.kuma.io"),
			)

			g.Expect(err).To(Succeed())
			g.Expect(response.Instance).To(Equal("universal"))
			g.Expect(response.Received.Headers["Host"]).To(ContainElement("example.kuma.io"))
		}, "30s", "1s").Should(Succeed())
	})
}
