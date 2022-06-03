package gateway

import (
	"fmt"
	"net"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	k8s_gateway "github.com/kumahq/kuma/test/e2e_env/kubernetes/gateway"
	"github.com/kumahq/kuma/test/e2e_env/multizone/env"
	universal_gateway "github.com/kumahq/kuma/test/e2e_env/universal/gateway"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func CrossMeshGatewayOnMultizone() {
	const gatewayClientNamespaceOtherMesh = "cross-mesh-kuma-client-other"
	const gatewayClientNamespaceSameMesh = "cross-mesh-kuma-client"
	const gatewayTestNamespace = "cross-mesh-kuma-test"

	const crossMeshHostname = "gateway.mesh"

	echoServerName := func(mesh string) string {
		return fmt.Sprintf("echo-server-%s", mesh)
	}
	echoServerService := func(mesh, namespace string) string {
		return fmt.Sprintf("%s_%s_svc_80", echoServerName(mesh), namespace)
	}

	const crossMeshGatewayName = "cross-mesh-gateway"
	const edgeGatewayName = "cross-mesh-edge-gateway"

	const gatewayMesh = "cross-mesh-gateway"
	const gatewayOtherMesh = "cross-mesh-other"

	const crossMeshGatewayPort = 9080
	const edgeGatewayPort = 9081

	echoServerApp := func(mesh string) InstallFunc {
		return testserver.Install(
			testserver.WithMesh(mesh),
			testserver.WithName(echoServerName(mesh)),
			testserver.WithNamespace(gatewayTestNamespace),
			testserver.WithArgs("echo", "--instance", mesh),
		)
	}

	crossMeshGatewayYaml := universal_gateway.MkGateway(
		crossMeshGatewayName, gatewayMesh, true, crossMeshHostname, echoServerService(gatewayMesh, gatewayTestNamespace), crossMeshGatewayPort,
	)
	crossMeshGatewayInstanceYaml := k8s_gateway.MkGatewayInstance(crossMeshGatewayName, gatewayTestNamespace, gatewayMesh)
	edgeGatewayYaml := universal_gateway.MkGateway(
		edgeGatewayName, gatewayOtherMesh, false, "", echoServerService(gatewayOtherMesh, gatewayTestNamespace), edgeGatewayPort,
	)
	edgeGatewayInstanceYaml := k8s_gateway.MkGatewayInstance(
		edgeGatewayName, gatewayTestNamespace, gatewayOtherMesh,
	)

	BeforeAll(func() {
		globalSetup := NewClusterSetup().
			Install(MTLSMeshUniversal(gatewayMesh)).
			Install(MTLSMeshUniversal(gatewayOtherMesh)).
			Install(YamlUniversal(crossMeshGatewayYaml)).
			Install(YamlUniversal(edgeGatewayYaml))
		Expect(globalSetup.Setup(env.Global)).To(Succeed())

		gatewayZoneSetup := NewClusterSetup().
			Install(NamespaceWithSidecarInjection(gatewayTestNamespace)).
			Install(NamespaceWithSidecarInjection(gatewayClientNamespaceOtherMesh)).
			Install(NamespaceWithSidecarInjection(gatewayClientNamespaceSameMesh)).
			Install(echoServerApp(gatewayMesh)).
			Install(echoServerApp(gatewayOtherMesh)).
			Install(DemoClientK8s(gatewayOtherMesh, gatewayClientNamespaceOtherMesh)).
			Install(DemoClientK8s(gatewayMesh, gatewayClientNamespaceSameMesh)).
			Install(YamlK8s(crossMeshGatewayInstanceYaml)).
			Install(YamlK8s(edgeGatewayInstanceYaml))
		Expect(gatewayZoneSetup.Setup(env.KubeZone1)).To(Succeed())

		otherZoneSetup := NewClusterSetup().
			Install(NamespaceWithSidecarInjection(gatewayClientNamespaceOtherMesh)).
			Install(NamespaceWithSidecarInjection(gatewayClientNamespaceSameMesh)).
			Install(DemoClientK8s(gatewayOtherMesh, gatewayClientNamespaceOtherMesh)).
			Install(DemoClientK8s(gatewayMesh, gatewayClientNamespaceSameMesh))
		Expect(otherZoneSetup.Setup(env.KubeZone2)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(env.KubeZone1.TriggerDeleteNamespace(gatewayClientNamespaceOtherMesh)).To(Succeed())
		Expect(env.KubeZone1.TriggerDeleteNamespace(gatewayClientNamespaceSameMesh)).To(Succeed())
		Expect(env.KubeZone1.TriggerDeleteNamespace(gatewayTestNamespace)).To(Succeed())

		Expect(env.KubeZone2.TriggerDeleteNamespace(gatewayClientNamespaceOtherMesh)).To(Succeed())
		Expect(env.KubeZone2.TriggerDeleteNamespace(gatewayClientNamespaceSameMesh)).To(Succeed())

		Expect(env.Global.DeleteMesh(gatewayMesh)).To(Succeed())
		Expect(env.Global.DeleteMesh(gatewayOtherMesh)).To(Succeed())
	})

	Context("when mTLS is enabled", func() {
		It("should proxy HTTP requests from a different mesh in the same zone", func() {
			gatewayAddr := net.JoinHostPort(crossMeshHostname, strconv.Itoa(crossMeshGatewayPort))
			k8s_gateway.SuccessfullyProxyRequestToGateway(
				env.KubeZone1, gatewayMesh,
				gatewayAddr,
				gatewayClientNamespaceOtherMesh,
			)
		})
		It("should proxy HTTP requests from the same mesh in the same zone", func() {
			gatewayAddr := net.JoinHostPort(crossMeshHostname, strconv.Itoa(crossMeshGatewayPort))
			k8s_gateway.SuccessfullyProxyRequestToGateway(
				env.KubeZone1, gatewayMesh,
				gatewayAddr,
				gatewayClientNamespaceSameMesh,
			)
		})
	})
}
