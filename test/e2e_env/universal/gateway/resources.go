package gateway

import (
	"fmt"
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/client"
	"github.com/kumahq/kuma/v3/test/framework/envs/universal"
)

func Resources() {
	meshName := "gateway-resources"
	gatewayName := "resources-gateway"
	const gatewayPort = 8080

	meshGatewayWithoutLimit := fmt.Sprintf(`
type: MeshGateway
name: %s
mesh: %s
selectors:
- match:
    kuma.io/service: %s
conf:
  listeners:
  - port: %d
    protocol: HTTP
`, gatewayName, meshName, gatewayName, gatewayPort)

	meshGatewayWithLimit := fmt.Sprintf(`
type: MeshGateway
name: %s
mesh: %s
selectors:
- match:
    kuma.io/service: %s
conf:
  listeners:
  - port: %d
    protocol: HTTP
    resources:
      connectionLimit: 1
`, gatewayName, meshName, gatewayName, gatewayPort)

	httpRoute := fmt.Sprintf(`
type: MeshHTTPRoute
name: %s
mesh: %s
spec:
  targetRef:
    kind: MeshGateway
    name: %s
  to:
  - targetRef:
      kind: Mesh
    rules:
    - matches:
      - path:
          type: PathPrefix
          value: /
      default:
        backendRefs:
        - kind: MeshService
          name: echo-service-resources
          weight: 100
`, gatewayName, meshName, gatewayName)

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MTLSMeshUniversal(meshName)).
			Install(MeshTrafficPermissionAllowAllUniversal(meshName)).
			Install(GatewayClientAppUniversal("resources-wait-client")).
			Install(GatewayClientAppUniversal("resources-curl-client")).
			Install(EchoServerApp(meshName, "echo-server-resources", "echo-service-resources", "universal")).
			Install(GatewayProxyUniversal(meshName, gatewayName)).
			Install(YamlUniversal(meshGatewayWithoutLimit)).
			Install(YamlUniversal(httpRoute)).
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, meshName)
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteApp("resources-wait-client")).To(Succeed())
		Expect(universal.Cluster.DeleteApp("resources-curl-client")).To(Succeed())
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	gatewayAddr := func() string {
		return net.JoinHostPort(universal.Cluster.GetApp(gatewayName).GetIP(), fmt.Sprintf("%d", gatewayPort))
	}

	Specify("connection limit is respected", func() {
		target := fmt.Sprintf("http://%s/", gatewayAddr())

		By("allowing connections without a limit")

		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(
				universal.Cluster, "resources-curl-client", target,
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(Equal("universal"))
		}, "30s", "1s").MustPassRepeatedly(3).Should(Succeed())

		By("allowing more than 1 connection without a limit")

		cancelFirstConn := HoldConnection(universal.Cluster, "resources-wait-client", gatewayName, gatewayPort)

		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(
				universal.Cluster, "resources-curl-client", target,
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(Equal("universal"))
		}, "1m", "1s").MustPassRepeatedly(3).Should(Succeed())

		By("not allowing more than 1 connection with a limit of 1")

		Expect(universal.Cluster.Install(YamlUniversal(meshGatewayWithLimit))).To(Succeed())

		// Close the first keep-alive, then re-occupy the single slot.
		cancelFirstConn()
		cancelSecondConn := HoldConnection(universal.Cluster, "resources-wait-client", gatewayName, gatewayPort)
		defer cancelSecondConn()

		Eventually(func(g Gomega) {
			response, err := client.CollectFailure(
				universal.Cluster, "resources-curl-client", target,
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Exitcode).To(Or(Equal(52), Equal(56)))
		}, "1m", "1s").MustPassRepeatedly(3).Should(Succeed())
	})
}
