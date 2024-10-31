package meshtcproute

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/sync/errgroup"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshtcproute/api/v1alpha1"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func Test() {
	meshName := "meshtcproute"

	BeforeAll(func() {
		// Global
		err := NewClusterSetup().
			Install(MTLSMeshUniversal(meshName)).
			Install(MeshTrafficPermissionAllowAllUniversal(meshName)).
			Setup(multizone.Global)
		Expect(err).ToNot(HaveOccurred())
		Expect(WaitForMesh(meshName, multizone.Zones())).To(Succeed())

		group := errgroup.Group{}
		NewClusterSetup().
			Install(Parallel(
				DemoClientUniversal(AppModeDemoClient, meshName,
					WithTransparentProxy(true),
				),
				TestServerUniversal("test-server-echo-1", meshName,
					WithArgs([]string{"echo", "--instance", "zone1"}),
					WithServiceVersion("v1"),
				),
			)).
			SetupInGroup(multizone.UniZone1, &group)

		NewClusterSetup().
			Install(Parallel(
				TestServerUniversal("test-server-echo-2", meshName,
					WithArgs([]string{"echo", "--instance", "zone2"}),
					WithServiceVersion("v2"),
				),
				TestServerUniversal("test-server-echo-3", meshName,
					WithArgs([]string{"echo", "--instance", "alias-zone2"}),
					WithServiceName("alias-test-server"),
					WithServiceVersion("v2"),
				),
			)).
			SetupInGroup(multizone.UniZone2, &group)

		Expect(group.Wait()).To(Succeed())

		Expect(DeleteMeshResources(
			multizone.Global,
			meshName,
			core_mesh.TrafficRouteResourceTypeDescriptor,
		)).To(Succeed())
	})

	E2EAfterEach(func() {
		Expect(DeleteMeshResources(
			multizone.Global,
			meshName,
			v1alpha1.MeshTCPRouteResourceTypeDescriptor,
		)).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(multizone.Global, meshName)
		DebugUniversal(multizone.UniZone1, meshName)
		DebugUniversal(multizone.UniZone2, meshName)
	})

	E2EAfterAll(func() {
		Expect(multizone.UniZone2.DeleteMeshApps(meshName)).To(Succeed())
		Expect(multizone.UniZone1.DeleteMeshApps(meshName)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(meshName)).To(Succeed())
	})

	It("should use MeshTCPRoute for cross-zone communication", func() {
		Expect(YamlUniversal(fmt.Sprintf(`
type: MeshTCPRoute
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
    - default:
        backendRefs:
        - kind: MeshServiceSubset
          name: alias-test-server
          weight: 100
          tags:
            kuma.io/zone: kuma-5
`, meshName))(multizone.Global)).To(Succeed())

		Eventually(func(g Gomega) {
			response, err := client.CollectResponsesByInstance(
				multizone.UniZone1,
				"demo-client",
				"test-server.mesh",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response).To(And(
				Not(HaveKey(MatchRegexp(`^zone1.*`))),
				Not(HaveKey(MatchRegexp(`^zone2.*`))),
				HaveKeyWithValue(
					MatchRegexp(`^alias-zone2.*`),
					Not(BeNil()),
				),
			))
		}, "30s", "500ms").Should(Succeed())
	})
}
