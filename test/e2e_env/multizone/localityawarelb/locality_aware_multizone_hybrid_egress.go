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

func LocalityAwareLBEgress() {
	const mesh = "locality-aware-lb-egress"
	const namespace = "locality-aware-lb-egress"

	meshLoadBalancingStrategyDemoClient := fmt.Sprintf(`
type: MeshLoadBalancingStrategy
name: mlbs-1
mesh: %s
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshService
        name: test-server_locality-aware-lb-egress_svc_80
      default:
        localityAwareness:
          crossZone:
            failover:
              - from: 
                  zones: ["kuma-4"]
                to:
                  type: Only
                  zones: ["kuma-1-zone"]
              - from: 
                  zones: ["kuma-4"]
                to:
                  type: Only
                  zones: ["kuma-5"]
              - from: 
                  zones: ["kuma-4"]
                to:
                  type: Any`, mesh)

	meshLoadBalancingStrategyDemoClientChangedPriority := fmt.Sprintf(`
type: MeshLoadBalancingStrategy
name: mlbs-1
mesh: %s
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshService
        name: test-server_locality-aware-lb-egress_svc_80
      default:
        localityAwareness:
          crossZone:
            failover:
              - from: 
                  zones: ["kuma-4"]
                to:
                  type: Only
                  zones: ["kuma-5"]
              - from: 
                  zones: ["kuma-4"]
                to:
                  type: Only
                  zones: ["kuma-1-zone"]
              - from: 
                  zones: ["kuma-4"]
                to:
                  type: Any`, mesh)

	BeforeAll(func() {
		// Global
		Expect(NewClusterSetup().
			Install(ResourceUniversal(samples.MeshMTLSBuilder().WithName(mesh).WithEgressRoutingEnabled().Build())).
			Setup(multizone.Global)).To(Succeed())
		Expect(WaitForMesh(mesh, multizone.Zones())).To(Succeed())

		// Universal Zone 4
		Expect(NewClusterSetup().
			Install(DemoClientUniversal(
				"demo-client_locality-aware-lb-egress_svc",
				mesh,
				WithTransparentProxy(true),
			)).
			Install(TestServerUniversal("test-server-zone-4", mesh,
				WithServiceName("test-server_locality-aware-lb-egress_svc_80"),
				WithArgs([]string{"echo", "--instance", "test-server-zone-4"}),
			)).
			Setup(multizone.UniZone1),
		).To(Succeed())

		// Universal Zone 5
		Expect(NewClusterSetup().
			Install(DemoClientUniversal(
				"demo-client_locality-aware-lb-egress_svc",
				mesh,
				WithTransparentProxy(true),
			)).
			Install(TestServerUniversal("test-server-zone-5", mesh,
				WithServiceName("test-server_locality-aware-lb-egress_svc_80"),
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
	})

	E2EAfterAll(func() {
		Expect(multizone.UniZone1.DeleteMeshApps(mesh)).To(Succeed())
		Expect(multizone.UniZone2.DeleteMeshApps(mesh)).To(Succeed())
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(mesh)).To(Succeed())
	})

	It("should route based on defined strategy with egress enabled", func() {
		// no lb priorities
		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(multizone.UniZone1, "demo-client_locality-aware-lb-egress_svc", "test-server_locality-aware-lb-egress_svc_80.mesh", client.WithNumberOfRequests(100))
		}, "1m", "10s").Should(
			And(
				HaveKeyWithValue(Equal(`test-server-zone-4`), BeNumerically("~", 66, 15)),
				HaveKeyWithValue(Equal(`test-server-zone-5`), BeNumerically("~", 17, 15)),
				HaveKeyWithValue(Equal(`test-server-zone-1`), BeNumerically("~", 17, 15)),
			),
		)

		// apply lb policy
		// kuma-4 - priority 0, kuma-1 - priority 1, kuma-5 - priority 2,
		Expect(multizone.Global.Install(YamlUniversal(meshLoadBalancingStrategyDemoClient))).To(Succeed())

		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(multizone.UniZone1, "demo-client_locality-aware-lb-egress_svc", "test-server_locality-aware-lb-egress_svc_80.mesh", client.WithNumberOfRequests(100))
		}, "1m", "10s").Should(
			HaveKeyWithValue(Equal(`test-server-zone-4`), BeNumerically("~", 100, 10)),
		)

		// kill test-server in kuma-4 zone
		Expect(multizone.UniZone1.DeleteApp("test-server-zone-4")).To(Succeed())

		// traffic goes to kuma-1 zone
		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(multizone.UniZone1, "demo-client_locality-aware-lb-egress_svc", "test-server_locality-aware-lb-egress_svc_80.mesh", client.WithNumberOfRequests(100))
		}, "1m", "10s").Should(
			HaveKeyWithValue(Equal(`test-server-zone-1`), BeNumerically("~", 100, 10)),
		)

		// apply lb policy with new priorities
		// kuma-4 - priority 0, kuma-5 - priority 1, kuma-1 - priority 2,
		Expect(multizone.Global.Install(YamlUniversal(meshLoadBalancingStrategyDemoClientChangedPriority))).To(Succeed())

		// traffic goes to kuma-5 zone
		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(multizone.UniZone1, "demo-client_locality-aware-lb-egress_svc", "test-server_locality-aware-lb-egress_svc_80.mesh", client.WithNumberOfRequests(100))
		}, "1m", "10s").Should(
			HaveKeyWithValue(Equal(`test-server-zone-5`), BeNumerically("~", 100, 10)),
		)

		// kill test-server from kuma-5 zone
		Expect(multizone.UniZone2.DeleteApp("test-server-zone-5")).To(Succeed())

		// then traffic should go to kuma-1 zone
		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(multizone.UniZone1, "demo-client_locality-aware-lb-egress_svc", "test-server_locality-aware-lb-egress_svc_80.mesh", client.WithNumberOfRequests(100))
		}, "1m", "10s").Should(
			HaveKeyWithValue(Equal(`test-server-zone-1`), BeNumerically("~", 100, 10)),
		)
	})
}
