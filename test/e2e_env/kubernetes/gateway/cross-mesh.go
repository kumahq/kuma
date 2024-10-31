package gateway

import (
	"fmt"
	"net"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func CrossMeshGatewayOnKubernetes() {
	const gatewayClientNamespaceOtherMesh = "cross-mesh-kuma-client-other"
	const gatewayClientNamespaceSameMesh = "cross-mesh-kuma-client"
	const gatewayTestNamespace = "cross-mesh-kuma-test"
	const gatewayTestNamespace2 = "cross-mesh-kuma-test2"
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

	BeforeAll(func() {
		setup := NewClusterSetup().
			Install(MTLSMeshKubernetes(gatewayMesh)).
			Install(MTLSMeshKubernetes(gatewayOtherMesh)).
			Install(MeshTrafficPermissionAllowAllKubernetes(gatewayMesh)).
			Install(MeshTrafficPermissionAllowAllKubernetes(gatewayOtherMesh)).
			Install(NamespaceWithSidecarInjection(gatewayTestNamespace)).
			Install(NamespaceWithSidecarInjection(gatewayTestNamespace2)).
			Install(NamespaceWithSidecarInjection(gatewayClientNamespaceOtherMesh)).
			Install(NamespaceWithSidecarInjection(gatewayClientNamespaceSameMesh)).
			Install(Namespace(gatewayClientOutsideMesh)).
			Install(Parallel(
				echoServerApp(gatewayMesh),
				echoServerApp(gatewayOtherMesh),
				democlient.Install(democlient.WithNamespace(gatewayClientNamespaceOtherMesh), democlient.WithMesh(gatewayOtherMesh)),
				democlient.Install(democlient.WithNamespace(gatewayClientNamespaceSameMesh), democlient.WithMesh(gatewayMesh)),
				democlient.Install(democlient.WithNamespace(gatewayClientOutsideMesh), democlient.WithMesh(gatewayMesh)), // this will not be in the mesh
			))

		Expect(setup.Setup(kubernetes.Cluster)).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugKube(kubernetes.Cluster, gatewayMesh, gatewayTestNamespace, gatewayTestNamespace2, gatewayClientNamespaceSameMesh)
		DebugKube(kubernetes.Cluster, gatewayOtherMesh, gatewayClientNamespaceOtherMesh)
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(gatewayClientNamespaceOtherMesh)).To(Succeed())
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(gatewayClientNamespaceSameMesh)).To(Succeed())
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(gatewayClientOutsideMesh)).To(Succeed())
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(gatewayTestNamespace)).To(Succeed())
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(gatewayTestNamespace2)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(gatewayMesh)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(gatewayOtherMesh)).To(Succeed())
	})

	Context("when mTLS is enabled", func() {
		crossMeshGatewayYaml := mkGateway(
			crossMeshGatewayName, fmt.Sprintf("%s_%s_svc", crossMeshGatewayName, gatewayTestNamespace), gatewayMesh, true, crossMeshHostname, echoServerService(gatewayMesh, gatewayTestNamespace), crossMeshGatewayPort,
		)
		crossMeshGatewayInstanceYaml := MkGatewayInstance(crossMeshGatewayName, gatewayTestNamespace, gatewayMesh)
		edgeGatewayYaml := mkGateway(
			edgeGatewayName, fmt.Sprintf("%s_%s_svc", edgeGatewayName, gatewayTestNamespace), gatewayOtherMesh, false, "", echoServerService(gatewayOtherMesh, gatewayTestNamespace), edgeGatewayPort,
		)
		edgeGatewayInstanceYaml := MkGatewayInstance(
			edgeGatewayName, gatewayTestNamespace, gatewayOtherMesh,
		)

		BeforeAll(func() {
			setup := NewClusterSetup().
				Install(YamlK8s(crossMeshGatewayYaml)).
				Install(YamlK8s(crossMeshGatewayInstanceYaml)).
				Install(YamlK8s(edgeGatewayYaml)).
				Install(YamlK8s(edgeGatewayInstanceYaml))
			Expect(setup.Setup(kubernetes.Cluster)).To(Succeed())
		})
		E2EAfterAll(func() {
			setup := NewClusterSetup().
				Install(DeleteYamlK8s(crossMeshGatewayYaml)).
				Install(DeleteYamlK8s(crossMeshGatewayInstanceYaml)).
				Install(DeleteYamlK8s(edgeGatewayYaml)).
				Install(DeleteYamlK8s(edgeGatewayInstanceYaml))
			Expect(setup.Setup(kubernetes.Cluster)).To(Succeed())
		})

		It("should proxy HTTP requests from a different mesh", func() {
			gatewayAddr := net.JoinHostPort(crossMeshHostname, strconv.Itoa(crossMeshGatewayPort))
			Eventually(SuccessfullyProxyRequestToGateway(
				kubernetes.Cluster, gatewayMesh,
				gatewayAddr,
				gatewayClientNamespaceOtherMesh,
			), "1m", "1s").Should(Succeed())
		})

		It("should proxy HTTP requests from the same mesh", func() {
			gatewayAddr := net.JoinHostPort(crossMeshHostname, strconv.Itoa(crossMeshGatewayPort))
			Eventually(SuccessfullyProxyRequestToGateway(
				kubernetes.Cluster, gatewayMesh,
				gatewayAddr,
				gatewayClientNamespaceSameMesh,
			), "1m", "1s").Should(Succeed())
		})

		It("doesn't allow HTTP requests from outside the mesh", func() {
			gatewayAddr := gatewayAddress(crossMeshGatewayName, gatewayTestNamespace, crossMeshGatewayPort)
			Consistently(FailToProxyRequestToGateway(
				kubernetes.Cluster,
				gatewayAddr,
				gatewayClientOutsideMesh,
			), "1m", "1s").Should(Succeed())
		})

		It("HTTP requests to a non-crossMesh gateway should still be proxied", func() {
			gatewayAddr := gatewayAddress(edgeGatewayName, gatewayTestNamespace, edgeGatewayPort)
			Eventually(SuccessfullyProxyRequestToGateway(
				kubernetes.Cluster, gatewayOtherMesh,
				gatewayAddr,
				gatewayClientNamespaceOtherMesh,
			)).Should(Succeed())
		})

		It("doesn't break when two cross-mesh gateways exist with the same service value", func() {
			const gatewayMesh2 = "cross-mesh-gateway2"
			crossMeshGatewayYaml2 := mkGateway(
				crossMeshGatewayName+"2", crossMeshGatewayName, gatewayMesh2, true, "gateway2.mesh", echoServerService(gatewayMesh2, gatewayTestNamespace), crossMeshGatewayPort,
			)
			crossMeshGatewayInstanceYaml2 := MkGatewayInstance(crossMeshGatewayName, gatewayTestNamespace2, gatewayMesh2)

			setup := NewClusterSetup().
				Install(MTLSMeshKubernetes(gatewayMesh2)).
				Install(MeshTrafficPermissionAllowAllKubernetes(gatewayMesh2)).
				Install(YamlK8s(crossMeshGatewayYaml2)).
				Install(YamlK8s(crossMeshGatewayInstanceYaml2))
			Expect(setup.Setup(kubernetes.Cluster)).To(Succeed())

			gatewayAddr := net.JoinHostPort(crossMeshHostname, strconv.Itoa(crossMeshGatewayPort))
			Consistently(FailToProxyRequestToGateway(
				kubernetes.Cluster,
				gatewayAddr,
				gatewayClientNamespaceOtherMesh,
			), "30s", "1s").ShouldNot(Succeed())

			setup = NewClusterSetup().
				Install(DeleteYamlK8s(crossMeshGatewayYaml2)).
				Install(DeleteYamlK8s(crossMeshGatewayInstanceYaml2))
			Expect(setup.Setup(kubernetes.Cluster)).To(Succeed())
			Expect(kubernetes.Cluster.DeleteMesh(gatewayMesh2)).To(Succeed())
		})
	})

	Context("with Gateway API", func() {
		const gatewayClass = `
apiVersion: gateway.networking.k8s.io/v1beta1
kind: GatewayClass
metadata:
  name: kuma-cross-mesh
spec:
  controllerName: "gateways.kuma.io/controller"
  parametersRef:
    group: kuma.io
    kind: MeshGatewayConfig
    name: default-cross-mesh
`
		const meshGatewayConfig = `
apiVersion: kuma.io/v1alpha1
kind: MeshGatewayConfig
metadata:
  name: default-cross-mesh
spec:
  crossMesh: true
`
		gateway := fmt.Sprintf(`
apiVersion: gateway.networking.k8s.io/v1beta1
kind: Gateway
metadata:
  name: %s
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  gatewayClassName: kuma-cross-mesh
  listeners:
  - name: proxy
    port: %d
    hostname: %s
    protocol: HTTP
`, crossMeshGatewayName, gatewayTestNamespace, gatewayMesh, crossMeshGatewayPort, crossMeshHostname)
		route := fmt.Sprintf(`
apiVersion: gateway.networking.k8s.io/v1beta1
kind: HTTPRoute
metadata:
  name: %s
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  parentRefs:
  - name: %s
  rules:
  - backendRefs:
    - name: %s
      port: 80
    matches:
    - path:
        type: PathPrefix
        value: /
`, crossMeshGatewayName, gatewayTestNamespace, gatewayMesh, crossMeshGatewayName, echoServerName(gatewayMesh))
		BeforeAll(func() {
			setup := NewClusterSetup().
				Install(YamlK8s(meshGatewayConfig)).
				Install(YamlK8s(gatewayClass)).
				Install(YamlK8s(gateway)).
				Install(YamlK8s(route))
			Expect(setup.Setup(kubernetes.Cluster)).To(Succeed())
		})
		E2EAfterAll(func() {
			setup := NewClusterSetup().
				Install(DeleteYamlK8s(gateway)).
				Install(DeleteYamlK8s(route))
			Expect(setup.Setup(kubernetes.Cluster)).To(Succeed())
		})
		It("should proxy HTTP requests from a different mesh", func() {
			gatewayAddr := net.JoinHostPort(crossMeshHostname, strconv.Itoa(crossMeshGatewayPort))
			Eventually(SuccessfullyProxyRequestToGateway(
				kubernetes.Cluster, gatewayMesh,
				gatewayAddr,
				gatewayClientNamespaceOtherMesh,
			), "1m", "1s").Should(Succeed())
		})
	})
}
