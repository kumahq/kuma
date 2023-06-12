package meshhttproute

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func Test() {
	meshName := "meshhttproute"

	BeforeAll(func() {
		// Global
		Expect(multizone.Global.Install(MTLSMeshUniversal(meshName))).To(Succeed())
		Expect(WaitForMesh(meshName, multizone.Zones())).To(Succeed())

		err := NewClusterSetup().
			Install(DemoClientUniversal(AppModeDemoClient, meshName, WithTransparentProxy(true))).
			Install(TestServerUniversal("dp-echo-1", meshName,
				WithArgs([]string{"echo", "--instance", "zone1"}),
				WithServiceVersion("v1"),
			)).
			Setup(multizone.UniZone1)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(TestServerUniversal("dp-echo-2", meshName,
				WithArgs([]string{"echo", "--instance", "zone2-v1"}),
				WithServiceVersion("v1"),
			)).
			Setup(multizone.UniZone2)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(TestServerUniversal("dp-echo-3", meshName,
				WithArgs([]string{"echo", "--instance", "zone2-v2"}),
				WithServiceVersion("v2"),
			)).
			Setup(multizone.UniZone2)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(TestServerUniversal("dp-echo-4", meshName,
				WithArgs([]string{"echo", "--instance", "alias-zone2"}),
				WithServiceName("alias-test-server"),
				WithServiceVersion("v2"),
			)).
			Setup(multizone.UniZone2)
		Expect(err).ToNot(HaveOccurred())

		Expect(DeleteMeshResources(multizone.Global, meshName, core_mesh.TrafficRouteResourceTypeDescriptor)).To(Succeed())
	})

	E2EAfterEach(func() {
		Expect(DeleteMeshResources(multizone.Global, meshName, v1alpha1.MeshHTTPRouteResourceTypeDescriptor)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(multizone.UniZone2.DeleteMeshApps(meshName)).To(Succeed())
		Expect(multizone.UniZone1.DeleteMeshApps(meshName)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(meshName)).To(Succeed())
	})

	It("should use MeshHTTPRoute for cross-zone communication", func() {
		Expect(YamlUniversal(fmt.Sprintf(`
type: MeshHTTPRoute
name: route-1
mesh: %s
spec:
  targetRef:
    kind: MeshService
    name: demo-client
  to:
    - targetRef:
        kind: MeshService
        name: test-server
      rules:
        - matches:
          - path:
              value: /
              type: PathPrefix
          default:
            backendRefs:
              - kind: MeshServiceSubset
                name: alias-test-server
                weight: 100
                tags:
                  kuma.io/zone: kuma-5
`, meshName))(multizone.Global)).To(Succeed())

		Eventually(func(g Gomega) {
			response, err := client.CollectResponsesByInstance(multizone.UniZone1, "demo-client", "test-server.mesh")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response).To(
				And(
					Not(HaveKey(MatchRegexp(`^zone1.*`))),
					Not(HaveKey(MatchRegexp(`^zone2.*`))),
					HaveKeyWithValue(MatchRegexp(`^alias-zone2.*`), Not(BeNil())),
				),
			)
		}, "30s", "500ms").Should(Succeed())
	})

	It("should use MeshHTTPRoute for cross-zone with MeshServiceSubset", func() {
		Expect(YamlUniversal(fmt.Sprintf(`
type: MeshHTTPRoute
name: route-1
mesh: %s
spec:
  targetRef:
    kind: MeshService
    name: demo-client
  to:
    - targetRef:
        kind: MeshService
        name: test-server
      rules:
        - matches:
          - path:
              value: /
              type: PathPrefix
          default:
            backendRefs:
              - kind: MeshServiceSubset
                name: test-server
                weight: 1
                tags:
                  version: v1
              - kind: MeshServiceSubset
                name: test-server
                weight: 1
                tags:
                  version: v2
`, meshName))(multizone.Global)).To(Succeed())

		Eventually(func(g Gomega) {
			response, err := client.CollectResponsesByInstance(multizone.UniZone1, "demo-client", "test-server.mesh")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response).To(
				And(
					HaveKey(MatchRegexp(`^zone2-v1.*`)),
					HaveKey(MatchRegexp(`^zone2-v2.*`)),
				),
			)
		}, "30s", "500ms").Should(Succeed())
	})
}
