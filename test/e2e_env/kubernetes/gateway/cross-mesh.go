package gateway

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
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
	const gatewayClientNamespace = "cross-mesh-kuma-client"
	const gatewayClientNamespaceOtherMesh = "cross-mesh-kuma-client2"
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

	// DeployCluster creates a Kuma cluster on Kubernetes using the
	// provided options, installing an echo service as well as a
	// gateway and a client container to send HTTP requests.
	DeployCluster := func() {
		setup := NewClusterSetup().
			Install(MTLSMeshKubernetes(gatewayMesh)).
			Install(MTLSMeshKubernetes(gatewayOtherMesh)).
			Install(NamespaceWithSidecarInjection(gatewayTestNamespace)).
			Install(NamespaceWithSidecarInjection(gatewayClientNamespace)).
			Install(NamespaceWithSidecarInjection(gatewayClientNamespaceOtherMesh)).
			Install(echoServerApp).
			Install(DemoClientK8s(gatewayOtherMesh, gatewayClientNamespace)).
			Install(DemoClientK8s(gatewayMesh, gatewayClientNamespaceOtherMesh))

		Expect(setup.Setup(env.Cluster)).To(Succeed())
	}

	// Before each test, install the gateway and routes.
	JustBeforeEach(func() {
		Expect(k8s.KubectlApplyFromStringE(
			env.Cluster.GetTesting(),
			env.Cluster.GetKubectlOptions(gatewayTestNamespace), fmt.Sprintf(`
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
`, edgeGateway, gatewayMesh, edgeGateway, gatewayPort)),
		).To(Succeed())

		Expect(
			env.Cluster.GetKumactlOptions().KumactlList("meshgateways", gatewayMesh),
		).To(ContainElement(edgeGateway))

		Expect(k8s.KubectlApplyFromStringE(
			env.Cluster.GetTesting(),
			env.Cluster.GetKubectlOptions(gatewayTestNamespace), fmt.Sprintf(`
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
`, edgeGateway, gatewayMesh, edgeGateway, echoServer, gatewayTestNamespace)),
		).To(Succeed())

		Expect(
			env.Cluster.GetKumactlOptions().KumactlList("meshgatewayroutes", gatewayMesh),
		).To(ContainElement(edgeGateway))
	})

	// Before each test, install the GatewayInstance to provision
	// dataplanes. Note that we expose the gateway inside the cluster
	// with a ClusterIP service. This makes it easier for the tests
	// to figure out the IP address to send requests to, since the
	// alternatives are a load balancer (we don't have one) or node port
	// (would need to inspect nodes).
	JustBeforeEach(func() {
		Expect(k8s.KubectlApplyFromStringE(
			env.Cluster.GetTesting(),
			env.Cluster.GetKubectlOptions(gatewayTestNamespace), fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshGatewayInstance
metadata:
  name: %s
  annotations:
    kuma.io/mesh: %s
spec:
  replicas: 1
  serviceType: ClusterIP
  tags:
    kuma.io/service: %s
`, edgeGateway, gatewayMesh, edgeGateway)),
		).To(Succeed())
	})

	// Before each test, verify that we have the Dataplanes that we expect to need.
	JustBeforeEach(func() {
		Expect(env.Cluster.VerifyKuma()).To(Succeed())

		// Synchronize on the dataplanes coming up.
		Eventually(func(g Gomega) {
			dataplanes, err := env.Cluster.GetKumactlOptions().KumactlList("dataplanes", gatewayMesh)
			g.Expect(err).ToNot(HaveOccurred())
			// Dataplane names are generated, so we check for a partial match.
			g.Expect(dataplanes).Should(ContainElement(ContainSubstring(echoServer)))
			g.Expect(dataplanes).Should(ContainElement(ContainSubstring(edgeGateway)))
		}, "60s", "1s").Should(Succeed())
	})

	E2EAfterAll(func() {
		Expect(env.Cluster.TriggerDeleteNamespace(gatewayClientNamespace)).To(Succeed())
		Expect(env.Cluster.TriggerDeleteNamespace(gatewayClientNamespaceOtherMesh)).To(Succeed())
		Expect(env.Cluster.TriggerDeleteNamespace(gatewayTestNamespace)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(gatewayMesh)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(gatewayOtherMesh)).To(Succeed())
	})

	Context("when mTLS is enabled", func() {
		BeforeEach(func() {
			DeployCluster()
		})

		It("should proxy simple HTTP requests from a different mesh", func() {
			proxyRequestToGateway(env.Cluster, "kubernetes",
				"gateway.mesh", gatewayPort,
				client.FromKubernetesPod(gatewayClientNamespace, "demo-client"))
		})

		It("should proxy simple HTTP requests from the same mesh", func() {
			proxyRequestToGateway(env.Cluster, "kubernetes",
				"gateway.mesh", gatewayPort,
				client.FromKubernetesPod(gatewayClientNamespaceOtherMesh, "demo-client"))
		})
	})
}
