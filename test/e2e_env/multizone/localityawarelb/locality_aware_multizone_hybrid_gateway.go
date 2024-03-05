package localityawarelb

import (
	"fmt"
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/test/resources/samples"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func LocalityAwareLBGateway() {
	const mesh = "locality-aware-lb-gateway"
	const namespace = "locality-aware-lb-gateway"

	meshLoadBalancingStrategyTestServer := fmt.Sprintf(`
type: MeshLoadBalancingStrategy
name: mlbs-1
mesh: %s
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshService
        name: test-server_locality-aware-lb-gateway_svc_80
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

	meshGateway := YamlUniversal(fmt.Sprintf(`
type: MeshGateway
mesh: %s
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
`, mesh))
	meshGatewayRoute := YamlUniversal(fmt.Sprintf(`
type: MeshGatewayRoute
mesh: %s
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
          value: /
      backends:
      - destination:
          kuma.io/service: test-server_locality-aware-lb-gateway_svc_80
`, mesh))

	BeforeAll(func() {
		// Global
		Expect(NewClusterSetup().
			Install(ResourceUniversal(samples.MeshMTLSBuilder().WithName(mesh).Build())).
			Install(MeshTrafficPermissionAllowAllUniversal(mesh)).
			Setup(multizone.Global)).To(Succeed())
		Expect(WaitForMesh(mesh, multizone.Zones())).To(Succeed())

		// Universal Zone 4
		Expect(NewClusterSetup().
			Install(GatewayProxyUniversal(mesh, "edge-gateway")).
			Install(TestServerUniversal("gateway-client", mesh, WithoutDataplane())).
			Install(TestServerUniversal("test-server-zone-4", mesh,
				WithServiceName("test-server_locality-aware-lb-gateway_svc_80"),
				WithArgs([]string{"echo", "--instance", "test-server-zone-4"}),
			)).
			Setup(multizone.UniZone1),
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

		// Universal Zone 5
		Expect(NewClusterSetup().
			Install(DemoClientUniversal(
				"demo-client_locality-aware-lb-gateway_svc",
				mesh,
				WithTransparentProxy(true),
			)).
			Install(TestServerUniversal("test-server-zone-5", mesh,
				WithServiceName("test-server_locality-aware-lb-gateway_svc_80"),
				WithArgs([]string{"echo", "--instance", "test-server-zone-5"}),
			)).
			Setup(multizone.UniZone2),
		).To(Succeed())

		Expect(multizone.Global.Install(meshGateway)).To(Succeed())
		Expect(multizone.Global.Install(meshGatewayRoute)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(multizone.UniZone1.DeleteMeshApps(mesh)).To(Succeed())
		Expect(multizone.UniZone2.DeleteMeshApps(mesh)).To(Succeed())
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(mesh)).To(Succeed())
	})

	It("should route based on defined strategy when making requests through gateway", func() {
		// no lb priorities
		gatewayIP := multizone.UniZone1.GetApp("edge-gateway").GetIP()
		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(multizone.UniZone1, "gateway-client", fmt.Sprintf("http://%s", net.JoinHostPort(gatewayIP, "8080")), client.WithHeader("Host", "example.kuma.io"), client.WithNumberOfRequests(100))
		}, "2m", "10s").Should(
			And(
				HaveKeyWithValue(Equal(`test-server-zone-4`), BeNumerically("~", 50, 10)),
				HaveKeyWithValue(Equal(`test-server-zone-5`), BeNumerically("~", 25, 10)),
				HaveKeyWithValue(Equal(`test-server-zone-1`), BeNumerically("~", 25, 10)),
			),
		)

		// apply lb policy
		// kuma-4 - priority 0, kuma-5 - priority 1, kuma-1 - priority 2
		Expect(multizone.Global.Install(YamlUniversal(meshLoadBalancingStrategyTestServer))).To(Succeed())

		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(multizone.UniZone1, "gateway-client", fmt.Sprintf("http://%s", net.JoinHostPort(gatewayIP, "8080")), client.WithHeader("Host", "example.kuma.io"), client.WithNumberOfRequests(100))
		}, "1m", "10s").Should(
			HaveKeyWithValue(Equal(`test-server-zone-4`), BeNumerically("~", 100, 10)),
		)

		// kill test-server in kuma-4 zone
		Expect(multizone.UniZone1.DeleteApp("test-server-zone-4")).To(Succeed())

		// traffic goes to kuma-5 zone
		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(multizone.UniZone1, "gateway-client", fmt.Sprintf("http://%s", net.JoinHostPort(gatewayIP, "8080")), client.WithHeader("Host", "example.kuma.io"), client.WithNumberOfRequests(100))
		}, "1m", "10s").Should(
			HaveKeyWithValue(Equal(`test-server-zone-5`), BeNumerically("~", 100, 10)),
		)

		// kill test-server from kuma-5 zone
		Expect(multizone.UniZone2.DeleteApp("test-server-zone-5")).To(Succeed())

		// traffic goes to kuma-1 zone
		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(multizone.UniZone1, "gateway-client", fmt.Sprintf("http://%s", net.JoinHostPort(gatewayIP, "8080")), client.WithHeader("Host", "example.kuma.io"), client.WithNumberOfRequests(100))
		}, "1m", "10s").Should(
			HaveKeyWithValue(Equal(`test-server-zone-1`), BeNumerically("~", 100, 10)),
		)
	})
}
