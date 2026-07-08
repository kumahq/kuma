package gateway

import (
	"fmt"
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	. "github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/client"
	"github.com/kumahq/kuma/v3/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/v3/test/framework/envs/multizone"
)

func GatewayHybrid() {
	meshName := "gateway-hybrid"
	namespace := "gateway-hybrid"

	const KubeResponse = "kubernetes"
	const UniResponse = "universal"

	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(MTLSMeshUniversal(meshName)).
			Install(MeshTrafficPermissionAllowAllUniversal(meshName)).
			Setup(multizone.Global)).To(Succeed())
		Expect(WaitForMesh(meshName, multizone.Zones())).To(Succeed())

		group := errgroup.Group{}
		group.Go(func() error {
			err := NewClusterSetup().
				Install(NamespaceWithSidecarInjection(namespace)).
				Install(testserver.Install(
					testserver.WithEchoArgs("echo", "--instance", KubeResponse),
					testserver.WithNamespace(namespace),
					testserver.WithMesh(meshName),
				)).
				Setup(multizone.KubeZone1)
			return errors.Wrap(err, multizone.KubeZone1.Name())
		})

		group.Go(func() error {
			err := NewClusterSetup().
				Install(Parallel(
					GatewayProxyUniversal(meshName, "edge-gateway"),
					TestServerUniversal("test-server", meshName,
						WithArgs([]string{"echo", "--instance", UniResponse}),
					),
					TestServerUniversal("gateway-client", meshName, WithoutDataplane()),
				)).
				Setup(multizone.UniZone1)
			return errors.Wrap(err, multizone.UniZone1.Name())
		})

		Expect(group.Wait()).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(multizone.Global, meshName)
		DebugUniversal(multizone.UniZone1, meshName)
		DebugKube(multizone.KubeZone1, meshName, namespace)
	})

	E2EAfterAll(func() {
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.UniZone1.DeleteMeshApps(meshName)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(meshName)).To(Succeed())
	})

	type testCase struct {
		path              string
		expectedInstances []string
	}

	Context("proxying through instances across all zones", func() {
		DescribeTable("gateway proxies the traffic to echo service",
			func(given testCase) {
				err := NewClusterSetup().
					Install(YamlUniversal(`
type: MeshGateway
mesh: gateway-hybrid
name: edge-gateway
selectors:
- match:
    kuma.io/service: edge-gateway
conf:
  listeners:
  - port: 8080
    protocol: HTTP
    hostname: example.kuma.io
    tags:
      hostname: example.kuma.io
`,
					)).
					Install(YamlUniversal(fmt.Sprintf(`
type: MeshHTTPRoute
mesh: gateway-hybrid
name: edge-gateway
spec:
  targetRef:
    kind: MeshGateway
    name: edge-gateway
  to:
  - targetRef:
      kind: Mesh
    rules:
    - matches:
      - path:
          type: PathPrefix
          value: /kubernetes
      default:
        backendRefs:
        - kind: MeshService
          labels:
            kuma.io/display-name: test-server
            kuma.io/zone: %[1]s
            k8s.kuma.io/namespace: %[3]s
          port: 80
    - matches:
      - path:
          type: PathPrefix
          value: /universal
      default:
        backendRefs:
        - kind: MeshService
          labels:
            kuma.io/display-name: test-server
            kuma.io/zone: %[2]s
          port: 80
    - matches:
      - path:
          type: PathPrefix
          value: /all
      default:
        backendRefs:
        - kind: MeshService
          labels:
            kuma.io/display-name: test-server
            kuma.io/zone: %[1]s
            k8s.kuma.io/namespace: %[3]s
          port: 80
        - kind: MeshService
          labels:
            kuma.io/display-name: test-server
            kuma.io/zone: %[2]s
          port: 80
`, multizone.KubeZone1.ZoneName(), multizone.UniZone1.ZoneName(), namespace))).
					Setup(multizone.Global)
				Expect(err).ToNot(HaveOccurred())

				Eventually(func(g Gomega) {
					gatewayIP := multizone.UniZone1.GetApp("edge-gateway").GetIP()
					responses, err := client.CollectResponsesByInstance(
						multizone.UniZone1,
						"gateway-client",
						fmt.Sprintf("http://%s%s", net.JoinHostPort(gatewayIP, "8080"), given.path),
						client.WithHeader("Host", "example.kuma.io"),
					)
					g.Expect(err).To(Succeed())

					g.Expect(responses).To(HaveLen(len(given.expectedInstances)))
					for _, expectedInstance := range given.expectedInstances {
						g.Expect(responses).To(HaveKey(expectedInstance))
					}
				}, "30s", "1s").Should(Succeed())
			},
			Entry("should proxy between all instances", testCase{
				path:              "/all",
				expectedInstances: []string{KubeResponse, UniResponse},
			}),
			Entry("should proxy to the kubernetes", testCase{
				path:              "/kubernetes",
				expectedInstances: []string{KubeResponse},
			}),
			Entry("should proxy to the universal", testCase{
				path:              "/universal",
				expectedInstances: []string{UniResponse},
			}),
		)
	}, Ordered)
	// Ordered above is important because CollectResponsesByInstance requires `gateway-client` app to exist in UniversalCluster.
	// Since apps are not synchronized between instances of test suite, it has to be run by instance that deploys it.
}
