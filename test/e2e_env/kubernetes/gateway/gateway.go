package gateway

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	meshretry_api "github.com/kumahq/kuma/pkg/plugins/policies/meshretry/api/v1alpha1"
	"github.com/kumahq/kuma/test/framework"
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
      kuma.io/service: simple-gateway_simple-gateway_svc
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
        - secret: kuma-io-certificate-k8s` +
		// secret names have to be unique because
		// we're removing secrets using owner reference, and we're relying on async namespace deletion,
		// so we could have a situation where the secrets are not yet deleted,
		// and we're trying to create a new mesh with the same secret name which k8s treats as changing the mesh of the secret
		// and results in "cannot change mesh of the Secret. Delete the Secret first and apply it again."
		// https://app.circleci.com/pipelines/github/kumahq/kuma/24848/workflows/f33edb5a-74cb-45ae-b0c2-4b14bf289585/jobs/497239
		`
      hostname: example.kuma.io
      tags:
        hostname: example.kuma.io
    - port: 8081
      protocol: HTTPS
      tls:
        mode: TERMINATE
        certificates:
        - secret: kuma-io-certificate-k8s` +
		// secret names have to be unique because
		// we're removing secrets using owner reference, and we're relying on async namespace deletion,
		// so we could have a situation where the secrets are not yet deleted,
		// and we're trying to create a new mesh with the same secret name which k8s treats as changing the mesh of the secret
		// and results in "cannot change mesh of the Secret. Delete the Secret first and apply it again."
		// https://app.circleci.com/pipelines/github/kumahq/kuma/24848/workflows/f33edb5a-74cb-45ae-b0c2-4b14bf289585/jobs/497239
		`
      hostname: otherexample.kuma.io
      tags:
        hostname: otherexample.kuma.io
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
  name: kuma-io-certificate-k8s
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
			Install(YamlK8s(MkGatewayInstanceNoServiceTag("simple-gateway", namespace, meshName))).
			Install(MeshTrafficPermissionAllowAllKubernetes(meshName)).
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

	AfterEachFailure(func() {
		DebugKube(kubernetes.Cluster, meshName, namespace, clientNamespace)
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(clientNamespace)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	meshGatewayRoutes := func(name, path, destination string) []string {
		return []string{
			fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshGatewayRoute
metadata:
  name: %s
mesh: simple-gateway
spec:
  selectors:
  - match:
      kuma.io/service: simple-gateway_simple-gateway_svc
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
`, name, path, destination, path, destination),
			fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshGatewayRoute
metadata:
  name: %s-specific
mesh: simple-gateway
spec:
  selectors:
  - match:
      kuma.io/service: simple-gateway_simple-gateway_svc
  conf:
    http:
      hostnames:
      - specific.kuma.io
      rules:
      - matches:
        - path:
            match: PREFIX
            value: %s
        filters:
        - requestHeader:
            add:
            - name: x-specific-hostname-header
              value: "true"
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
`, name, path, destination, path, destination),
			fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshGatewayRoute
metadata:
  name: %s-specific-listener
mesh: simple-gateway
spec:
  selectors:
  - match:
      kuma.io/service: simple-gateway_simple-gateway_svc
      hostname: otherexample.kuma.io
  conf:
    http:
      rules:
      - matches:
        - path:
            match: PREFIX
            value: "%s-specific-listener"
        filters:
        - requestHeader:
            add:
            - name: x-listener-by-hostname-header
              value: "true"
        backends:
        - destination:
            kuma.io/service: %s
`, name, path, destination),
		}
	}

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
		})
	}

	basicRouting("MeshGatewayRoute", meshGatewayRoutes("internal-service", "/", "echo-server_simple-gateway_svc_80"))
	basicRouting("MeshHTTPRoute", httpRoute("internal-service", "/", "echo-server_simple-gateway_svc_80"))

	Context("Rate Limit", func() {
		rt := `apiVersion: kuma.io/v1alpha1
kind: RateLimit
metadata:
  name: gateway-rate-limit
mesh: simple-gateway
spec:
  sources:
  - match:
      kuma.io/service: simple-gateway_simple-gateway_svc
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

	Context("MeshLoadBalancingStrategy", func() {
		mlbs := fmt.Sprintf(`
kind: MeshLoadBalancingStrategy
apiVersion: kuma.io/v1alpha1
metadata:
  name: ring-hash
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: MeshGateway
    name: simple-gateway
  to:
    - targetRef:
        kind: MeshService
        name: test-server-mlbs_%[2]s_svc_80
      default:
        loadBalancer:
          type: RingHash
          ringHash:
            hashPolicies:
              - type: Header
                header:
                  name: x-header
`, Config.KumaNamespace, meshName)
		routes := httpRoute("test-server-mlbs", "/mlbs", "test-server-mlbs_simple-gateway_svc_80")

		BeforeAll(func() {
			testServerApp := "test-server-mlbs"
			err := NewClusterSetup().
				Install(testserver.Install(
					testserver.WithMesh(meshName),
					testserver.WithNamespace(namespace),
					testserver.WithName(testServerApp),
					testserver.WithStatefulSet(true),
					testserver.WithReplicas(3),
				)).
				Install(YamlK8s(routes...)).
				Setup(kubernetes.Cluster)
			Expect(err).ToNot(HaveOccurred())

			for _, fn := range []InstallFunc{
				WaitNumPods(namespace, 3, testServerApp),
				WaitPodsAvailable(namespace, testServerApp)} {
				Expect(fn(kubernetes.Cluster)).To(Succeed())
			}
		})

		AfterAll(func() {
			err := NewClusterSetup().
				Install(DeleteYamlK8s(routes...)).
				Install(DeleteYamlK8s(mlbs)).
				Setup(kubernetes.Cluster)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should route to only one instance", func() {
			Eventually(func(g Gomega) {
				responses, err := client.CollectResponsesByInstance(
					kubernetes.Cluster, "demo-client",
					"http://simple-gateway.simple-gateway:8080/mlbs",
					client.WithHeader("host", "example.kuma.io"),
					client.FromKubernetesPod(clientNamespace, "demo-client"),
					client.WithHeader("x-header", "value"),
				)

				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(responses).To(HaveLen(3))
			}, "30s", "1s").Should(Succeed())

			Expect(framework.YamlK8s(mlbs)(kubernetes.Cluster)).To(Succeed())

			Eventually(func(g Gomega) {
				responses, err := client.CollectResponsesByInstance(
					kubernetes.Cluster, "demo-client",
					"http://simple-gateway.simple-gateway:8080/mlbs",
					client.WithHeader("host", "example.kuma.io"),
					client.FromKubernetesPod(clientNamespace, "demo-client"),
					client.WithHeader("x-header", "value"),
				)

				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(responses).To(HaveLen(1))
			}, "30s", "1s").MustPassRepeatedly(5).Should(Succeed())
		})
	})

	Context("MeshRateLimit per route", func() {
		httpRoute := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
metadata:
  name: http-route-1
  namespace: %s
  labels:
    kuma.io/mesh: %s
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
          value: "/rate-limited"
      default:
        backendRefs:
        - kind: MeshService
          name: "echo-server_simple-gateway_svc_80"`, Config.KumaNamespace, meshName)
		httpRoute2 := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
metadata:
  name: http-route-2
  namespace: %s
  labels:
    kuma.io/mesh: %s
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
          value: "/non-rate-limited"
      default:
        backendRefs:
        - kind: MeshService
          name: "echo-server_simple-gateway_svc_80"`, Config.KumaNamespace, meshName)
		mrl := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshRateLimit
metadata:
  name: mesh-rate-limit-1
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: MeshHTTPRoute
    name: http-route-1
  to:
    - targetRef:
        kind: Mesh
      default:
        local:
          http:
            requestRate:
              num: 1
              interval: 10s
            onRateLimit:
              status: 428
              headers:
                add:
                - name: "x-kuma-rate-limited"
                  value: "true"`, Config.KumaNamespace, meshName)

		BeforeAll(func() {
			err := NewClusterSetup().
				Install(YamlK8s(httpRoute, httpRoute2)).
				Install(YamlK8s(mrl)).
				Setup(kubernetes.Cluster)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterAll(func() {
			err := NewClusterSetup().
				Install(DeleteYamlK8s(mrl)).
				Install(DeleteYamlK8s(httpRoute, httpRoute2)).
				Setup(kubernetes.Cluster)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should rate limit only specific path", func() {
			Eventually(func(g Gomega) {
				response, err := client.CollectFailure(
					kubernetes.Cluster, "demo-client",
					"http://simple-gateway.simple-gateway:8080/rate-limited",
					client.WithHeader("host", "example.kuma.io"),
					client.FromKubernetesPod(clientNamespace, "demo-client"),
					client.NoFail(),
					client.OutputFormat(`{ "received": { "status": %{response_code} } }`),
				)

				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.ResponseCode).To(Equal(428))
			}, "30s", "1s", MustPassRepeatedly(5)).Should(Succeed())
			Consistently(func(g Gomega) {
				_, err := client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client",
					"http://simple-gateway.simple-gateway:8080/non-rate-limited",
					client.WithHeader("host", "example.kuma.io"),
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
			}, "30s", "1s", MustPassRepeatedly(5)).Should(Succeed())
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

	Context("MeshRetry per route", func() {
		httpRoute := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
metadata:
  name: http-route-1
  namespace: %s
  labels:
    kuma.io/mesh: %s
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
          value: "/with-retry"
      default:
        backendRefs:
        - kind: MeshService
          name: "echo-server_simple-gateway_svc_80"`, Config.KumaNamespace, meshName)
		httpRoute2 := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
metadata:
  name: http-route-2
  namespace: %s
  labels:
    kuma.io/mesh: %s
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
          value: "/no-retry"
      default:
        backendRefs:
        - kind: MeshService
          name: "echo-server_simple-gateway_svc_80"`, Config.KumaNamespace, meshName)
		meshFaultInjection := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshFaultInjection
metadata:
  namespace: %s
  name: mesh-fault-injecton
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: MeshService
    name: echo-server_simple-gateway_svc_80
  from:
    - targetRef:
        kind: Mesh
      default:
        http:
          - abort:
              httpStatus: 500
              percentage: "50.0"
`, Config.KumaNamespace, meshName)
		mr := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshRetry
metadata:
  name: mesh-retry-1
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: MeshHTTPRoute
    name: http-route-1
  to:
    - targetRef:
        kind: Mesh
      default:
        http:
          numRetries: 5
          retryOn:
            - "5xx"`, Config.KumaNamespace, meshName)

		BeforeAll(func() {
			// remove default MeshRetry
			Expect(DeleteMeshResources(kubernetes.Cluster, meshName, meshretry_api.MeshRetryResourceTypeDescriptor)).To(Succeed())

			err := NewClusterSetup().
				Install(YamlK8s(httpRoute, httpRoute2)).
				Install(YamlK8s(mr)).
				Install(YamlK8s(meshFaultInjection)).
				Setup(kubernetes.Cluster)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterAll(func() {
			err := NewClusterSetup().
				Install(DeleteYamlK8s(mr)).
				Install(DeleteYamlK8s(httpRoute, httpRoute2)).
				Install(DeleteYamlK8s(meshFaultInjection)).
				Setup(kubernetes.Cluster)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should retry only specific path", func() {
			Eventually(func(g Gomega) {
				response, err := client.CollectFailure(
					kubernetes.Cluster, "demo-client",
					"http://simple-gateway.simple-gateway:8080/no-retry",
					client.WithHeader("host", "example.kuma.io"),
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)

				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.ResponseCode).To(Equal(500))
			}, "30s", "1s").Should(Succeed())

			Eventually(func(g Gomega) {
				_, err := client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client",
					"http://simple-gateway.simple-gateway:8080/with-retry",
					client.WithHeader("host", "example.kuma.io"),
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
			}, "30s", "1s", MustPassRepeatedly(5)).Should(Succeed())
		})
	})

	Context("MeshTimeout per route", func() {
		httpRoute := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
metadata:
  name: http-route-1
  namespace: %s
  labels:
    kuma.io/mesh: %s
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
          value: "/timeout"
      default:
        backendRefs:
        - kind: MeshService
          name: "echo-server_simple-gateway_svc_80"`, Config.KumaNamespace, meshName)
		httpRoute2 := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
metadata:
  name: http-route-2
  namespace: %s
  labels:
    kuma.io/mesh: %s
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
          value: "/no-timeout"
      default:
        backendRefs:
        - kind: MeshService
          name: "echo-server_simple-gateway_svc_80"`, Config.KumaNamespace, meshName)
		mt := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: mesh-timeout-1
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: MeshHTTPRoute
    name: http-route-1
  to:
    - targetRef:
        kind: Mesh
      default:
        http:
          requestTimeout: 2s
          streamIdleTimeout: 10s`, Config.KumaNamespace, meshName)

		BeforeAll(func() {
			err := NewClusterSetup().
				Install(YamlK8s(httpRoute, httpRoute2)).
				Install(YamlK8s(mt)).
				Setup(kubernetes.Cluster)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterAll(func() {
			err := NewClusterSetup().
				Install(DeleteYamlK8s(mt)).
				Install(DeleteYamlK8s(httpRoute, httpRoute2)).
				Setup(kubernetes.Cluster)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should timeout only on one route", func() {
			// timeout on specific route
			Eventually(func(g Gomega) {
				response, err := client.CollectFailure(
					kubernetes.Cluster, "demo-client", "http://simple-gateway.simple-gateway:8080/timeout",
					client.FromKubernetesPod(clientNamespace, "demo-client"),
					client.WithHeader("host", "example.kuma.io"),
					client.WithHeader("x-set-response-delay-ms", "3000"), // delay response by 3 seconds
				)
				g.Expect(err).ToNot(HaveOccurred())
				// check that request timed out
				g.Expect(response.ResponseCode).To(Equal(504))
			}, "1m", "1s", MustPassRepeatedly(5)).Should(Succeed())

			// no timeout on another route
			Eventually(func(g Gomega) {
				start := time.Now()
				_, err := client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client", "http://simple-gateway.simple-gateway:8080/no-timeout",
					client.FromKubernetesPod(clientNamespace, "demo-client"),
					client.WithHeader("host", "example.kuma.io"),
					client.WithHeader("x-set-response-delay-ms", "3000"), // delay response by 3 seconds
				)
				g.Expect(err).ToNot(HaveOccurred())
				// check that resonse was received after more than 3 seconds
				g.Expect(time.Since(start)).To(BeNumerically(">", 3*time.Second))
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
			route := meshGatewayRoutes("es-echo-server", "/external-service", "external-service")
			setup := NewClusterSetup().Install(YamlK8s(route...))
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

			Expect(NewClusterSetup().Install(DeleteYamlK8s(route...)).Setup(kubernetes.Cluster)).To(Succeed())
		})

		It("should automatically set host header from external service address when rewrite.hostToBackendHostname is set to true", func() {
			route := meshGatewayRoutes("es-echo-server", "/external-service", "external-service")
			setup := NewClusterSetup().Install(YamlK8s(route...))
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

			Expect(NewClusterSetup().Install(DeleteYamlK8s(route...)).Setup(kubernetes.Cluster)).To(Succeed())
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
      kuma.io/service: simple-gateway_simple-gateway_svc
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
