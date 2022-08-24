package gateway

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/universal/env"
	"github.com/kumahq/kuma/test/framework"
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

// demoClientDataplaneWithOutbound is taken from DemoClientDataplane and adds
// another outbound.
func demoClientDataplaneWithOutbound(name, mesh, outboundService, outboundMesh string, port int) string {
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

func MkGateway(name, mesh string, crossMesh bool, hostname, backendService string, port int) string {
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
          kuma.io/service: %s
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

// gatewayClientAppUniversal runs an empty container that will
// function as a client for a gateway.
func gatewayClientAppUniversal(name string) InstallFunc {
	return func(cluster Cluster) error {
		return cluster.DeployApp(
			WithName(name),
			WithoutDataplane(),
			WithVerbose(),
		)
	}
}

func echoServerApp(mesh, name, service, instance string) InstallFunc {
	return func(cluster Cluster) error {
		return TestServerUniversal(
			name,
			mesh,
			WithArgs([]string{"echo", "--instance", instance}),
			WithServiceName(service),
		)(cluster)
	}
}

func proxySimpleRequests(cluster Cluster, instance, gateway, host string, opts ...client.CollectResponsesOptsFn) {
	targetPath := path.Join("test", GinkgoT().Name())

	Logf("expecting 200 response from %q", gateway)
	Eventually(func(g Gomega) {
		var escaped []string
		for _, segment := range strings.Split(targetPath, "/") {
			escaped = append(escaped, url.PathEscape(segment))
		}

		target := fmt.Sprintf("http://%s/%s", gateway, path.Join(escaped...))

		opts = append(opts, client.WithHeader("Host", host))
		response, err := client.CollectResponse(cluster, "gateway-client", target, opts...)

		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(response.Instance).To(Equal(instance))
		g.Expect(response.Received.Headers["Host"]).To(ContainElement(host))
	}, "60s", "1s").Should(Succeed())
}

// proxySecureRequests tests that basic HTTPS requests are proxied to the echo-server.
func proxySecureRequests(cluster Cluster, instance string, gateway string, opts ...client.CollectResponsesOptsFn) {
	Logf("expecting 200 response from %q", gateway)
	Eventually(func(g Gomega) {
		target := fmt.Sprintf("https://%s/%s",
			gateway, path.Join("https", "test", url.PathEscape(GinkgoT().Name())),
		)

		opts = append(opts,
			client.Insecure(),
			client.WithHeader("Host", "example.kuma.io"))
		response, err := client.CollectResponse(cluster, "gateway-client", target, opts...)

		g.Expect(err).To(Succeed())
		g.Expect(response.Instance).To(Equal(instance))
		g.Expect(response.Received.Headers["Host"]).To(ContainElement("example.kuma.io"))
	}, "60s", "1s").Should(Succeed())
}

// proxyRequestsWithMissingPermission deletes the default TrafficPermission and expects
// that to cause proxying requests to fail.
//
// In mTLS mode, only the presence of TrafficPermission rules allow services to receive
// traffic, so removing the permission should cause requests to fail. We use this to
// prove that mTLS is enabled.
func proxyRequestsWithMissingPermission(cluster Cluster, mesh, gateway, host string, opts ...client.CollectResponsesOptsFn) {
	PermissionName := "allow-all-" + mesh

	Logf("deleting TrafficPermission %q", PermissionName)
	Expect(cluster.GetKumactlOptions().KumactlDelete(
		"traffic-permission", PermissionName, mesh),
	).To(Succeed())

	Logf("expecting 503 response from %q", gateway)
	Eventually(func(g Gomega) {
		target := fmt.Sprintf("http://%s/%s",
			gateway, path.Join("test", url.PathEscape(GinkgoT().Name())),
		)

		opts = append(opts, client.WithHeader("Host", host))
		status, err := client.CollectFailure(cluster, "gateway-client", target, opts...)

		g.Expect(err).To(Succeed())
		g.Expect(status.ResponseCode).To(Equal(503))
	}, "30s", "1s").Should(Succeed())
}

// proxyRequestsWithRateLimit tests that requests to gateway are rate-limited with a 429 response.
func proxyRequestsWithRateLimit(cluster framework.Cluster, gateway string, opts ...client.CollectResponsesOptsFn) {
	framework.Logf("expecting 429 response from %q", gateway)
	Eventually(func(g Gomega) {
		target := fmt.Sprintf("http://%s/%s",
			gateway, path.Join("test", url.PathEscape(GinkgoT().Name())),
		)

		opts = append(opts,
			client.NoFail(),
			client.OutputFormat(`{ "received": { "status": %{response_code} } }`),
			client.WithHeader("Host", "example.kuma.io"),
		)
		response, err := client.CollectResponse(cluster, "gateway-client", target, opts...)

		g.Expect(err).To(Succeed())
		g.Expect(response.Received.StatusCode).To(Equal(429))
	}, "30s", "1s").Should(Succeed())
}
