package localityawarelb

import (
	core_meta "github.com/kumahq/kuma/v2/pkg/core/metadata"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/sync/errgroup"

	. "github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/client"
	"github.com/kumahq/kuma/v2/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/v2/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/v2/test/framework/envs/multizone"
)

func MeshMzService() {
	namespace := "mlb-mzms"
	meshName := "mlb-mzms"

	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(MTLSMeshWithMeshServicesUniversal(meshName, "Everywhere")).
			Install(MeshTrafficPermissionAllowAllUniversal(meshName)).
			Install(ResourceUniversal(builders.MeshMultiZoneService().
				WithName("test-server").
				WithMesh(meshName).
				WithLabels(map[string]string{"test-name": "mzmsconnectivity"}).
				WithServiceLabelSelector(map[string]string{"kuma.io/display-name": "test-server"}).
				AddIntPortWithName(80, core_meta.ProtocolHTTP, "80").
				Build())).Setup(multizone.Global)).To(Succeed())
		Expect(WaitForMesh(meshName, multizone.Zones())).To(Succeed())

		group := errgroup.Group{}

		NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh(meshName),
				testserver.WithEchoArgs("echo", "--instance", "kube-test-server-1"),
			)).
			Install(democlient.Install(democlient.WithNamespace(namespace), democlient.WithMesh(meshName))).
			SetupInGroup(multizone.KubeZone1, &group)

		NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(democlient.Install(democlient.WithNamespace(namespace), democlient.WithMesh(meshName))).
			SetupInGroup(multizone.KubeZone2, &group)

		NewClusterSetup().
			Install(TestServerUniversal("test-server", meshName, WithArgs([]string{"echo", "--instance", "uni-test-server"}))).
			SetupInGroup(multizone.UniZone1, &group)
		Expect(group.Wait()).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(multizone.Global, meshName)
		DebugUniversal(multizone.UniZone1, meshName)
		DebugUniversal(multizone.UniZone2, meshName)
		DebugKube(multizone.KubeZone1, meshName, namespace)
		DebugKube(multizone.KubeZone2, meshName, namespace)
	})

	E2EAfterAll(func() {
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.KubeZone2.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.UniZone1.DeleteMeshApps(meshName)).To(Succeed())
		Expect(multizone.UniZone2.DeleteMeshApps(meshName)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(meshName)).To(Succeed())
	})

	responseFromInstance := func(cluster Cluster) func() (string, error) {
		return func() (string, error) {
			var opts []client.CollectResponsesOptsFn
			if _, ok := cluster.(*K8sCluster); ok {
				opts = append(opts, client.FromKubernetesPod(meshName, "demo-client"))
			}
			response, err := client.CollectEchoResponse(cluster, "demo-client", "http://test-server.mzsvc.mesh.local:80", opts...)
			if err != nil {
				return "", err
			}
			return response.Instance, nil
		}
	}

	It("should fallback only to first zone", func() {
		// given traffic to other zones
		Eventually(responseFromInstance(multizone.KubeZone2), "30s", "1s").
			Should(Equal("kube-test-server-1"))
		Eventually(responseFromInstance(multizone.KubeZone2), "30s", "1s").
			Should(Equal("uni-test-server"))

		// when
		policy := `
type: MeshLoadBalancingStrategy
name: mlb-mzms
mesh: mlb-mzms
spec:
  targetRef:
    kind: Mesh
  to:
  - targetRef:
      kind: MeshMultiZoneService
      labels:
        kuma.io/display-name: test-server
    default:
      localityAwareness:
        crossZone:
          failover:
          - to:
              type: Only
              zones:
              - "kuma-1"
          - to:
              type: Any
          failoverThreshold:
            percentage: 100
`
		err := multizone.Global.Install(YamlUniversal(policy))

		// then
		Expect(err).ToNot(HaveOccurred())

		Eventually(responseFromInstance(multizone.KubeZone2), "30s", "1s").
			MustPassRepeatedly(5).Should(Equal("kube-test-server-1"))
	})

	It("should be locality aware unless disabled", func() {
		// given traffic only to the local zone
		Eventually(responseFromInstance(multizone.KubeZone1), "30s", "1s").
			MustPassRepeatedly(5).Should(Equal("kube-test-server-1"))

		// when
		policy := `
type: MeshLoadBalancingStrategy
name: mlb-mzms
mesh: mlb-mzms
spec:
  targetRef:
    kind: Mesh
  to:
  - targetRef:
      kind: MeshMultiZoneService
      labels:
        kuma.io/display-name: test-server
    default:
      localityAwareness:
        disabled: true
`
		err := multizone.Global.Install(YamlUniversal(policy))

		// then
		Expect(err).ToNot(HaveOccurred())

		Eventually(responseFromInstance(multizone.KubeZone1), "30s", "1s").
			Should(Equal("uni-test-server"))
	})
}
