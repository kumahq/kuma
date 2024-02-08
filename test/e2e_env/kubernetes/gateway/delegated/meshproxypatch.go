package delegated

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/plugins/policies/meshproxypatch/api/v1alpha1"
	"github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
	"github.com/kumahq/kuma/test/server/types"
)

func MeshProxyPatch(config *Config) func() {
	GinkgoHelper()

	return func() {
		framework.E2EAfterEach(func() {
			Expect(framework.DeleteMeshResources(
				kubernetes.Cluster,
				config.Mesh,
				v1alpha1.MeshProxyPatchResourceTypeDescriptor,
			)).To(Succeed())
		})

		It("should add a header using Lua filter", func() {
			// given
			meshProxyPatch := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1 
kind: MeshProxyPatch
metadata:
  name: backend-lua-filter-delegated
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
<<<<<<< HEAD:test/e2e_env/kubernetes/gateway/delegated/meshproxypatch.go
`, config.CpNamespace, config.Mesh)
=======
`, Config.KumaNamespace, config.mesh)
>>>>>>> 6cf0b3eea (test(e2e): upgrade KIC (#9157)):test/e2e_env/kubernetes/gateway/delegated_meshproxypatch.go

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
