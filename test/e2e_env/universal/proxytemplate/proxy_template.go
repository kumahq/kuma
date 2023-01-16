package proxytemplate

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func ProxyTemplate() {
	const mesh = "proxy-template"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshUniversal(mesh)).
			Install(TestServerUniversal("test-server", mesh,
				WithTransparentProxy(true),
				WithArgs([]string{"echo", "--instance", "echo-v1"}),
				WithServiceName("test-server"),
			)).
			Install(DemoClientUniversal(AppModeDemoClient, mesh, WithTransparentProxy(true))).
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})
	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(mesh)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(mesh)).To(Succeed())
	})

	It("should add a header using Lua filter", func() {
		// given
		proxyTemplate := fmt.Sprintf(`
type: ProxyTemplate
mesh: %s
name: backend-lua-filter
selectors:
  - match:
      kuma.io/service: demo-client
conf:
  imports:
    - default-proxy # apply modifications on top of resources generated by Kuma
  modifications:
    - httpFilter:
        operation: addBefore
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
		err := YamlUniversal(proxyTemplate)(universal.Cluster)

		// then
		Expect(err).ToNot(HaveOccurred())
		Eventually(func(g Gomega) {
			responses, err := client.CollectResponses(universal.Cluster, "demo-client", "test-server.mesh")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses[0].Received.Headers["X-Header"]).To(ContainElements("test"))
		}, "30s", "1s").Should(Succeed())
	})
}
