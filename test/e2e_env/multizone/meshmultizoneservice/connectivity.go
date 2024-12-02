package meshmultizoneservice

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"golang.org/x/sync/errgroup"

	"github.com/kumahq/kuma/test/e2e_env/kubernetes/gateway"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
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

		meshGateway := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshGateway
metadata:
  name: edge-gateway-mmzs
  labels:
    kuma.io/origin: zone
mesh: %s
spec:
  selectors:
  - match:
      kuma.io/service: edge-gateway-mmzs_%s_svc
  conf:
    listeners:
    - port: 8080
      protocol: HTTP
`, meshName, namespace)
		gatewayRoute := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
metadata:
  name: route-mzms
  namespace: %s
  labels:
    kuma.io/mesh: %s
    kuma.io/origin: zone
spec:
  targetRef:
    kind: MeshGateway
    name: edge-gateway-mmzs
  to:
    - targetRef:
        kind: Mesh
      rules:
        - matches:
            - path:
                type: PathPrefix
                value: /mmzs
          default:
            backendRefs:
              - kind: MeshMultiZoneService
                labels:
                  kuma.io/display-name: test-server
                port: 80
                weight: 1
`, Config.KumaNamespace, meshName)

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
			Install(YamlK8s(meshGateway)).
			Install(YamlK8s(gatewayRoute)).
			Install(YamlK8s(gateway.MkGatewayInstance("edge-gateway-mmzs", namespace, meshName))).
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

	type testCase struct {
		address       string
		instanceMatch types.GomegaMatcher
	}

	DescribeTable("Gateway in Kubernetes",
		func(given testCase) {
			Eventually(func(g Gomega) {
				response, err := client.CollectEchoResponse(
					multizone.KubeZone1, "demo-client",
					fmt.Sprintf("http://edge-gateway-mmzs.%s:8080/%s", namespace, given.address),
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)

				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.Instance).To(given.instanceMatch)
			}, "30s", "1s").Should(Succeed())
		},
		Entry("should access MeshMultiZoneService", testCase{
			address:       "mmzs",
			instanceMatch: Equal("kube-test-server-1"),
		}),
	)

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
