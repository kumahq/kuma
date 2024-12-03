package meshhttproute

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/sync/errgroup"

	"github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func Test() {
	_ = Describe("No Zone Egress", func() {
		test("meshhttproute", samples.MeshMTLSBuilder())
	}, Ordered)

	_ = Describe("Zone Egress", func() {
		test("meshhttproute-ze", samples.MeshMTLSBuilder().WithEgressRoutingEnabled())
	}, FlakeAttempts(3), Ordered)
}

func test(meshName string, meshBuilder *builders.MeshBuilder) {
	GinkgoHelper()

	BeforeAll(func() {
		// Global
		err := NewClusterSetup().
			Install(ResourceUniversal(meshBuilder.WithName(meshName).Build())).
			Install(MeshTrafficPermissionAllowAllUniversal(meshName)).
			Setup(multizone.Global)
		Expect(err).ToNot(HaveOccurred())
		Expect(WaitForMesh(meshName, multizone.Zones())).To(Succeed())

		group := errgroup.Group{}
		NewClusterSetup().
			Install(Parallel(
				DemoClientUniversal(AppModeDemoClient, meshName, WithTransparentProxy(true)),
				TestServerUniversal("dp-echo-1", meshName,
					WithArgs([]string{"echo", "--instance", "zone1-v1"}),
					WithServiceVersion("v1"),
				),
			)).
			SetupInGroup(multizone.UniZone1, &group)

		NewClusterSetup().
			Install(Parallel(
				TestServerUniversal("dp-echo-2-v1", meshName,
					WithArgs([]string{"echo", "--instance", "zone2-v1"}),
					WithServiceVersion("v1"),
				),
				TestServerUniversal("dp-echo-2-v2", meshName,
					WithArgs([]string{"echo", "--instance", "zone2-v2"}),
					WithServiceVersion("v2"),
				),
				TestServerUniversal("dp-echo-2-v3", meshName,
					WithArgs([]string{"echo", "--instance", "zone2-v3"}),
					WithServiceVersion("v3"),
				),
				TestServerUniversal("dp-echo-4", meshName,
					WithArgs([]string{"echo", "--instance", "alias-zone2"}),
					WithServiceName("alias-test-server"),
					WithServiceVersion("v2"),
				),
			)).
			SetupInGroup(multizone.UniZone2, &group)
		Expect(group.Wait()).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(multizone.Global, meshName)
		DebugUniversal(multizone.UniZone1, meshName)
		DebugUniversal(multizone.UniZone2, meshName)
	})

	E2EAfterEach(func() {
		Expect(DeleteMeshResources(multizone.Global, meshName, v1alpha1.MeshHTTPRouteResourceTypeDescriptor)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(multizone.UniZone2.DeleteMeshApps(meshName)).To(Succeed())
		Expect(multizone.UniZone1.DeleteMeshApps(meshName)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(meshName)).To(Succeed())
	})

	// Disabled because of flakes: https://github.com/kumahq/kuma/issues/9346
	XIt("should use MeshHTTPRoute for cross-zone communication", func() {
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

	// Disabled because of flakes: https://github.com/kumahq/kuma/issues/9346
	XIt("should use MeshHTTPRoute for cross-zone with MeshServiceSubset", func() {
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
                  kuma.io/zone: kuma-5
                  version: v1
              - kind: MeshServiceSubset
                name: test-server
                weight: 1
                tags:
                  kuma.io/zone: kuma-5
                  version: v2
`, meshName))(multizone.Global)).To(Succeed())

		Eventually(func(g Gomega) {
			response, err := client.CollectResponsesByInstance(multizone.UniZone1, "demo-client", "test-server.mesh")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response).To(
				And(
					HaveKey(MatchRegexp(`^.*-v1.*`)),
					HaveKey(MatchRegexp(`^zone2-v2.*`)),
					Not(HaveKey(MatchRegexp(`^zone2-v3.*`))),
				),
			)
		}, "30s", "500ms").Should(Succeed())
	})
}
