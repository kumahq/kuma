package gateway

import (
	"fmt"
	"net"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/universal/env"
	. "github.com/kumahq/kuma/test/framework"
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

	echoServerApp := func(mesh string) InstallFunc {
		return TestServerUniversal(
			echoServerName(mesh),
			mesh,
			WithArgs([]string{"echo", "--instance", mesh}),
			WithServiceName(echoServerName(mesh)),
		)
	}

	crossMeshGatewayYaml := mkGateway(
		crossMeshGatewayName, gatewayMesh, true, crossMeshHostname, echoServerName(gatewayMesh), crossMeshGatewayPort,
	)
	crossMeshGatewayDataplane := mkGatewayDataplane(crossMeshGatewayName, gatewayMesh)
	edgeGatewayYaml := mkGateway(
		edgeGatewayName, gatewayOtherMesh, false, "", echoServerName(gatewayOtherMesh), edgeGatewayPort,
	)
	edgeGatewayDataplane := mkGatewayDataplane(edgeGatewayName, gatewayOtherMesh)

	BeforeAll(func() {
		setup := NewClusterSetup().
			Install(MTLSMeshUniversal(gatewayMesh)).
			Install(MTLSMeshUniversal(gatewayOtherMesh)).
			Install(DemoClientUniversal(demoClientName(gatewayMesh), gatewayMesh, WithTransparentProxy(true))).
			Install(DemoClientUniversal(demoClientName(gatewayOtherMesh), gatewayOtherMesh, WithTransparentProxy(true))).
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
			successfullyProxyRequestToGateway(
				env.Cluster, gatewayOtherMesh, gatewayMesh,
				gatewayAddr,
			)
		})

		It("should proxy HTTP requests from the same mesh", func() {
			gatewayAddr := net.JoinHostPort(crossMeshHostname, strconv.Itoa(crossMeshGatewayPort))
			successfullyProxyRequestToGateway(
				env.Cluster, gatewayMesh, gatewayMesh,
				gatewayAddr,
			)
		})

		Specify("HTTP requests to a non-crossMesh gateway should still be proxied", func() {
			gatewayAddr := net.JoinHostPort(env.Cluster.GetApp(edgeGatewayName).GetIP(), strconv.Itoa(edgeGatewayPort))
			successfullyProxyRequestToGateway(
				env.Cluster, gatewayOtherMesh, gatewayOtherMesh,
				gatewayAddr,
			)
		})
	})
}
