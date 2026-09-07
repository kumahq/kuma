package meshproxypatch

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/client"
	"github.com/kumahq/kuma/v3/test/framework/envoy_admin/stats"
	"github.com/kumahq/kuma/v3/test/framework/envs/universal"
)

func MeshProxyPatch() {
	const mesh = "mesh-proxy-patch"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshUniversal(mesh)).
			Install(TestServerUniversal(
				"test-server", mesh,
				WithTransparentProxy(true),
				WithArgs([]string{"echo", "--instance", "echo-v1"}),
				WithServiceName("test-server"),
			)).
			Install(DemoClientUniversal(AppModeDemoClient, mesh, WithTransparentProxy(true), WithLabels(map[string]string{"kuma.io/display-name": AppModeDemoClient}))).
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, mesh)
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(mesh)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(mesh)).To(Succeed())
	})

	It("should add a header using Lua filter", func() {
		// given
		proxyTemplate := fmt.Sprintf(`
type: MeshProxyPatch
mesh: %s
name: backend-lua-filter
spec:
  targetRef:
    kind: Dataplane
    labels:
      kuma.io/display-name: demo-client
  default:
    appendModifications:
      - httpFilter:
          operation: AddBefore
          match:
            name: envoy.filters.http.router
            origin: outbound
          value: |
            name: envoy.filters.http.lua
            typedConfig:
              '@type': type.googleapis.com/envoy.extensions.filters.http.lua.v3.Lua
              inline_code: |
                function envoy_on_request(request_handle)
                  request_handle:headers():add("X-Header", "test")
                end
`, mesh)

		// when
		err := universal.Cluster.Install(YamlUniversal(proxyTemplate))

		// then
		Expect(err).ToNot(HaveOccurred())
		Eventually(func(g Gomega) {
			responses, err := client.CollectResponses(universal.Cluster, "demo-client", "test-server.svc.mesh.local")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses[0].Received.Headers["X-Header"]).To(ContainElements("test"))
		}, "30s", "1s").Should(Succeed())
	})

	It("should apply a circuit breaker threshold patched onto the generated one", func() {
		// given a cluster that already carries a DEFAULT-priority threshold
		policy := fmt.Sprintf(`
type: MeshProxyPatch
mesh: %s
name: circuit-breaker-thresholds
spec:
  targetRef:
    kind: Dataplane
    labels:
      kuma.io/display-name: demo-client
  default:
    appendModifications:
      - cluster:
          operation: Patch
          match:
            name: self_transparentproxy_passthrough_outbound_ipv4
          value: |
            circuitBreakers:
              thresholds:
                - maxConnections: 8192
`, mesh)

		// when
		Expect(universal.Cluster.Install(YamlUniversal(policy))).To(Succeed())

		// then the limit Envoy resolved changes, not just the one /config_dump reports
		admin := universal.Cluster.GetApp(AppModeDemoClient).GetEnvoyAdminTunnel()
		Eventually(func(g Gomega) {
			// anchored: the filter is a regex, and an unanchored remaining_cx
			// also matches remaining_cx_pools
			s, err := admin.GetStats("cluster.self_transparentproxy_passthrough_outbound_ipv4.circuit_breakers.default.remaining_cx$")
			g.Expect(err).ToNot(HaveOccurred())
			// remaining_cx counts down as the cluster opens connections, so
			// assert against Envoy's 1024 default rather than an exact 8192
			g.Expect(s).To(stats.BeGreaterThan(float64(1024)))
		}, "30s", "1s").Should(Succeed())
	})
}
