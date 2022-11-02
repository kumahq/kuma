package gateway

import (
	"encoding/base64"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/kubernetes/env"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func misconfiguredMTLSProvidedMeshKubernetes() InstallFunc {
	mesh := `
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: misconfiguredmesh
spec:
  mtls:
    enabledBackend: ca-1
    backends:
      - name: ca-1
        type: provided
        conf:
          cert:
            secret: will-be-deleted
          key:
            secret: will-be-deleted
`
	return YamlK8s(mesh)
}

func secret(payload string) InstallFunc {
	secret := fmt.Sprintf(`
apiVersion: v1
kind: Secret
metadata:
  name: will-be-deleted
  namespace: %s
  labels:
    kuma.io/mesh: misconfiguredmesh
data:
  value: %s
type: system.kuma.io/secret
`, Config.KumaNamespace, payload)
	return YamlK8s(secret)
}

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
		cert, key, err := CreateCertsFor("example.kuma.io")
		Expect(err).To(Succeed())

		payload := base64.StdEncoding.EncodeToString([]byte(strings.Join([]string{key, cert}, "\n")))

		setup := NewClusterSetup().
			Install(MTLSMeshKubernetes(gatewayMesh)).
			Install(MTLSMeshKubernetes(gatewayOtherMesh)).
			Install(secret(payload)).
			// We want to make sure meshes continue to work in the presence of a
			// misconfigured mesh
			Install(misconfiguredMTLSProvidedMeshKubernetes()).
			Install(NamespaceWithSidecarInjection(gatewayTestNamespace)).
			Install(NamespaceWithSidecarInjection(gatewayTestNamespace2)).
			Install(NamespaceWithSidecarInjection(gatewayClientNamespaceOtherMesh)).
			Install(NamespaceWithSidecarInjection(gatewayClientNamespaceSameMesh)).
			Install(Namespace(gatewayClientOutsideMesh)).
			Install(echoServerApp(gatewayMesh)).
			Install(echoServerApp(gatewayOtherMesh)).
			Install(DemoClientK8s(gatewayOtherMesh, gatewayClientNamespaceOtherMesh)).
			Install(DemoClientK8s(gatewayMesh, gatewayClientNamespaceSameMesh)).
			Install(DemoClientK8s(gatewayMesh, gatewayClientOutsideMesh)) // this will not be in the mesh

		Expect(setup.Setup(env.Cluster)).To(Succeed())

		Expect(
			k8s.RunKubectlE(env.Cluster.GetTesting(), env.Cluster.GetKubectlOptions(Config.KumaNamespace), "delete", "secret", "will-be-deleted"),
		).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(env.Cluster.TriggerDeleteNamespace(gatewayClientNamespaceOtherMesh)).To(Succeed())
		Expect(env.Cluster.TriggerDeleteNamespace(gatewayClientNamespaceSameMesh)).To(Succeed())
		Expect(env.Cluster.TriggerDeleteNamespace(gatewayClientOutsideMesh)).To(Succeed())
		Expect(env.Cluster.TriggerDeleteNamespace(gatewayTestNamespace)).To(Succeed())
		Expect(env.Cluster.TriggerDeleteNamespace(gatewayTestNamespace2)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(gatewayMesh)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(gatewayOtherMesh)).To(Succeed())
	})

	Context("when mTLS is enabled", func() {
		crossMeshGatewayYaml := mkGateway(
			crossMeshGatewayName, crossMeshGatewayName, gatewayMesh, true, crossMeshHostname, echoServerService(gatewayMesh, gatewayTestNamespace), crossMeshGatewayPort,
		)
		crossMeshGatewayInstanceYaml := MkGatewayInstance(crossMeshGatewayName, gatewayTestNamespace, gatewayMesh)
		edgeGatewayYaml := mkGateway(
			edgeGatewayName, edgeGatewayName, gatewayOtherMesh, false, "", echoServerService(gatewayOtherMesh, gatewayTestNamespace), edgeGatewayPort,
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
			Expect(setup.Setup(env.Cluster)).To(Succeed())
		})
		E2EAfterAll(func() {
			setup := NewClusterSetup().
				Install(DeleteYamlK8s(crossMeshGatewayYaml)).
				Install(DeleteYamlK8s(crossMeshGatewayInstanceYaml)).
				Install(DeleteYamlK8s(edgeGatewayYaml)).
				Install(DeleteYamlK8s(edgeGatewayInstanceYaml))
			Expect(setup.Setup(env.Cluster)).To(Succeed())
		})

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

		Specify("doesn't break when two cross-mesh gateways exist with the same service value", func() {
			const gatewayMesh2 = "cross-mesh-gateway2"
			crossMeshGatewayYaml2 := mkGateway(
				crossMeshGatewayName+"2", crossMeshGatewayName, gatewayMesh2, true, "gateway2.mesh", echoServerService(gatewayMesh2, gatewayTestNamespace), crossMeshGatewayPort,
			)
			crossMeshGatewayInstanceYaml2 := MkGatewayInstance(crossMeshGatewayName, gatewayTestNamespace2, gatewayMesh2)

			setup := NewClusterSetup().
				Install(MTLSMeshKubernetes(gatewayMesh2)).
				Install(YamlK8s(crossMeshGatewayYaml2)).
				Install(YamlK8s(crossMeshGatewayInstanceYaml2))
			Expect(setup.Setup(env.Cluster)).To(Succeed())

			gatewayAddr := net.JoinHostPort(crossMeshHostname, strconv.Itoa(crossMeshGatewayPort))
			Consistently(FailToProxyRequestToGateway(
				env.Cluster,
				gatewayAddr,
				gatewayClientNamespaceOtherMesh,
			), "30s", "1s").ShouldNot(Succeed())

			setup = NewClusterSetup().
				Install(DeleteYamlK8s(crossMeshGatewayYaml2)).
				Install(DeleteYamlK8s(crossMeshGatewayInstanceYaml2))
			Expect(setup.Setup(env.Cluster)).To(Succeed())
			Expect(env.Cluster.DeleteMesh(gatewayMesh2)).To(Succeed())
		})
	})

	Context("with Gateway API", Label("arm-not-supported"), func() {
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
  annotations:
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
  annotations:
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
			Expect(setup.Setup(env.Cluster)).To(Succeed())
		})
		E2EAfterAll(func() {
			setup := NewClusterSetup().
				Install(DeleteYamlK8s(gateway)).
				Install(DeleteYamlK8s(route))
			Expect(setup.Setup(env.Cluster)).To(Succeed())
		})
		It("should proxy HTTP requests from a different mesh", func() {
			gatewayAddr := net.JoinHostPort(crossMeshHostname, strconv.Itoa(crossMeshGatewayPort))
			SuccessfullyProxyRequestToGateway(
				env.Cluster, gatewayMesh,
				gatewayAddr,
				gatewayClientNamespaceOtherMesh,
			)
		})
	})
}
