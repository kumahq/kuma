package meshservice

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func MeshServiceTargeting() {
	meshName := "real-resource-mesh"
	namespace := "real-resource-ns"
	addressSuffix := "realreasource"
	addressToMeshService := func(service string) string {
		return fmt.Sprintf("%s.%s.%s.%s", service, namespace, Kuma1, addressSuffix)
	}

	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(MeshUniversal(meshName)).
			Setup(multizone.Global)).To(Succeed())
		Expect(WaitForMesh(meshName, multizone.Zones())).To(Succeed())

		err := NewClusterSetup().
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
			Install(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: HostnameGenerator
metadata:
  name: e2e-connectivity
  namespace: %s
  labels:
    kuma.io/origin: zone
spec:
  template: '{{ .DisplayName }}.{{ .Namespace }}.{{ .Zone }}.%s'
  selector:
    meshService:
      matchLabels:
        kuma.io/origin: zone
        kuma.io/managed-by: k8s-controller
`, Config.KumaNamespace, addressSuffix))).
			Setup(multizone.KubeZone1)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugUniversal(multizone.Global, meshName)
		DebugKube(multizone.KubeZone1, meshName)
	})

	E2EAfterEach(func() {
		Expect(DeleteMeshResources(multizone.KubeZone1, meshName, v1alpha1.MeshHTTPRouteResourceTypeDescriptor)).To(Succeed())
		Expect(DeleteMeshResources(multizone.KubeZone1, meshName, core_mesh.ExternalServiceResourceTypeDescriptor)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(meshName)).To(Succeed())
	})

	It("should configure URLRewrite", func() {
		// when
		Expect(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
metadata:
  name: route-3
  namespace: %s
  labels:
    kuma.io/mesh: %s
    kuma.io/origin: zone
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshService
        name: test-server
        sectionName: main
      rules: 
        - matches:
            - path: 
                type: PathPrefix
                value: /prefix
          default:
            filters:
              - type: URLRewrite
                urlRewrite:
                  path:
                    type: ReplacePrefixMatch
                    replacePrefixMatch: /hello/
`, namespace, meshName))(multizone.KubeZone1)).To(Succeed())
		// then receive redirect response
		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(multizone.KubeZone1, "test-client", fmt.Sprintf("%s/prefix/world", addressToMeshService("test-server")), client.FromKubernetesPod(namespace, "test-client"))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Received.Path).To(Equal("/hello/world"))
		}, "30s", "1s").Should(Succeed())
	})
}
