package localityawarelb

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/test/resources/samples"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)
func LocalityAwareLB() {
	const mesh = "locality-aware-lb"
	const namespace = "locality-aware-lb"

	meshLoadBalancingStrategyDemoClient := fmt.Sprintf(`
type: MeshLoadBalancingStrategy
name: mlbs-1
mesh: %s
spec:
  targetRef:
    kind: MeshService
    name: demo-client_locality-aware-lb_svc
  to:
    - targetRef:
        kind: MeshService
        name: test-server_locality-aware-lb_svc_80
      default:
        localityAwareness:
          localZone:
            affinityTags:
            - key: k8s.io/node
              weight: 900
            - key: k8s.io/az
              weight: 90
          crossZone:
            failover:
              - from:
                  zones: ["kuma-4", "kuma-5"]
                to:
                  type: Only
                  zones: ["kuma-4", "kuma-5"]
              - from:
                  zones: ["kuma-1-zone", "kuma-2-zone"]
                to:
                  type: Only
                  zones: ["kuma-1-zone", "kuma-2-zone"]
              - from:
                  zones: ["kuma-4"]
                to:
                  type: Only
                  zones: ["kuma-1-zone"]`, mesh)

	BeforeAll(func() {
		// Global
		Expect(NewClusterSetup().
			Install(ResourceUniversal(samples.MeshMTLSBuilder().WithName(mesh).Build())).
			Install(YamlUniversal(meshLoadBalancingStrategyDemoClient)).
			Setup(multizone.Global)).To(Succeed())
		Expect(WaitForMesh(mesh, multizone.Zones())).To(Succeed())

		// Universal Zone 4
		Expect(NewClusterSetup().
			Install(DemoClientUniversal(
				"demo-client_locality-aware-lb_svc",
				mesh,
				WithTransparentProxy(true),
				WithAdditionalTags(
					map[string]string{
						"k8s.io/node": "node-1",
						"k8s.io/az":   "az-1",
					}),
			)).
			Install(TestServerUniversal("test-server-node-1", mesh, WithAdditionalTags(
				map[string]string{
					"k8s.io/node": "node-1",
					"k8s.io/az":   "az-1",
				}),
				WithServiceName("test-server_locality-aware-lb_svc_80"),
				WithArgs([]string{"echo", "--instance", "test-server-node-1-zone-4"}),
			)).
			Install(TestServerUniversal("test-server-az-1", mesh, WithAdditionalTags(
				map[string]string{
					"k8s.io/az": "az-1",
				}),
				WithServiceName("test-server_locality-aware-lb_svc_80"),
				WithArgs([]string{"echo", "--instance", "test-server-az-1-zone-4"}),
			)).
			Install(TestServerUniversal("test-server-node-2", mesh, WithAdditionalTags(
				map[string]string{
					"k8s.io/node": "node-2",
				}),
				WithServiceName("test-server_locality-aware-lb_svc_80"),
				WithArgs([]string{"echo", "--instance", "test-server-node-2-zone-4"}),
			)).
			Install(TestServerUniversal("test-server-no-tags", mesh,
				WithArgs([]string{"echo", "--instance", "test-server-no-tags-zone-4"}),
				WithServiceName("test-server_locality-aware-lb_svc_80"),
			)).
			Setup(multizone.UniZone1),
		).To(Succeed())

		// Universal Zone 5
		Expect(NewClusterSetup().
			Install(DemoClientUniversal(
				"demo-client_locality-aware-lb_svc",
				mesh,
				WithTransparentProxy(true),
			)).
			Install(TestServerUniversal("test-server", mesh,
				WithServiceName("test-server_locality-aware-lb_svc_80"),
				WithArgs([]string{"echo", "--instance", "test-server-zone-5"}),
			)).
			Setup(multizone.UniZone2),
		).To(Succeed())

		// Kubernetes Zone 1
		Expect(NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(democlient.Install(democlient.WithMesh(mesh), democlient.WithNamespace(namespace))).
			Install(testserver.Install(
				testserver.WithName("test-server"),
				testserver.WithMesh(mesh),
				testserver.WithNamespace(namespace),
				testserver.WithEchoArgs("echo", "--instance", "test-server-zone-1"),
			)).
			Setup(multizone.KubeZone1)).ToNot(HaveOccurred())

		// Kubernetes Zone 2
		Expect(NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(democlient.Install(democlient.WithMesh(mesh), democlient.WithNamespace(namespace))).
			Install(testserver.Install(
				testserver.WithName("test-server"),
				testserver.WithMesh(mesh),
				testserver.WithNamespace(namespace),
				testserver.WithEchoArgs("echo", "--instance", "test-server-zone-2"),
			)).
			Setup(multizone.KubeZone2)).ToNot(HaveOccurred())
	})

	E2EAfterAll(func() {
		Expect(multizone.UniZone1.DeleteMeshApps(mesh)).To(Succeed())
		Expect(multizone.UniZone2.DeleteMeshApps(mesh)).To(Succeed())
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.KubeZone2.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(mesh)).To(Succeed())

	})

	It("should route based on defined strategy", func() {
		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(multizone.UniZone1, "demo-client_locality-aware-lb_svc", "test-server_locality-aware-lb_svc_80.mesh", client.WithNumberOfRequests(100))
		}, "30s", "500ms").Should(
			And(
				HaveKeyWithValue(Equal(`test-server-node-1-zone-4`), BeNumerically("~", 93, 5)),
				HaveKeyWithValue(Equal(`test-server-az-1-zone-4`), BeNumerically("~", 7, 5)),
				Or(
					HaveKeyWithValue(Equal(`test-server-node-2-zone-4`), BeNumerically("~", 2, 2)),
					HaveKeyWithValue(Equal(`test-server-no-tags-zone-4`), BeNumerically("~", 2, 2)),
					Not(HaveKeyWithValue(Equal(`test-server-node-2-zone-4`), BeNumerically("~", 2, 2))),
					Not(HaveKeyWithValue(Equal(`test-server-no-tags-zone-4`), BeNumerically("~", 2, 2))),
				),
			),
		)

		// when app with the highest weight is down
		Expect(multizone.UniZone1.DeleteApp("test-server-node-1")).To(Succeed())

		// then traffic goes to the next highest
		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(multizone.UniZone1, "demo-client_locality-aware-lb_svc", "test-server_locality-aware-lb_svc_80.mesh", client.WithNumberOfRequests(100))
		}, "30s", "500ms").Should(
			And(
				HaveKeyWithValue(Equal(`test-server-az-1-zone-4`), BeNumerically("~", 96, 5)),
				Or(
					HaveKeyWithValue(Equal(`test-server-node-2-zone-4`), BeNumerically("~", 2, 2)),
					HaveKeyWithValue(Equal(`test-server-no-tags-zone-4`), BeNumerically("~", 2, 2)),
				),
			),
		)

		// when the next app with the highest weight is down
		Expect(multizone.UniZone1.DeleteApp("test-server-az-1")).To(Succeed())

		// then traffic goes to the next highest
		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(multizone.UniZone1, "demo-client_locality-aware-lb_svc", "test-server_locality-aware-lb_svc_80.mesh", client.WithNumberOfRequests(100))
		}, "30s", "500ms").Should(
			And(
				HaveKeyWithValue(Equal(`test-server-node-2-zone-4`), BeNumerically("~", 50, 5)),
				HaveKeyWithValue(Equal(`test-server-no-tags-zone-4`), BeNumerically("~", 50, 5)),
			),
		)

		// when all apps in local zone down
		Expect(multizone.UniZone1.DeleteApp("test-server-node-2")).To(Succeed())
		Expect(multizone.UniZone1.DeleteApp("test-server-no-tags")).To(Succeed())

		// then traffic goes to the zone with the next priority, kuma-5
		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(multizone.UniZone1, "demo-client_locality-aware-lb_svc", "test-server_locality-aware-lb_svc_80.mesh", client.WithNumberOfRequests(100))
		}, "30s", "500ms").Should(
			And(
				HaveKeyWithValue(Equal(`test-server-zone-5`), BeNumerically("~", 100, 5)),
			),
		)

		// when zone kuma-5 is disabled
		Expect(YamlUniversal(`
name: kuma-5
type: Zone
enabled: false
`)(multizone.Global)).To(Succeed())

		// then traffic should goes to k8s
		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(multizone.UniZone1, "demo-client_locality-aware-lb_svc", "test-server_locality-aware-lb_svc_80.mesh", client.WithNumberOfRequests(100))
		}, "30s", "500ms").Should(
			And(
				HaveKeyWithValue(Equal(`test-server-zone-1`), BeNumerically("~", 100, 5)),
			),
		)

		// when zone kuma-1-zone is disabled
		Expect(YamlUniversal(`
name: kuma-1-zone
type: Zone
enabled: false
`)(multizone.Global)).To(Succeed())

		// then traffic should goes to k8s
		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				multizone.UniZone1, "demo-client_locality-aware-lb_svc", "test-server_locality-aware-lb_svc_80.mesh",
			)
			g.Expect(err).To(HaveOccurred())
		}).Should(Succeed())
	})
}
