package delegated

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
	"github.com/kumahq/kuma/test/server/types"
)

func MeshProxyPatch(config *Config) func() {
	GinkgoHelper()

	return func() {
		// Disabled because of flakes: https://github.com/kumahq/kuma/issues/9348
		XIt("should add a header using Lua filter", func() {
			// given
			meshProxyPatch := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1 
kind: MeshProxyPatch
metadata:
  name: backend-lua-filter
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Mesh
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
`, config.CpNamespace, config.Mesh)

			// when
			err := kubernetes.Cluster.Install(framework.YamlK8s(meshProxyPatch))

			// then
			Expect(err).ToNot(HaveOccurred())
			Eventually(func() ([]types.EchoResponse, error) {
				return client.CollectResponses(
					kubernetes.Cluster,
					"demo-client",
					fmt.Sprintf("http://%s/test-server", config.KicIP),
					client.FromKubernetesPod(config.NamespaceOutsideMesh, "demo-client"),
				)
			}, "30s", "1s").Should(ContainElement(HaveField(
				`Received.Headers`,
				HaveKeyWithValue("X-Header", ContainElement("test")),
			)))
		})
	}
}
