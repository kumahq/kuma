package gateway

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/kubernetes/env"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func CrossMeshGatewayOnKubernetes() {
	// ClientNamespace is a namespace to deploy gateway client
	// applications. Mesh sidecar injection is not enabled there.
	const gatewayClientNamespaceOtherMesh = "cross-mesh-kuma-client-other"
	const gatewayClientNamespaceSameMesh = "cross-mesh-kuma-client"
	const gatewayTestNamespace = "cross-mesh-kuma-test"

	const echoServer = "cross-mesh-echo-server"
	const edgeGateway = "cross-mesh-edge-gateway"

	const gatewayMesh = "cross-mesh-gateway"
	const gatewayOtherMesh = "cross-mesh-other"

	const gatewayPort = 8080

	echoServerApp := testserver.Install(
		testserver.WithMesh(gatewayMesh),
		testserver.WithName(echoServer),
		testserver.WithNamespace(gatewayTestNamespace),
		testserver.WithArgs("echo", "--instance", "kubernetes"),
	)

	meshGateway := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshGateway
metadata:
  name: %s
mesh: %s
spec:
  selectors:
  - match:
      kuma.io/service: %s
  conf:
    listeners:
    - port: %d
      protocol: HTTP
      crossMesh: true
      hostname: gateway.mesh
      tags:
        hostname: gateway.mesh
`, edgeGateway, gatewayMesh, edgeGateway, gatewayPort)

	meshGatewayRoute := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshGatewayRoute
metadata:
  name: %s
mesh: %s
spec:
  selectors:
  - match:
      kuma.io/service: %s
  conf:
    http:
      rules:
      - matches:
        - path:
            match: PREFIX
            value: /
        backends:
        - destination:
            kuma.io/service: %s_%s_svc_80 # Matches the echo-server we deployed.
`, edgeGateway, gatewayMesh, edgeGateway, echoServer, gatewayTestNamespace)

	meshGatewayInstance := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshGatewayInstance
metadata:
  name: %s
  namespace: %s
  annotations:
    kuma.io/mesh: %s
spec:
  replicas: 1
  serviceType: ClusterIP
  tags:
    kuma.io/service: %s
`, edgeGateway, gatewayTestNamespace, gatewayMesh, edgeGateway)

	BeforeAll(func() {
		setup := NewClusterSetup().
			Install(MTLSMeshKubernetes(gatewayMesh)).
			Install(MTLSMeshKubernetes(gatewayOtherMesh)).
			Install(NamespaceWithSidecarInjection(gatewayTestNamespace)).
			Install(NamespaceWithSidecarInjection(gatewayClientNamespaceOtherMesh)).
			Install(NamespaceWithSidecarInjection(gatewayClientNamespaceSameMesh)).
			Install(echoServerApp).
			Install(DemoClientK8s(gatewayOtherMesh, gatewayClientNamespaceOtherMesh)).
			Install(DemoClientK8s(gatewayMesh, gatewayClientNamespaceSameMesh)).
			Install(YamlK8s(meshGateway)).
			Install(YamlK8s(meshGatewayRoute)).
			Install(YamlK8s(meshGatewayInstance))

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
		It("should proxy simple HTTP requests from a different mesh", func() {
			proxyRequestToGateway(env.Cluster, "kubernetes",
				"gateway.mesh", gatewayPort,
				client.FromKubernetesPod(gatewayClientNamespaceOtherMesh, "demo-client"))
		})

		It("should proxy simple HTTP requests from the same mesh", func() {
			proxyRequestToGateway(env.Cluster, "kubernetes",
				"gateway.mesh", gatewayPort,
				client.FromKubernetesPod(gatewayClientNamespaceSameMesh, "demo-client"))
		})
	})
}
