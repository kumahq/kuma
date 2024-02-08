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
    - port: 8083
      protocol: TCP
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
	var clusterIP string

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MTLSMeshKubernetes(meshName)).
			Install(Namespace(clientNamespace)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(testserver.Install(
				testserver.WithName("demo-client"),
				testserver.WithNamespace(clientNamespace),
			)).
			Install(testserver.Install(
				testserver.WithName("echo-server"),
				testserver.WithMesh(meshName),
				testserver.WithNamespace(namespace),
				testserver.WithEchoArgs("echo", "--instance", "echo-server"),
			)).
			Install(YamlK8s(httpsSecret())).
			Install(YamlK8s(meshGateway)).
			Install(YamlK8s(MkGatewayInstance("simple-gateway", namespace, meshName))).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())

		Eventually(func(g Gomega) {
			var err error
			clusterIP, err = k8s.RunKubectlAndGetOutputE(
				kubernetes.Cluster.GetTesting(),
				kubernetes.Cluster.GetKubectlOptions(namespace),
				"get", "service", "simple-gateway", "-ojsonpath={.spec.clusterIP}",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(clusterIP).ToNot(BeEmpty())
		}, "30s", "1s").Should(Succeed())
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(clientNamespace)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

<<<<<<< HEAD
	route := func(name, path, destination string) string {
		return fmt.Sprintf(`
=======
	meshGatewayRoutes := func(name, path, destination string) []string {
		return []string{
			fmt.Sprintf(`
>>>>>>> 67ee1be51 (fix(MeshGateway): fix MeshTCPRoute on MeshGateway (#9167))
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
        - rewrite:
            hostToBackendHostname: true
        backends:
        - destination:
            kuma.io/service: %s
`, name, path, destination, path, destination)
	}

<<<<<<< HEAD
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
=======
	httpRoute := func(name, path, destination string) []string {
		return []string{
			fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
metadata:
  name: %s
  namespace: %s
  labels:
    kuma.io/mesh: simple-gateway
spec:
  targetRef:
    kind: MeshGateway
    name: simple-gateway
  to:
  - targetRef:
      kind: Mesh
    rules:
    - matches:
      - path:
          type: PathPrefix
          value: "%s"
      default:
        backendRefs:
        - kind: MeshService
          name: "%s"
  - targetRef:
      kind: Mesh
    hostnames:
    - specific.kuma.io
    rules:
    - matches:
      - path:
          type: PathPrefix
          value: "%s"
      default:
        filters:
        - type: RequestHeaderModifier
          requestHeaderModifier:
            add:
            - name: x-specific-hostname-header
              value: "true"
        backendRefs:
        - kind: MeshService
          name: "%s"
`, name, Config.KumaNamespace, path, destination, path, destination),
			fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
metadata:
  name: %s
  namespace: %s
  labels:
    kuma.io/mesh: simple-gateway
spec:
  targetRef:
    kind: MeshGateway
    name: simple-gateway
    tags:
      hostname: otherexample.kuma.io 
  to:
  - targetRef:
      kind: Mesh
    rules:
    - matches:
      - path:
          type: PathPrefix
          value: "%s-specific-listener"
      default:
        filters:
        - type: RequestHeaderModifier
          requestHeaderModifier:
            add:
            - name: x-listener-by-hostname-header
              value: "true"
        backendRefs:
        - kind: MeshService
          name: "%s"
`, name+"-hostname-specific", Config.KumaNamespace, path, destination),
		}
	}

	basicRouting := func(name string, routeYAMLs []string) {
		Context(fmt.Sprintf("Mesh service - %s", name), func() {
			BeforeAll(func() {
				Expect(NewClusterSetup().
					Install(YamlK8s(routeYAMLs...)).
					Setup(kubernetes.Cluster),
				).To(Succeed())
			})

			E2EAfterAll(func() {
				Expect(NewClusterSetup().
					Install(DeleteYamlK8s(routeYAMLs...)).
					Setup(kubernetes.Cluster),
				).To(Succeed())
			})

			It("should proxy to service via HTTP without port in host", func() {
				Eventually(func(g Gomega) {
					response, err := client.CollectEchoResponse(
						kubernetes.Cluster, "demo-client",
						"http://simple-gateway.simple-gateway:8080/",
						client.WithHeader("host", "example.kuma.io"),
						client.FromKubernetesPod(clientNamespace, "demo-client"),
					)

					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(response.Instance).To(Equal("echo-server"))
					g.Expect(response.Received.Headers).ToNot(HaveKey("X-Specific-Hostname-Header"))
				}, "1m", "1s").Should(Succeed())
			})
			It("should proxy to service via HTTP with port in host", func() {
				Eventually(func(g Gomega) {
					response, err := client.CollectEchoResponse(
						kubernetes.Cluster, "demo-client",
						"http://simple-gateway.simple-gateway:8080/",
						client.WithHeader("host", "example.kuma.io:8080"),
						client.FromKubernetesPod(clientNamespace, "demo-client"),
					)

					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(response.Instance).To(Equal("echo-server"))
					g.Expect(response.Received.Headers).ToNot(HaveKey("X-Specific-Hostname-Header"))
				}, "30s", "1s").Should(Succeed())
			})

			It("should proxy to service via HTTPS", func() {
				Eventually(func(g Gomega) {
					response, err := client.CollectEchoResponse(
						kubernetes.Cluster, "demo-client",
						"https://example.kuma.io:8081/",
						client.Resolve("example.kuma.io:8081", clusterIP),
						client.FromKubernetesPod(clientNamespace, "demo-client"),
						client.Insecure(),
					)

					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(response.Instance).To(Equal("echo-server"))
					g.Expect(response.Received.Headers).ToNot(HaveKey("X-Specific-Hostname-Header"))
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
					g.Expect(response.Received.Headers).ToNot(HaveKey("X-Specific-Hostname-Header"))
				}, "30s", "1s").Should(Succeed())
			})

			It("should route based on host name", func() {
				Eventually(func(g Gomega) {
					response, err := client.CollectEchoResponse(
						kubernetes.Cluster, "demo-client",
						"http://simple-gateway.simple-gateway:8082/",
						client.WithHeader("host", "specific.kuma.io"),
						client.FromKubernetesPod(clientNamespace, "demo-client"),
					)

					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(response.Received.Headers["Host"]).To(HaveLen(1))
					g.Expect(response.Received.Headers["Host"]).To(ContainElements("specific.kuma.io"))
					g.Expect(response.Received.Headers["X-Specific-Hostname-Header"]).To(ContainElements("true"))
				}, "30s", "1s").Should(Succeed())
			})

			It("should match routes by SNI", func() {
				Eventually(func(g Gomega) {
					response, err := client.CollectEchoResponse(
						kubernetes.Cluster, "demo-client",
						"https://otherexample.kuma.io:8081/-specific-listener",
						client.Resolve("otherexample.kuma.io:8081", clusterIP),
						client.FromKubernetesPod(clientNamespace, "demo-client"),
						client.Insecure(),
					)

					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(response.Received.Headers["Host"]).To(HaveLen(1))
					g.Expect(response.Received.Headers["Host"]).To(ContainElements("otherexample.kuma.io:8081"))
					g.Expect(response.Received.Headers["X-Listener-By-Hostname-Header"]).To(ContainElements("true"))
				}, "30s", "1s").Should(Succeed())
			})

			It("should isolate routes by SNI", func() {
				Eventually(func(g Gomega) {
					response, err := client.CollectEchoResponse(
						kubernetes.Cluster, "demo-client",
						"https://example.kuma.io:8081/-specific-listener",
						client.Resolve("example.kuma.io:8081", clusterIP),
						client.FromKubernetesPod(clientNamespace, "demo-client"),
						client.Insecure(),
					)

					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(response.Received.Headers["Host"]).To(HaveLen(1))
					g.Expect(response.Received.Headers["Host"]).To(ContainElements("example.kuma.io:8081"))
					g.Expect(response.Received.Headers["X-Listener-By-Hostname-Header"]).NotTo(ContainElements("true"))
				}, "30s", "1s").Should(Succeed())
			})

			It("should check both SNI and Host", func() {
				Consistently(func(g Gomega) {
					status, err := client.CollectFailure(
						kubernetes.Cluster, "demo-client",
						"https://otherexample.kuma.io:8081/-specific-listener",
						client.Resolve("otherexample.kuma.io:8081", clusterIP),
						// Note the header differs from the SNI
						client.WithHeader("host", "example.kuma.io"),
						client.FromKubernetesPod(clientNamespace, "demo-client"),
						client.Insecure(),
					)

					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(status.ResponseCode).To(Equal(404))
				}, "30s", "1s").Should(Succeed())
			})
>>>>>>> 67ee1be51 (fix(MeshGateway): fix MeshTCPRoute on MeshGateway (#9167))
		})

<<<<<<< HEAD
		It("should proxy to service via HTTP without port in host", func() {
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
		It("should proxy to service via HTTP with port in host", func() {
			Eventually(func(g Gomega) {
				response, err := client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client",
					"http://simple-gateway.simple-gateway:8080/",
					client.WithHeader("host", "example.kuma.io:8080"),
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
=======
	basicRouting("MeshGatewayRoute", meshGatewayRoutes("internal-service", "/", "echo-server_simple-gateway_svc_80"))
	basicRouting("MeshHTTPRoute", httpRoute("internal-service", "/", "echo-server_simple-gateway_svc_80"))
>>>>>>> 67ee1be51 (fix(MeshGateway): fix MeshTCPRoute on MeshGateway (#9167))

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
		routes := meshGatewayRoutes("rt-echo-server", "/rt", "rt-echo-server_simple-gateway_svc_80")

		BeforeAll(func() {
			err := NewClusterSetup().
				Install(testserver.Install(
					testserver.WithName("rt-echo-server"),
					testserver.WithMesh(meshName),
					testserver.WithNamespace(namespace),
					testserver.WithEchoArgs("echo", "--instance", "rt-echo-server"),
				)).
				Install(YamlK8s(rt)).
<<<<<<< HEAD
				Install(YamlK8s(route("rt-echo-server", "/rt", "rt-echo-server_simple-gateway_svc_80"))).
=======
				Install(YamlK8s(routes...)).
				Setup(kubernetes.Cluster)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterAll(func() {
			err := NewClusterSetup().
				Install(DeleteYamlK8s(routes...)).
>>>>>>> 67ee1be51 (fix(MeshGateway): fix MeshTCPRoute on MeshGateway (#9167))
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

	Context("TCPRoute", func() {
		routes := []string{
			fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTCPRoute
metadata:
  name: simple-gateway-tcp-route
  namespace: %s
  labels:
    kuma.io/mesh: simple-gateway
spec:
  targetRef:
    kind: MeshGateway
    name: simple-gateway
  to:
    - targetRef:
        kind: Mesh
      rules:
        - default:
            backendRefs:
              - kind: MeshService
                name: test-tcp-server_simple-gateway_svc_80 
`, Config.KumaNamespace),
		}
		BeforeAll(func() {
			err := NewClusterSetup().
				Install(testserver.Install(
					testserver.WithName("test-tcp-server"),
					testserver.WithServicePortAppProtocol("tcp"),
					testserver.WithMesh(meshName),
					testserver.WithNamespace(namespace),
				)).
				Install(YamlK8s(routes...)).
				Setup(kubernetes.Cluster)
			Expect(err).ToNot(HaveOccurred())
		})
		AfterAll(func() {
			err := NewClusterSetup().
				Install(DeleteYamlK8s(routes...)).
				Setup(kubernetes.Cluster)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should work on TCP service", func() {
			Eventually(func(g Gomega) {
				response, err := client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client",
					"http://simple-gateway.simple-gateway:8083/",
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.Instance).To(HavePrefix("test-tcp-server"))
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
<<<<<<< HEAD
			route := route("es-echo-server", "/external-service", "external-service")
			setup := NewClusterSetup().Install(YamlK8s(route))
=======
			route := meshGatewayRoutes("es-echo-server", "/external-service", "external-service")
			setup := NewClusterSetup().Install(YamlK8s(route...))
>>>>>>> 67ee1be51 (fix(MeshGateway): fix MeshTCPRoute on MeshGateway (#9167))
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
<<<<<<< HEAD
			route := route("es-echo-server", "/external-service", "external-service")
			setup := NewClusterSetup().Install(YamlK8s(route))
=======
			route := meshGatewayRoutes("es-echo-server", "/external-service", "external-service")
			setup := NewClusterSetup().Install(YamlK8s(route...))
>>>>>>> 67ee1be51 (fix(MeshGateway): fix MeshTCPRoute on MeshGateway (#9167))
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
