package gateway

import (
	"encoding/base64"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func Gateway() {
	meshName := "simple-gateway"
	namespace := "simple-gateway"
	clientNamespace := "client-simple-gateway"

	meshGateway := `
apiVersion: kuma.io/v1alpha1
kind: MeshGateway
metadata:
  name: simple-gateway
mesh: simple-gateway
spec:
  selectors:
  - match:
      kuma.io/service: simple-gateway
  conf:
    listeners:
    - port: 8080
      protocol: HTTP
      hostname: example.kuma.io
      tags:
        hostname: example.kuma.io
    - port: 8081
      protocol: HTTPS
      tls:
        mode: TERMINATE
        certificates:
        - secret: example-kuma-io-certificate
      tags:
        hostname: example.kuma.io
    - port: 8082
      protocol: HTTP
      hostname: '*'
`
	httpsSecret := func() string {
		cert, key, err := CreateCertsFor("example.kuma.io")
		Expect(err).To(Succeed())
		secretData := base64.StdEncoding.EncodeToString([]byte(strings.Join([]string{key, cert}, "\n")))
		return fmt.Sprintf(`
apiVersion: v1
kind: Secret
metadata:
  name: example-kuma-io-certificate
  namespace: %s
  labels:
    kuma.io/mesh: simple-gateway 
data:
  value: %s
type: system.kuma.io/secret
`, Config.KumaNamespace, secretData)
	}

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MTLSMeshKubernetes(meshName)).
			Install(Namespace(clientNamespace)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(testserver.Install(
				testserver.WithName("demo-client"),
				testserver.WithNamespace(clientNamespace),
			)).
			Install(YamlK8s(httpsSecret())).
			Install(YamlK8s(meshGateway)).
			Install(YamlK8s(MkGatewayInstance("simple-gateway", namespace, meshName))).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(clientNamespace)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	route := func(name, path, destination string) string {
		return fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshGatewayRoute
metadata:
  name: %s
mesh: simple-gateway
spec:
  selectors:
  - match:
      kuma.io/service: simple-gateway
  conf:
    http:
      rules:
      - matches:
        - path:
            match: PREFIX
            value: %s
        backends:
        - destination:
            kuma.io/service: %s
      - matches:
        - path:
            match: PREFIX
            value: /rewrite-host%s
        filters:
        - rewrite
            hostToBackendHostname: true
        backends:
        - destination:
            kuma.io/service: %s
`, name, path, destination, path, destination)
	}

	Context("Mesh service", func() {
		BeforeAll(func() {
			err := NewClusterSetup().
				Install(testserver.Install(
					testserver.WithName("echo-server"),
					testserver.WithMesh(meshName),
					testserver.WithNamespace(namespace),
					testserver.WithEchoArgs("echo", "--instance", "echo-server"),
				)).
				Install(YamlK8s(route("internal-service", "/", "echo-server_simple-gateway_svc_80"))).
				Setup(kubernetes.Cluster)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should proxy to service via HTTP", func() {
			Eventually(func(g Gomega) {
				response, err := client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client",
					"http://simple-gateway.simple-gateway:8080/",
					client.WithHeader("host", "example.kuma.io"),
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)

				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.Instance).To(Equal("echo-server"))
			}, "30s", "1s").Should(Succeed())
		})

		It("should proxy to service via HTTPS", func() {
			Eventually(func(g Gomega) {
				response, err := client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client",
					"https://simple-gateway.simple-gateway:8081/",
					client.FromKubernetesPod(clientNamespace, "demo-client"),
					client.Insecure(),
				)

				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.Instance).To(Equal("echo-server"))
			}, "30s", "1s").Should(Succeed())
		})

		It("should automatically set host header from service address", func() {
			Eventually(func(g Gomega) {
				response, err := client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client",
					"http://simple-gateway.simple-gateway:8082/",
					client.WithHeader("host", "example.kuma.io"),
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)

				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.Received.Headers["Host"]).To(HaveLen(1))
				g.Expect(response.Received.Headers["Host"]).To(ContainElements("example.kuma.io"))
			}, "30s", "1s").Should(Succeed())
		})
	})

	Context("Rate Limit", func() {
		rt := `apiVersion: kuma.io/v1alpha1
kind: RateLimit
metadata:
  name: gateway-rate-limit
mesh: simple-gateway
spec:
  sources:
  - match:
      kuma.io/service: simple-gateway
  destinations:
  - match:
      kuma.io/service: rt-echo-server_simple-gateway_svc_80
  conf:
    http:
      requests: 5
      interval: 10s`

		BeforeAll(func() {
			err := NewClusterSetup().
				Install(testserver.Install(
					testserver.WithName("rt-echo-server"),
					testserver.WithMesh(meshName),
					testserver.WithNamespace(namespace),
					testserver.WithEchoArgs("echo", "--instance", "rt-echo-server"),
				)).
				Install(YamlK8s(rt)).
				Install(YamlK8s(route("rt-echo-server", "/rt", "rt-echo-server_simple-gateway_svc_80"))).
				Setup(kubernetes.Cluster)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should rate limit", func() {
			Eventually(func(g Gomega) {
				response, err := client.CollectFailure(
					kubernetes.Cluster, "demo-client",
					"http://simple-gateway.simple-gateway:8080/rt",
					client.WithHeader("host", "example.kuma.io"),
					client.FromKubernetesPod(clientNamespace, "demo-client"),
					client.NoFail(),
					client.OutputFormat(`{ "received": { "status": %{response_code} } }`),
				)

				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.ResponseCode).To(Equal(429))
			}, "30s", "1s").Should(Succeed())
		})
	})

	Context("External Service", func() {
		externalService := `
apiVersion: kuma.io/v1alpha1
kind: ExternalService
mesh: simple-gateway
metadata:
  name: simple-gateway-external-service
spec:
  tags:
    kuma.io/service: external-service
    kuma.io/protocol: http
  networking:
    address: es-echo-server.client-simple-gateway.svc.cluster.local:80`

		BeforeAll(func() {
			err := NewClusterSetup().
				Install(testserver.Install(
					testserver.WithName("es-echo-server"),
					testserver.WithNamespace(clientNamespace),
					testserver.WithEchoArgs("echo", "--instance", "es-echo-server"),
				)).
				Install(YamlK8s(externalService)).
				Setup(kubernetes.Cluster)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should proxy to service via HTTP", func() {
			route := route("es-echo-server", "/external-service", "external-service")
			setup := NewClusterSetup().Install(YamlK8s(route))
			Expect(setup.Setup(kubernetes.Cluster)).To(Succeed())

			Eventually(func(g Gomega) {
				response, err := client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client",
					"http://simple-gateway.simple-gateway:8080/external-service",
					client.WithHeader("host", "example.kuma.io"),
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)

				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.Instance).To(Equal("es-echo-server"))
			}, "30s", "1s").Should(Succeed())

			Expect(NewClusterSetup().Install(DeleteYamlK8s(route)).Setup(kubernetes.Cluster)).To(Succeed())
		})

		It("should automatically set host header from external service address when rewrite.hostToBackendHostname is set to true", func() {
			route := route("es-echo-server", "/external-service", "external-service")
			setup := NewClusterSetup().Install(YamlK8s(route))
			Expect(setup.Setup(kubernetes.Cluster)).To(Succeed())

			// don't rewrite host to backend hostname
			Eventually(func(g Gomega) {
				response, err := client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client",
					"http://simple-gateway.simple-gateway:8080/external-service",
					client.WithHeader("host", "example.kuma.io"),
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)

				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.Received.Headers["Host"]).To(HaveLen(1))
				g.Expect(response.Received.Headers["Host"]).To(ContainElements("example.kuma.io"))
			}, "30s", "1s").Should(Succeed())

			// rewrite host to backend hostname
			Eventually(func(g Gomega) {
				response, err := client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client",
					"http://simple-gateway.simple-gateway:8080/rewrite-host/external-service",
					client.WithHeader("host", "example.kuma.io"),
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)

				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.Received.Headers["Host"]).To(HaveLen(1))
				g.Expect(response.Received.Headers["Host"]).To(ContainElements("es-echo-server.client-simple-gateway.svc.cluster.local"))
			}, "30s", "1s").Should(Succeed())

			Expect(NewClusterSetup().Install(DeleteYamlK8s(route)).Setup(kubernetes.Cluster)).To(Succeed())
		})

		It("should handle external-service excluded by tags", func() {
			route := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshGatewayRoute
metadata:
  name: %s
mesh: simple-gateway
spec:
  selectors:
  - match:
      kuma.io/service: simple-gateway
  conf:
    http:
      rules:
      - matches:
        - path:
            match: PREFIX
            value: %s
        backends:
        - destination:
            kuma.io/service: %s
            nonexistent: tag
`, "es-echo-server-broken", "/external-service", "external-service")

			setup := NewClusterSetup().Install(YamlK8s(route))
			Expect(setup.Setup(kubernetes.Cluster)).To(Succeed())

			Eventually(func(g Gomega) {
				response, err := client.CollectFailure(
					kubernetes.Cluster, "demo-client",
					"http://simple-gateway.simple-gateway:8080/external-service",
					client.WithHeader("host", "example.kuma.io"),
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)

				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.ResponseCode).To(Equal(503))
			}, "30s", "1s").Should(Succeed())

			Expect(NewClusterSetup().Install(DeleteYamlK8s(route)).Setup(kubernetes.Cluster)).To(Succeed())
		})
	})
}
