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
	"github.com/kumahq/kuma/test/framework/envoy_admin"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func LocalityAwareLBEgress() {
	const mesh = "locality-aware-lb-egress"
	const namespace = "locality-aware-lb-egress"
	var egressTunnel envoy_admin.Tunnel

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

		// TODO: should be removed after fixing egress issues
		// Workaround to wait for all remote zones to set up
		egressTunnel = multizone.UniZone1.GetZoneEgressEnvoyTunnel()
		Eventually(func(g Gomega) int {
			egressClusters, err := egressTunnel.GetClusters()
			g.Expect(err).ToNot(HaveOccurred())
			cluster := egressClusters.GetCluster(fmt.Sprintf("%s:test-server_locality-aware-lb-egress_svc_80", mesh))
			g.Expect(cluster).ToNot(BeNil())
			return len(cluster.HostStatuses)
		}, "2m", "5s").Should(Equal(2))

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
			return client.CollectResponsesByInstance(multizone.UniZone1, "demo-client_locality-aware-lb-egress_svc", "test-server_locality-aware-lb-egress_svc_80.mesh", client.WithNumberOfRequests(300))
		}, "2m", "10s").Should(
			And(
				HaveKeyWithValue(Equal(`test-server-zone-4`), BeNumerically("~", 200, 40)),
				HaveKeyWithValue(Equal(`test-server-zone-5`), BeNumerically("~", 50, 40)),
				HaveKeyWithValue(Equal(`test-server-zone-1`), BeNumerically("~", 50, 40)),
			),
		)

		// apply lb policy
		// kuma-4 - priority 0, kuma-1 - priority 1, kuma-5 - priority 2,
		Expect(multizone.Global.Install(YamlUniversal(meshLoadBalancingStrategyDemoClient))).To(Succeed())

		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(multizone.UniZone1, "demo-client_locality-aware-lb-egress_svc", "test-server_locality-aware-lb-egress_svc_80.mesh", client.WithNumberOfRequests(50))
		}, "1m", "10s").Should(
			HaveKeyWithValue(Equal(`test-server-zone-4`), BeNumerically("~", 50, 25)),
		)

		// kill test-server in kuma-4 zone
		Expect(multizone.UniZone1.DeleteApp("test-server-zone-4")).To(Succeed())

		// traffic goes to kuma-1 zone
		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(multizone.UniZone1, "demo-client_locality-aware-lb-egress_svc", "test-server_locality-aware-lb-egress_svc_80.mesh", client.WithNumberOfRequests(50))
		}, "1m", "10s").Should(
			HaveKeyWithValue(Equal(`test-server-zone-1`), BeNumerically("~", 50, 25)),
		)

		// apply lb policy with new priorities
		// kuma-4 - priority 0, kuma-5 - priority 1, kuma-1 - priority 2,
		Expect(multizone.Global.Install(YamlUniversal(meshLoadBalancingStrategyDemoClientChangedPriority))).To(Succeed())

		// TODO: should be removed after fixing egress issues
		// egress config refresh workaround
		Eventually(func(g Gomega) int {
			egressClusters, err := egressTunnel.GetClusters()
			g.Expect(err).ToNot(HaveOccurred())
			cluster := egressClusters.GetCluster(fmt.Sprintf("%s:test-server_locality-aware-lb-egress_svc_80", mesh))
			g.Expect(cluster).ToNot(BeNil())
			return cluster.GetPriorityForZone("kuma-5")
		}, "2m", "5s").Should(Equal(1))
		Expect(multizone.UniZone1.GetApp("demo-client_locality-aware-lb-egress_svc").ReStart()).To(Succeed())

		// traffic goes to kuma-5 zone
		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(multizone.UniZone1, "demo-client_locality-aware-lb-egress_svc", "test-server_locality-aware-lb-egress_svc_80.mesh", client.WithNumberOfRequests(50))
		}, "1m", "10s").Should(
			HaveKeyWithValue(Equal(`test-server-zone-5`), BeNumerically("~", 50, 25)),
		)

		// kill test-server from kuma-5 zone
		Expect(multizone.UniZone2.DeleteApp("test-server-zone-5")).To(Succeed())

		// then traffic should go to kuma-1 zone
		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(multizone.UniZone1, "demo-client_locality-aware-lb-egress_svc", "test-server_locality-aware-lb-egress_svc_80.mesh", client.WithNumberOfRequests(50))
		}, "1m", "10s").Should(
			HaveKeyWithValue(Equal(`test-server-zone-1`), BeNumerically("~", 50, 25)),
		)
	})
}
