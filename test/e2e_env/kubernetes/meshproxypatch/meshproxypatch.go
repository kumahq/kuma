package meshproxypatch

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/kubernetes/env"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func MeshProxyPatch() {
	const meshName = "mesh-proxy-patch"
	const namespace = "mesh-proxy-patch"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshKubernetes(meshName)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(testserver.Install(
				testserver.WithName("test-client"),
				testserver.WithMesh(meshName),
				testserver.WithNamespace(namespace),
			)).
			Install(testserver.Install(
				testserver.WithName("test-server"),
				testserver.WithMesh(meshName),
				testserver.WithNamespace(namespace),
			)).
			Setup(env.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})
	E2EAfterAll(func() {
		Expect(env.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should add a header using Lua filter", func() {
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
    kind: MeshService
    name: test-client_mesh-proxy-patch_svc_80
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
`, Config.KumaNamespace, meshName)

		// when
		err := env.Cluster.Install(YamlK8s(meshProxyPatch))

		// then
		Expect(err).ToNot(HaveOccurred())
		Eventually(func(g Gomega) {
			responses, err := client.CollectResponses(env.Cluster, "test-client", "test-server_mesh-proxy-patch_svc_80.mesh", client.FromKubernetesPod(namespace, "test-client"))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses[0].Received.Headers["X-Header"]).To(ContainElements("test"))
		}, "30s", "1s").Should(Succeed())
	})
}
