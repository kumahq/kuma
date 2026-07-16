package meshhttproute

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/sync/errgroup"

	"github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	"github.com/kumahq/kuma/v3/pkg/test/resources/samples"
	. "github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/client"
	"github.com/kumahq/kuma/v3/test/framework/envs/multizone"
)

func Test() {
	_ = Describe("No Zone Egress", func() {
		test("meshhttproute", samples.MeshMTLSBuilder(), false)
	}, Ordered)

	_ = Describe("Zone Egress", func() {
		test("meshhttproute-ze", samples.MeshMTLSBuilder().WithEgressRoutingEnabled(), true)
	}, FlakeAttempts(3), Ordered)
}

func test(meshName string, meshBuilder *builders.MeshBuilder, withEgress bool) {
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
				DemoClientUniversal(AppModeDemoClient, meshName, WithTransparentProxy(true),
					WithLabels(map[string]string{"kuma.io/service": AppModeDemoClient})),
				TestServerUniversal(
					"dp-echo-1", meshName,
					WithArgs([]string{"echo", "--instance", "zone1-v1"}),
					WithServiceVersion("v1"),
				),
			)).
			SetupInGroup(multizone.UniZone1, &group)

		NewClusterSetup().
			Install(Parallel(
				TestServerUniversal(
					"dp-echo-2-v1", meshName,
					WithArgs([]string{"echo", "--instance", "zone2-v1"}),
					WithServiceVersion("v1"),
				),
				TestServerUniversal(
					"dp-echo-2-v2", meshName,
					WithArgs([]string{"echo", "--instance", "zone2-v2"}),
					WithServiceVersion("v2"),
				),
				TestServerUniversal(
					"dp-echo-2-v3", meshName,
					WithArgs([]string{"echo", "--instance", "zone2-v3"}),
					WithServiceVersion("v3"),
				),
				TestServerUniversal(
					"dp-echo-4", meshName,
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

	It("should use MeshHTTPRoute for cross-zone communication", func() {
		Expect(YamlUniversal(fmt.Sprintf(`
type: MeshHTTPRoute
name: route-alias-test-server
mesh: %s
spec:
  targetRef:
    kind: Dataplane
    labels:
      kuma.io/service: demo-client
  to:
    - targetRef:
        kind: MeshService
        labels:
          kuma.io/display-name: test-server
          kuma.io/zone: kuma-5
      rules:
        - matches:
          - path:
              value: /
              type: PathPrefix
          default:
            backendRefs:
              - kind: MeshService
                labels:
                  kuma.io/display-name: alias-test-server
                  kuma.io/zone: kuma-5
                port: 80
                weight: 100
`, meshName))(multizone.Global)).To(Succeed())

		Eventually(func(g Gomega) {
			response, err := client.CollectResponsesByInstance(multizone.UniZone1, "demo-client", "test-server.svc.kuma-5.mesh.local")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response).To(
				And(
					Not(HaveKey(MatchRegexp(`^zone1.*`))),
					Not(HaveKey(MatchRegexp(`^zone2.*`))),
					HaveKeyWithValue(MatchRegexp(`^alias-zone2.*`), Not(BeNil())),
				),
			)
		}, "30s", "1s").MustPassRepeatedly(3).Should(Succeed())

		response, err := client.CollectResponsesByInstance(multizone.UniZone1, "demo-client", "test-server.svc.kuma-5.mesh.local", client.WithNumberOfRequests(100))
		Expect(err).ToNot(HaveOccurred())
		Expect(response).To(
			And(
				Not(HaveKey(MatchRegexp(`^zone1.*`))),
				Not(HaveKey(MatchRegexp(`^zone2.*`))),
				HaveKeyWithValue(MatchRegexp(`^alias-zone2.*`), Not(BeNil())),
			),
		)
	})

	if withEgress {
		It("should use MeshHTTPRoute for cross-zone with a MeshMultiZoneService", func() {
			Expect(YamlUniversal(fmt.Sprintf(`
type: MeshMultiZoneService
name: test-server
mesh: %s
labels:
  kuma.io/origin: global
spec:
  selector:
    meshService:
      matchLabels:
        kuma.io/display-name: test-server
  ports:
  - port: 80
    appProtocol: http
`, meshName))(multizone.Global)).To(Succeed())

			Expect(YamlUniversal(fmt.Sprintf(`
type: MeshLoadBalancingStrategy
name: mzms-no-locality
mesh: %s
spec:
  to:
  - targetRef:
      kind: MeshMultiZoneService
      labels:
        kuma.io/display-name: test-server
      sectionName: '80'
    default:
      localityAwareness:
        disabled: true
`, meshName))(multizone.Global)).To(Succeed())

			Expect(YamlUniversal(fmt.Sprintf(`
type: MeshHTTPRoute
name: route-test-server-mzms
mesh: %s
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshMultiZoneService
        labels:
          kuma.io/display-name: test-server
      rules:
        - matches:
          - path:
              value: /
              type: PathPrefix
          default:
            backendRefs:
              - kind: MeshMultiZoneService
                labels:
                  kuma.io/display-name: test-server
                port: 80
                weight: 1
`, meshName))(multizone.Global)).To(Succeed())

			Eventually(func(g Gomega) {
				response, err := client.CollectResponsesByInstance(multizone.UniZone1, "demo-client", "test-server.mzsvc.mesh.local", client.WithNumberOfRequests(100))
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response).To(
					And(
						HaveKey(MatchRegexp(`^zone1-v1.*`)),
						HaveKey(MatchRegexp(`^zone2-.*`)),
					),
				)
			}, "60s", "5s").Should(Succeed())
		})
	}
}
