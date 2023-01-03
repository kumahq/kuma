package gateway

import (
	"fmt"
	"net"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/universal/env"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
)

func CrossMeshGatewayOnUniversal() {
	const crossMeshHostname = "gateway.mesh"

	echoServerName := func(mesh string) string {
		return fmt.Sprintf("echo-server-%s", mesh)
	}

	const crossMeshGatewayName = "cross-mesh-gateway"
	const edgeGatewayName = "cross-mesh-edge-gateway"

	const gatewayMesh = "cross-mesh-gateway"
	const gatewayOtherMesh = "cross-mesh-other"

	const crossMeshGatewayPort = 9080
	const edgeGatewayPort = 9081

	demoClientInMesh := demoClientName(gatewayMesh)
	demoClientOtherMesh := demoClientName(gatewayOtherMesh)
	demoClientOutsideMesh := demoClientName("outside")
	demoClientNoTransparent := demoClientName(gatewayOtherMesh + "-no-transparent")

	echoServerApp := func(mesh string) InstallFunc {
		return TestServerUniversal(
			echoServerName(mesh),
			mesh,
			WithArgs([]string{"echo", "--instance", mesh}),
			WithServiceName(echoServerName(mesh)),
		)
	}

	crossMeshGatewayYaml := MkGateway(
		crossMeshGatewayName, gatewayMesh, true, crossMeshHostname, echoServerName(gatewayMesh), crossMeshGatewayPort,
	)
	crossMeshGatewayDataplane := mkGatewayDataplane(crossMeshGatewayName, gatewayMesh)
	edgeGatewayYaml := MkGateway(
		edgeGatewayName, gatewayOtherMesh, false, "", echoServerName(gatewayOtherMesh), edgeGatewayPort,
	)
	edgeGatewayDataplane := mkGatewayDataplane(edgeGatewayName, gatewayOtherMesh)

	BeforeAll(func() {
		By("installing one cross-mesh gateway and one non-cross-mesh gateway")
		setup := NewClusterSetup().
			Install(MTLSMeshUniversal(gatewayMesh)).
			Install(MTLSMeshUniversal(gatewayOtherMesh)).
			Install(DemoClientUniversal(demoClientInMesh, gatewayMesh, WithTransparentProxy(true))).
			Install(DemoClientUniversal(demoClientOtherMesh, gatewayOtherMesh, WithTransparentProxy(true))).
			Install(DemoClientUniversal(demoClientOutsideMesh, "", WithoutDataplane())).
			Install(DemoClientUniversal(
				demoClientNoTransparent,
				gatewayOtherMesh,
				WithTransparentProxy(false),
				WithYaml(demoClientDataplaneWithOutbound(demoClientNoTransparent, gatewayOtherMesh, crossMeshGatewayName, gatewayMesh, crossMeshGatewayPort)),
			)).
			Install(echoServerApp(gatewayMesh)).
			Install(echoServerApp(gatewayOtherMesh)).
			Install(YamlUniversal(crossMeshGatewayYaml)).
			Install(YamlUniversal(edgeGatewayYaml)).
			Install(crossMeshGatewayDataplane).
			Install(edgeGatewayDataplane)

		Expect(setup.Setup(env.Cluster)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(env.Cluster.DeleteMeshApps(gatewayMesh)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(gatewayMesh)).To(Succeed())
		Expect(env.Cluster.DeleteMeshApps(gatewayOtherMesh)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(gatewayOtherMesh)).To(Succeed())
	})

	Context("when mTLS is enabled", func() {
		It("should proxy HTTP requests from a different mesh", func() {
			gatewayAddr := net.JoinHostPort(crossMeshHostname, strconv.Itoa(crossMeshGatewayPort))
			Eventually(successfullyProxyRequestToGateway(
				env.Cluster, demoClientOtherMesh, gatewayMesh, gatewayAddr,
			), "30s", "1s").Should(Succeed())
		})

		It("should proxy HTTP requests from the same mesh", func() {
			gatewayAddr := net.JoinHostPort(crossMeshHostname, strconv.Itoa(crossMeshGatewayPort))
			Eventually(successfullyProxyRequestToGateway(
				env.Cluster, demoClientInMesh, gatewayMesh, gatewayAddr,
			)).Should(Succeed())
		})

		It("doesn't allow HTTP requests from outside the mesh", func() {
			gatewayAddr := net.JoinHostPort(crossMeshHostname, strconv.Itoa(crossMeshGatewayPort))
			Consistently(failToProxyRequestToGateway(
				env.Cluster, demoClientOutsideMesh, gatewayAddr, env.Cluster.GetApp(crossMeshGatewayName).GetIP(),
			), "10s", "1s").Should(Succeed())
		})

		Specify("HTTP requests to a non-crossMesh gateway should still be proxied", func() {
			gatewayAddr := net.JoinHostPort(env.Cluster.GetApp(edgeGatewayName).GetIP(), strconv.Itoa(edgeGatewayPort))
			Eventually(successfullyProxyRequestToGateway(
				env.Cluster, demoClientOtherMesh, gatewayOtherMesh, gatewayAddr,
			)).Should(Succeed())
		})

		It("should be reachable without transparent proxy", func() {
			gatewayAddr := net.JoinHostPort(crossMeshHostname, strconv.Itoa(crossMeshGatewayPort))
			Eventually(successfullyProxyRequestToGateway(
				env.Cluster, demoClientNoTransparent, gatewayMesh,
				gatewayAddr, client.Resolve(gatewayAddr, "127.0.0.1"),
			)).Should(Succeed())
		})
	})
}
