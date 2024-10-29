package meshhttproute

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
	"github.com/kumahq/kuma/test/server/types"
)

func MeshService() {
	meshName := "meshhttproutems"
	namespace := "meshhttproutems"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MTLSMeshWithMeshServicesUniversal(meshName, "Exclusive")).
			Install(MeshTrafficPermissionAllowAllUniversal(meshName)).
			Install(YamlUniversal(`
type: MeshMultiZoneService
name: test-server
mesh: meshhttproutems
labels:
  test-name: meshhttproutems
spec:
  selector:
    meshService:
      matchLabels:
        kuma.io/display-name: test-server
        k8s.kuma.io/namespace: meshhttproutems
        kuma.io/zone: kuma-2 # pick specific zone so we do not rely on default fallback
  ports:
  - port: 80
    appProtocol: http
`)).
			Setup(multizone.Global)
		Expect(err).ToNot(HaveOccurred())
		Expect(WaitForMesh(meshName, multizone.Zones())).To(Succeed())

		group := errgroup.Group{}
		group.Go(func() error {
			err = NewClusterSetup().
				Install(NamespaceWithSidecarInjection(namespace)).
				Install(Parallel(
					testserver.Install(
						testserver.WithNamespace(namespace),
						testserver.WithMesh(meshName),
						testserver.WithEchoArgs("echo", "--instance", "kube-test-server-1"),
					),
					democlient.Install(democlient.WithNamespace(namespace), democlient.WithMesh(meshName)),
				)).
				Setup(multizone.KubeZone1)
			return errors.Wrap(err, multizone.KubeZone1.Name())
		})

		group.Go(func() error {
			err = NewClusterSetup().
				Install(NamespaceWithSidecarInjection(namespace)).
				Install(testserver.Install(
					testserver.WithNamespace(namespace),
					testserver.WithMesh(meshName),
					testserver.WithEchoArgs("echo", "--instance", "kube-test-server-2"),
				)).
				Setup(multizone.KubeZone2)
			return errors.Wrap(err, multizone.KubeZone2.Name())
		})
		Expect(group.Wait()).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(multizone.Global, meshName)
		DebugKube(multizone.KubeZone1, meshName, namespace)
		DebugKube(multizone.KubeZone2, meshName, namespace)
		DebugUniversal(multizone.UniZone1, meshName)
		DebugUniversal(multizone.UniZone2, meshName)
	})

	E2EAfterAll(func() {
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.KubeZone2.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.UniZone2.DeleteMeshApps(meshName)).To(Succeed())
		Expect(multizone.UniZone1.DeleteMeshApps(meshName)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(meshName)).To(Succeed())
	})

	It("should use MeshHTTPRoute for cross-zone communication", func() {
		Expect(YamlUniversal(fmt.Sprintf(`
type: MeshHTTPRoute
name: meshhttproutems-route-1
mesh: %s
spec:
  targetRef:
    kind: MeshSubset
    tags:
      app: demo-client
  to:
    - targetRef:
        kind: MeshService
        labels:
          kuma.io/display-name: test-server
          kuma.io/zone: kuma-1
          k8s.kuma.io/namespace: meshhttproutems
      rules:
        - matches:
          - path:
              value: /zone1
              type: PathPrefix
          default:
            backendRefs:
              - kind: MeshService
                labels:
                  kuma.io/display-name: test-server
                  kuma.io/zone: kuma-1
                  k8s.kuma.io/namespace: meshhttproutems
                port: 80
        - matches:
          - path:
              value: /zone2
              type: PathPrefix
          default:
            backendRefs:
              - kind: MeshService
                labels:
                  kuma.io/display-name: test-server
                  kuma.io/zone: kuma-2
                  k8s.kuma.io/namespace: meshhttproutems
                port: 80
        - matches:
          - path:
              value: /multizone
              type: PathPrefix
          default:
            backendRefs:
              - kind: MeshMultiZoneService
                labels:
                  kuma.io/display-name: test-server
                port: 80
`, meshName))(multizone.Global)).To(Succeed())

		execRequest := func(path string) (types.EchoResponse, error) {
			return client.CollectEchoResponse(multizone.KubeZone1, "demo-client", "test-server:80"+path,
				client.FromKubernetesPod(meshName, "demo-client"),
			)
		}

		Eventually(func(g Gomega) {
			response, err := execRequest("/zone1")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(Equal("kube-test-server-1"))

			response, err = execRequest("/zone2")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(Equal("kube-test-server-2"))

			response, err = execRequest("/multizone")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(Equal("kube-test-server-2"))
		}, "30s", "500ms").MustPassRepeatedly(5).Should(Succeed())
	})
}
