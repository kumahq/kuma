package gateway

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/universal/env"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
)

func demoClientName(mesh string) string {
	return fmt.Sprintf("demo-client-%s", mesh)
}

func successfullyProxyRequestToGateway(cluster Cluster, clientName, instance, gatewayAddr string, opt ...client.CollectResponsesOptsFn) {
	Logf("expecting 200 response from %q", gatewayAddr)
	target := fmt.Sprintf("http://%s/%s",
		gatewayAddr, path.Join("test", url.PathEscape(GinkgoT().Name())),
	)

	response, err := client.CollectResponse(
		cluster, clientName, target,
		opt...,
	)

	Expect(err).NotTo(HaveOccurred())
	Expect(response.Instance).To(Equal(instance))
}

func failToProxyRequestToGateway(cluster Cluster, containerName, gatewayAddr, host string) func(Gomega) {
	return func(g Gomega) {
		Logf("expecting 200 response from %q", gatewayAddr)
		target := fmt.Sprintf("http://%s/%s",
			gatewayAddr, path.Join("test", url.PathEscape(GinkgoT().Name())),
		)

		response, err := client.CollectFailure(
			cluster, containerName, target,
			client.Resolve(gatewayAddr, host),
		)

		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(response.Exitcode).To(Or(Equal(56), Equal(7), Equal(28)))
	}
}

// DemoClientDataplaneWithOutbound is taken from DemoClientDataplane and adds
// another outbound.
func DemoClientDataplaneWithOutbound(name, mesh, outboundService, outboundMesh string, port int) string {
	return fmt.Sprintf(`
type: Dataplane
mesh: %s
name: {{ name }}
networking:
  address: {{ address }}
  inbound:
  - port: %s
    servicePort: %s
    tags:
      kuma.io/service: %s
      team: client-owners
  outbound:
  - port: 4000
    tags:
      kuma.io/service: echo-server_kuma-test_svc_%s
  - port: 4001
    tags:
      kuma.io/service: echo-server_kuma-test_svc_%s
  - port: 5000
    tags:
      kuma.io/service: external-service
  - port: %d
    tags:
      kuma.io/service: %s
      kuma.io/mesh: %s
`, mesh, "13000", "3000", name, "80", "8080", port, outboundService, outboundMesh)
}

func mkGateway(name, mesh string, crossMesh bool, hostname, backendService string, port int) string {
	meshGateway := fmt.Sprintf(`
type: MeshGateway
name: %s
mesh: %s
selectors:
- match:
    kuma.io/service: %s
conf:
  listeners:
  - port: %d
    protocol: HTTP
    crossMesh: %t
    hostname: %s
`, name, mesh, name, port, crossMesh, hostname)

	route := fmt.Sprintf(`
type: MeshGatewayRoute
name: %s
mesh: %s
selectors:
- match:
    kuma.io/service: %s
conf:
  http:
    rules:
    - matches:
      - path:
          match: PREFIX
          value: /
      backends:
      - destination:
          kuma.io/service: %s # Matches the echo-server we deployed.
`, name, mesh, name, backendService)

	return strings.Join([]string{meshGateway, route}, "\n---\n")
}

func mkGatewayDataplane(name, mesh string) InstallFunc {
	return func(cluster Cluster) error {
		token, err := env.Cluster.GetKuma().GenerateDpToken(mesh, name)
		if err != nil {
			return err
		}

		dataplane := fmt.Sprintf(`
type: Dataplane
name: {{ name }}
mesh: %s
networking:
  address:  {{ address }}
  gateway:
    type: BUILTIN
    tags:
      kuma.io/service: %s
`, mesh, name)

		return env.Cluster.DeployApp(
			WithName(name),
			WithToken(token),
			WithVerbose(),
			WithYaml(dataplane),
		)
	}
}
