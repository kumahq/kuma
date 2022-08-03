package gateway

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/kubernetes/env"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func Mtls() {
	meshName := "gateway-mtls"
	namespace := "gateway-mtls"
	clientNamespace := "gateway-mtls-client"

	meshGateway := `
apiVersion: kuma.io/v1alpha1
kind: MeshGateway
metadata:
  name: mtls-edge-gateway
mesh: gateway-mtls
spec:
  selectors:
  - match:
      kuma.io/service: mtls-edge-gateway
  conf:
    listeners:
    - port: 8080
      protocol: HTTP
      hostname: example.kuma.io
      tags:
        hostname: example.kuma.io
    - port: 8081
      protocol: TCP
      tags:
        protocol: tcp
`

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MTLSMeshKubernetes(meshName)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Namespace(clientNamespace)).
			Install(testserver.Install(
				testserver.WithName("demo-client"),
				testserver.WithNamespace(clientNamespace),
			)).
			Install(YamlK8s(meshGateway)).
			Install(YamlK8s(MkGatewayInstance("mtls-edge-gateway", namespace, meshName))).
			Setup(env.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterAll(func() {
		Expect(env.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(env.Cluster.TriggerDeleteNamespace(clientNamespace)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	Context("HTTP", func() {
		meshGatewayRouteHTTP := `
apiVersion: kuma.io/v1alpha1
kind: MeshGatewayRoute
metadata:
  name: mtls-edge-gateway-http
mesh: gateway-mtls
spec:
  selectors:
  - match:
      kuma.io/service: mtls-edge-gateway
      hostname: example.kuma.io
  conf:
    http:
      rules:
      - matches:
        - path:
            match: PREFIX
            value: /prefix/
        filters:
        - requestHeader:
            set:
              - name: Host
                value: other.example.kuma.io
        - rewrite:
            replacePrefixMatch: "/"
        backends:
        - destination:
            kuma.io/service: echo-server_gateway-mtls_svc_80
      - matches:
        - path:
            match: PREFIX
            value: /non-accessible
        backends:
        - destination:
            kuma.io/service: non-accessible-echo-server_gateway-mtls_svc_80
      - matches:
        - path:
            match: PREFIX
            value: /
        backends:
        - destination:
            kuma.io/service: echo-server_gateway-mtls_svc_80
`
		BeforeAll(func() {
			err := NewClusterSetup().
				Install(testserver.Install(
					testserver.WithMesh(meshName),
					testserver.WithName("echo-server"),
					testserver.WithNamespace(namespace),
					testserver.WithEchoArgs("echo", "--instance", "kubernetes"),
				)).
				Install(testserver.Install(
					testserver.WithMesh(meshName),
					testserver.WithName("non-accessible-echo-server"),
					testserver.WithNamespace(namespace),
					testserver.WithEchoArgs("echo", "--instance", "non-accessible-echo-server"),
				)).
				Install(YamlK8s(meshGatewayRouteHTTP)).
				Setup(env.Cluster)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should proxy simple HTTP requests", func() {
			Eventually(func(g Gomega) {
				response, err := client.CollectResponse(
					env.Cluster, "demo-client",
					"http://mtls-edge-gateway.gateway-mtls:8080/",
					client.WithHeader("host", "example.kuma.io"),
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)

				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.Instance).To(Equal("kubernetes"))
			}, "30s", "1s").Should(Succeed())
		})

		It("should rewrite HTTP requests", func() {
			Eventually(func(g Gomega) {
				response, err := client.CollectResponse(
					env.Cluster, "demo-client", "http://mtls-edge-gateway.gateway-mtls:8080/prefix/xyz",
					client.WithHeader("host", "example.kuma.io"),
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)

				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.Received.Headers["Host"]).To(ContainElement("other.example.kuma.io"))
				g.Expect(response.Received.Path).To(Equal("/xyz"))
			}, "30s", "1s").Should(Succeed())
		})

		It("should not access a service for which we don't have traffic permission", func() {
			tp := `
apiVersion: kuma.io/v1alpha1
kind: TrafficPermission
metadata:
  name: tp-non-accessible-echo-server
mesh: gateway-mtls
spec:
  sources:
  - match:
      kuma.io/service: not-mtls-edge-gateway
  destinations:
  - match:
      kuma.io/service: non-accessible-echo-server_gateway-mtls_svc_80
`
			Expect(env.Cluster.Install(YamlK8s(tp))).To(Succeed())

			Eventually(func(g Gomega) {
				status, err := client.CollectFailure(
					env.Cluster, "demo-client", "http://mtls-edge-gateway.gateway-mtls:8080/non-accessible",
					client.WithHeader("host", "example.kuma.io"),
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)

				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(status.ResponseCode).To(Equal(503))
			}, "30s", "1s").Should(Succeed())
		})
	})

	Context("TCP", func() {
		tcpRoute := `
apiVersion: kuma.io/v1alpha1
kind: MeshGatewayRoute
metadata:
  name: mtls-gateway-tcp
mesh: gateway-mtls
spec:
  selectors:
  - match:
      kuma.io/service: mtls-edge-gateway
      protocol: tcp
  conf:
    tcp:
      rules:
      - backends:
        - destination:
            kuma.io/service: tcp-server_gateway-mtls_svc_80
`

		BeforeAll(func() {
			err := NewClusterSetup().
				Install(YamlK8s(tcpRoute)).
				Install(testserver.Install(
					testserver.WithMesh(meshName),
					testserver.WithName("tcp-server"),
					testserver.WithNamespace(namespace),
					testserver.WithHealthCheckTCPArgs("health-check", "tcp", "--port", "80"),
				)).
				Setup(env.Cluster)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should proxy TCP connections", func() {
			Eventually(func(g Gomega) {
				response, err := client.CollectTCPResponse(env.Cluster, "demo-client", "telnet://mtls-edge-gateway.gateway-mtls:8081", "request",
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)

				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response).Should(Equal("response"))
			}, "30s", "1s").Should(Succeed())
		})
	})
}
