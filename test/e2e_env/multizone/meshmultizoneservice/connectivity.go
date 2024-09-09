package meshmultizoneservice

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func Connectivity() {
	namespace := "mzmsconnectivity"
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
        k8s.kuma.io/namespace: mzmsconnectivity
  ports:
  - port: 80
    appProtocol: http
`)).
			Setup(multizone.Global)).To(Succeed())
		Expect(WaitForMesh(meshName, multizone.Zones())).To(Succeed())

		err := NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh(meshName),
				testserver.WithEchoArgs("echo", "--instance", "kube-test-server-1"),
			)).
			Install(democlient.Install(democlient.WithNamespace(namespace), democlient.WithMesh(meshName))).
			Setup(multizone.KubeZone1)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh(meshName),
				testserver.WithEchoArgs("echo", "--instance", "kube-test-server-2"),
			)).
			Setup(multizone.KubeZone2)
		Expect(err).ToNot(HaveOccurred())

		uniServiceYAML := `
type: MeshService
name: test-server
mesh: mzmsconnectivity
labels:
  kuma.io/origin: zone
  kuma.io/env: universal
  k8s.kuma.io/namespace: mzmsconnectivity # add a label to aggregate kube and uni service
  kuma.io/display-name: test-server # add a label to aggregate kube and uni service
spec:
  selector:
    dataplaneTags:
      kuma.io/service: test-server
  ports:
  - port: 80
    targetPort: 80
    appProtocol: http
`

		err = NewClusterSetup().
			Install(DemoClientUniversal("demo-client", meshName, WithTransparentProxy(true))).
			Install(TestServerUniversal("test-server", meshName, WithArgs([]string{"echo", "--instance", "uni-test-server"}))).
			Install(YamlUniversal(uniServiceYAML)).
			Setup(multizone.UniZone1)
		Expect(err).ToNot(HaveOccurred())
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
