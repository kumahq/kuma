package meshproxypatch

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/universal/env"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
)

func MeshProxyPatch() {
	const mesh = "mesh-proxy-patch"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshUniversal(mesh)).
			Install(TestServerUniversal("test-server", mesh,
				WithTransparentProxy(true),
				WithArgs([]string{"echo", "--instance", "echo-v1"}),
				WithServiceName("test-server"),
			)).
			Install(DemoClientUniversal(AppModeDemoClient, mesh, WithTransparentProxy(true))).
			Setup(env.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})
	E2EAfterAll(func() {
		Expect(env.Cluster.DeleteMeshApps(mesh)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(mesh)).To(Succeed())
	})

	It("should add a header using Lua filter", func() {
		// given
		proxyTemplate := fmt.Sprintf(`
type: MeshProxyPatch
mesh: %s
name: backend-lua-filter
spec:
  targetRef:
    kind: MeshService
    name: demo-client
  default:
    appendModifications:
      - httpFilter:
          operation: AddBefore
          match:
            name: envoy.filters.http.router
            origin: outbound
            listenerTags:
              kuma.io/service: test-server
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
		err := env.Cluster.Install(YamlUniversal(proxyTemplate))

		// then
		Expect(err).ToNot(HaveOccurred())
		Eventually(func(g Gomega) {
			responses, err := client.CollectResponses(env.Cluster, "demo-client", "test-server.mesh")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses[0].Received.Headers["X-Header"]).To(ContainElements("test"))
		}, "30s", "1s").Should(Succeed())
	})
}
