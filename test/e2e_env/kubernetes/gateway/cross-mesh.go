package gateway

import (
	"fmt"
	"net"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/kubernetes/env"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func CrossMeshGatewayOnKubernetes() {
	const gatewayClientNamespaceOtherMesh = "cross-mesh-kuma-client-other"
	const gatewayClientNamespaceSameMesh = "cross-mesh-kuma-client"
	const gatewayTestNamespace = "cross-mesh-kuma-test"
	const gatewayClientOutsideMesh = "cross-mesh-kuma-client-outside"

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
			testserver.WithEchoArgs("echo", "--instance", mesh),
		)
	}

	crossMeshGatewayYaml := mkGateway(
		crossMeshGatewayName, gatewayTestNamespace, gatewayMesh, true, crossMeshHostname, echoServerService(gatewayMesh, gatewayTestNamespace), crossMeshGatewayPort,
	)
	crossMeshGatewayInstanceYaml := MkGatewayInstance(crossMeshGatewayName, gatewayTestNamespace, gatewayMesh)
	edgeGatewayYaml := mkGateway(
		edgeGatewayName, gatewayTestNamespace, gatewayOtherMesh, false, "", echoServerService(gatewayOtherMesh, gatewayTestNamespace), edgeGatewayPort,
	)
	edgeGatewayInstanceYaml := MkGatewayInstance(
		edgeGatewayName, gatewayTestNamespace, gatewayOtherMesh,
	)

	BeforeAll(func() {
		setup := NewClusterSetup().
			Install(MTLSMeshKubernetes(gatewayMesh)).
			Install(MTLSMeshKubernetes(gatewayOtherMesh)).
			Install(NamespaceWithSidecarInjection(gatewayTestNamespace)).
			Install(NamespaceWithSidecarInjection(gatewayClientNamespaceOtherMesh)).
			Install(NamespaceWithSidecarInjection(gatewayClientNamespaceSameMesh)).
			Install(Namespace(gatewayClientOutsideMesh)).
			Install(echoServerApp(gatewayMesh)).
			Install(echoServerApp(gatewayOtherMesh)).
			Install(DemoClientK8s(gatewayOtherMesh, gatewayClientNamespaceOtherMesh)).
			Install(DemoClientK8s(gatewayMesh, gatewayClientNamespaceSameMesh)).
			Install(DemoClientK8s(gatewayMesh, gatewayClientOutsideMesh)). // this will not be in the mesh
			Install(YamlK8s(crossMeshGatewayYaml)).
			Install(YamlK8s(crossMeshGatewayInstanceYaml)).
			Install(YamlK8s(edgeGatewayYaml)).
			Install(YamlK8s(edgeGatewayInstanceYaml))

		Expect(setup.Setup(env.Cluster)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(env.Cluster.TriggerDeleteNamespace(gatewayClientNamespaceOtherMesh)).To(Succeed())
		Expect(env.Cluster.TriggerDeleteNamespace(gatewayClientNamespaceSameMesh)).To(Succeed())
		Expect(env.Cluster.TriggerDeleteNamespace(gatewayTestNamespace)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(gatewayMesh)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(gatewayOtherMesh)).To(Succeed())
	})

	Context("when mTLS is enabled", func() {
		It("should proxy HTTP requests from a different mesh", func() {
			gatewayAddr := net.JoinHostPort(crossMeshHostname, strconv.Itoa(crossMeshGatewayPort))
			SuccessfullyProxyRequestToGateway(
				env.Cluster, gatewayMesh,
				gatewayAddr,
				gatewayClientNamespaceOtherMesh,
			)
		})

		It("should proxy HTTP requests from the same mesh", func() {
			gatewayAddr := net.JoinHostPort(crossMeshHostname, strconv.Itoa(crossMeshGatewayPort))
			SuccessfullyProxyRequestToGateway(
				env.Cluster, gatewayMesh,
				gatewayAddr,
				gatewayClientNamespaceSameMesh,
			)
		})

		It("doesn't allow HTTP requests from outside the mesh", func() {
			gatewayAddr := gatewayAddress(crossMeshGatewayName, gatewayTestNamespace, crossMeshGatewayPort)
			Consistently(FailToProxyRequestToGateway(
				env.Cluster,
				gatewayAddr,
				gatewayClientOutsideMesh,
			), "10s", "1s").Should(Succeed())
		})

		Specify("HTTP requests to a non-crossMesh gateway should still be proxied", func() {
			gatewayAddr := gatewayAddress(edgeGatewayName, gatewayTestNamespace, edgeGatewayPort)
			SuccessfullyProxyRequestToGateway(
				env.Cluster, gatewayOtherMesh,
				gatewayAddr,
				gatewayClientNamespaceOtherMesh,
			)
		})
	})
}
