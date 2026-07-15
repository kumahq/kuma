package meshmultizoneservice

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/sync/errgroup"

	. "github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/client"
	"github.com/kumahq/kuma/v3/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/v3/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/v3/test/framework/envs/multizone"
)

func Connectivity() {
	namespace := "mzmsconnectivity"
	clientNamespace := "mzmsconnectivity-client"
	meshName := "mzmsconnectivity"

	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(MTLSMeshWithMeshServicesUniversal(meshName, "Everywhere")).
			Install(MeshTrafficPermissionAllowAllUniversal(meshName)).
			Install(YamlUniversal(`
type: MeshMultiZoneService
name: test-server
mesh: mzmsconnectivity
labels:
  test-name: mzmsconnectivity
spec:
  selector:
    meshService:
      matchLabels:
        kuma.io/display-name: test-server
  ports:
  - port: 80
    appProtocol: http
`)).
			Setup(multizone.Global)).To(Succeed())
		Expect(WaitForMesh(meshName, multizone.Zones())).To(Succeed())

		group := errgroup.Group{}
		NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Namespace(clientNamespace)).
			Install(Parallel(
				testserver.Install(
					testserver.WithNamespace(namespace),
					testserver.WithMesh(meshName),
					testserver.WithEchoArgs("echo", "--instance", "kube-test-server-1"),
				),
				testserver.Install(
					testserver.WithName("demo-client"),
					testserver.WithNamespace(clientNamespace),
				),
				democlient.Install(democlient.WithNamespace(namespace), democlient.WithMesh(meshName)),
			)).
			SetupInGroup(multizone.KubeZone1, &group)

		NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh(meshName),
				testserver.WithEchoArgs("echo", "--instance", "kube-test-server-2"),
			)).
			SetupInGroup(multizone.KubeZone2, &group)

		NewClusterSetup().
			Install(Parallel(
				DemoClientUniversal("demo-client", meshName, WithTransparentProxy(true)),
				TestServerUniversal("test-server", meshName, WithArgs([]string{"echo", "--instance", "uni-test-server"})),
			)).
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
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(clientNamespace)).To(Succeed())
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

	It("should access app from kube cluster and fallback to other zones", func() {
		// given traffic to local zone only
		Eventually(responseFromInstance(multizone.KubeZone1), "30s", "1s").
			MustPassRepeatedly(5).Should(Equal("kube-test-server-1"))

		// when
		err := ScaleApp(multizone.KubeZone1, "test-server", namespace, 0)

		// then
		Expect(err).ToNot(HaveOccurred())
		Eventually(responseFromInstance(multizone.KubeZone1), "30s", "1s").
			MustPassRepeatedly(5).Should(Or(Equal("kube-test-server-2"), Equal("uni-test-server")))
	})

	It("should access app from uni cluster and fallback to other zones", func() {
		// given traffic to local zone only
		Eventually(responseFromInstance(multizone.UniZone1), "30s", "1s").
			MustPassRepeatedly(5).Should(Equal("uni-test-server"))

		// when
		err := multizone.UniZone1.DeleteApp("test-server")

		// then
		Expect(err).ToNot(HaveOccurred())
		Eventually(responseFromInstance(multizone.UniZone1), "30s", "1s").
			MustPassRepeatedly(5).Should(Equal("kube-test-server-2"))
	})
}
