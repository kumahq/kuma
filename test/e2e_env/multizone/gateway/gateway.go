package gateway

import (
	"fmt"
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func GatewayHybrid() {
	meshName := "gateway-hybrid"
	namespace := "gateway-hybrid"

	const KubeResponse = "kubernetes"
	const UniResponse = "universal"

	BeforeAll(func() {
		Expect(multizone.Global.Install(MTLSMeshUniversal(meshName))).To(Succeed())
		Expect(WaitForMesh(meshName, multizone.Zones())).To(Succeed())

		err := NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(testserver.Install(
				testserver.WithEchoArgs("echo", "--instance", KubeResponse),
				testserver.WithNamespace(namespace),
				testserver.WithMesh(meshName),
			)).
			Setup(multizone.KubeZone1)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(GatewayProxyUniversal(meshName, "edge-gateway")).
			Install(TestServerUniversal("test-server", meshName,
				WithArgs([]string{"echo", "--instance", UniResponse}),
				WithServiceName("test-server_gateway-hybrid_svc_80"),
			)).
			Install(TestServerUniversal("gateway-client", meshName, WithoutDataplane())).
			Setup(multizone.UniZone1)
		Expect(err).ToNot(HaveOccurred())
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
type: MeshGatewayRoute
mesh: gateway-hybrid
name: edge-gateway
selectors:
- match:
    kuma.io/service: edge-gateway
conf:
  http:
    rules:
    - matches:
      - path:
          match: PREFIX
          value: /kubernetes
      backends:
      - destination:
          kuma.io/service: test-server_gateway-hybrid_svc_80
          kuma.io/zone: %s 
    - matches:
      - path:
          match: PREFIX
          value: /universal
      backends:
      - destination:
          kuma.io/service: test-server_gateway-hybrid_svc_80
          kuma.io/zone: kuma-4
    - matches:
      - path:
          match: PREFIX
          value: /all
      backends:
      - destination:
          kuma.io/service: test-server_gateway-hybrid_svc_80
`, multizone.KubeZone1.ZoneName()))).
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
