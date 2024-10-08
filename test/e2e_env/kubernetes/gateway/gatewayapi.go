package gateway

import (
	"fmt"
	"io"
	"net"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	client "github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func GatewayAPI() {
	if Config.IPV6 {
		fmt.Println("IPv6 tests use kind which doesn't support the LoadBalancer ServiceType")
		return
	}

	meshName := "gatewayapi"
	namespace := "gatewayapi"
	externalServicesNamespace := "gatewayapi-external-services"

	externalService := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: ExternalService
mesh: %s
metadata:
  name: external-service
spec:
  tags:
    kuma.io/service: external-service
    kuma.io/protocol: http
  networking:
    address: external-service.gatewayapi-external-services.svc.cluster.local:80
`, meshName)

	BeforeAll(func() {
		setup := NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Namespace(externalServicesNamespace)).
			Install(MTLSMeshKubernetes(meshName)).
			Install(MeshTrafficPermissionAllowAllKubernetes(meshName)).
			Install(testserver.Install(
				testserver.WithName("test-server-1"),
				testserver.WithMesh(meshName),
				testserver.WithNamespace(namespace),
				testserver.WithEchoArgs("echo", "--instance", "test-server-1"),
			)).
			Install(testserver.Install(
				testserver.WithName("test-server-2"),
				testserver.WithMesh(meshName),
				testserver.WithNamespace(namespace),
				testserver.WithEchoArgs("echo", "--instance", "test-server-2"),
			)).
			Install(testserver.Install(
				testserver.WithName("external-service"),
				testserver.WithNamespace(externalServicesNamespace),
				testserver.WithEchoArgs("echo", "--instance", "external-service"),
			)).
			Install(YamlK8s(externalService))
		Expect(setup.Setup(kubernetes.Cluster)).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugKube(kubernetes.Cluster, meshName, namespace, externalServicesNamespace)
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(externalServicesNamespace)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	GatewayIP := func(name string) string {
		var ip string
		Eventually(func(g Gomega) {
			out, err := k8s.RunKubectlAndGetOutputE(
				kubernetes.Cluster.GetTesting(),
				kubernetes.Cluster.GetKubectlOptions(namespace),
				"get", "gateway", name, "-ojsonpath={.status.addresses[0].value}",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(out).ToNot(BeEmpty())
			ip = out
		}, "120s", "1s").Should(Succeed(), "could not get a LoadBalancer IP of the Gateway")
		return ip
	}

	Describe("GatewayClass parametersRef", Ordered, func() {
		gatewayName := "kuma-ha"
		haGatewayClass := `
apiVersion: gateway.networking.k8s.io/v1beta1
kind: GatewayClass
metadata:
  name: ha-kuma
spec:
  controllerName: gateways.kuma.io/controller
  parametersRef:
    kind: MeshGatewayConfig
    group: kuma.io
    name: ha-config`

		haConfig := `
apiVersion: kuma.io/v1alpha1
kind: MeshGatewayConfig
metadata:
  name: ha-config
spec:
  replicas: 3`

		haGateway := fmt.Sprintf(`
apiVersion: gateway.networking.k8s.io/v1beta1
kind: Gateway
metadata:
  name: %s
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  gatewayClassName: ha-kuma
  listeners:
  - name: proxy
    port: 10080
    protocol: HTTP
`, gatewayName, namespace, meshName)

		BeforeAll(func() {
			Expect(YamlK8s(haConfig)(kubernetes.Cluster)).To(Succeed())
			Expect(YamlK8s(haGatewayClass)(kubernetes.Cluster)).To(Succeed())
			Expect(YamlK8s(haGateway)(kubernetes.Cluster)).To(Succeed())
			Expect(WaitPodsAvailable(namespace, gatewayName)(kubernetes.Cluster)).To(Succeed())
		})
		E2EAfterAll(func() {
			Expect(k8s.RunKubectlE(kubernetes.Cluster.GetTesting(), kubernetes.Cluster.GetKubectlOptions(namespace), "delete", "gateway", gatewayName)).To(Succeed())
			Expect(k8s.RunKubectlE(kubernetes.Cluster.GetTesting(), kubernetes.Cluster.GetKubectlOptions(namespace), "delete", "gatewayclass", "ha-kuma")).To(Succeed())
			Expect(k8s.RunKubectlE(kubernetes.Cluster.GetTesting(), kubernetes.Cluster.GetKubectlOptions(namespace), "delete", "meshgatewayconfig", "ha-config")).To(Succeed())
		})

		It("should create the right number of pods", func() {
			Expect(kubernetes.Cluster.WaitApp(gatewayName, namespace, 3)).To(Succeed())
		})

		It("should create the right number of pods after updating MeshGatewayConfig", func() {
			newHaConfig := `
apiVersion: kuma.io/v1alpha1
kind: MeshGatewayConfig
metadata:
  name: ha-config
spec:
  replicas: 4`
			Expect(YamlK8s(newHaConfig)(kubernetes.Cluster)).To(Succeed())

			Expect(kubernetes.Cluster.WaitApp(gatewayName, namespace, 4)).To(Succeed())
		})
	})

	Context("HTTP Gateway", Ordered, func() {
		const gatewayName = "kuma-http"

		gateway := fmt.Sprintf(`
apiVersion: gateway.networking.k8s.io/v1beta1
kind: Gateway
metadata:
  name: %s
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  gatewayClassName: kuma
  listeners:
  - name: proxy
    port: 10080
    protocol: HTTP
`, gatewayName, namespace, meshName)

		var address string

		BeforeAll(func() {
			Expect(YamlK8s(gateway)(kubernetes.Cluster)).To(Succeed())
			address = net.JoinHostPort(GatewayIP(gatewayName), "10080")
			Expect(WaitPodsAvailable(namespace, gatewayName)(kubernetes.Cluster)).To(Succeed())
		})
		AfterEachFailure(func() {
			DebugKube(kubernetes.Cluster, meshName, namespace)
		})
		E2EAfterAll(func() {
			Expect(k8s.RunKubectlE(kubernetes.Cluster.GetTesting(), kubernetes.Cluster.GetKubectlOptions(namespace), "delete", "gateway", gatewayName)).To(Succeed())
		})

		It("should send default static payload for no route", func() {
			Eventually(func(g Gomega) {
				resp, err := client.MakeDirectRequest("http://" + address)

				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(resp.StatusCode).To(Equal(404))

				defer resp.Body.Close()
				body, err := io.ReadAll(resp.Body)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(body).ToNot(BeEmpty())
			}, "30s", "1s").Should(Succeed())
		})

		It("should route the traffic to test-server by path", func() {
			// given
			route := fmt.Sprintf(`
apiVersion: gateway.networking.k8s.io/v1beta1
kind: HTTPRoute
metadata:
  name: test-server-paths
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  parentRefs:
  - name: %s
  rules:
  - backendRefs:
    - name: test-server-1
      port: 80
    matches:
    - path:
        type: PathPrefix
        value: /1
  - backendRefs:
    - name: test-server-2
      port: 80
    matches:
    - path:
        type: PathPrefix
        value: /2
`, namespace, meshName, gatewayName)

			// when
			err := YamlK8s(route)(kubernetes.Cluster)

			// then
			Expect(err).ToNot(HaveOccurred())

			Eventually(func(g Gomega) {
				resp, err := client.CollectResponseDirectly("http://" + address + "/1")
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(resp.Instance).To(Equal("test-server-1"))
			}, "30s", "1s").Should(Succeed())

			Eventually(func(g Gomega) {
				resp, err := client.CollectResponseDirectly("http://" + address + "/2")
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(resp.Instance).To(Equal("test-server-2"))
			}, "30s", "1s").Should(Succeed())

			Expect(k8s.KubectlDeleteFromStringE(
				kubernetes.Cluster.GetTesting(),
				kubernetes.Cluster.GetKubectlOptions(namespace),
				route,
			)).To(Succeed())
		})

		It("should route the traffic to test-server by header", func() {
			// given
			routes := fmt.Sprintf(`
apiVersion: gateway.networking.k8s.io/v1beta1
kind: HTTPRoute
metadata:
  name: test-server-1
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  parentRefs:
  - name: %s
  hostnames:
  - "test-server-1.com"
  rules:
  - backendRefs:
    - name: test-server-1
      port: 80
---
apiVersion: gateway.networking.k8s.io/v1beta1
kind: HTTPRoute
metadata:
  name: test-server-2
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  parentRefs:
  - name: %s
  hostnames:
  - "test-server-2.com"
  rules:
  - backendRefs:
    - name: test-server-2
      port: 80
`, namespace, meshName, gatewayName, namespace, meshName, gatewayName)

			// when
			err := YamlK8s(routes)(kubernetes.Cluster)

			// then
			Expect(err).ToNot(HaveOccurred())

			Eventually(func(g Gomega) {
				resp, err := client.CollectResponseDirectly("http://"+address, client.WithHeader("host", "test-server-1.com"))
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(resp.Instance).To(Equal("test-server-1"))
			}, "30s", "1s").Should(Succeed())

			Eventually(func(g Gomega) {
				resp, err := client.CollectResponseDirectly("http://"+address, client.WithHeader("host", "test-server-2.com"))
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(resp.Instance).To(Equal("test-server-2"))
			}, "30s", "1s").Should(Succeed())

			Expect(k8s.KubectlDeleteFromStringE(
				kubernetes.Cluster.GetTesting(),
				kubernetes.Cluster.GetKubectlOptions(namespace),
				routes,
			)).To(Succeed())
		})

		It("should route to external service", func() {
			// given
			routes := fmt.Sprintf(`
apiVersion: gateway.networking.k8s.io/v1beta1
kind: HTTPRoute
metadata:
  name: external-service
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  parentRefs:
  - name: %s
  hostnames:
  - "external-service.com"
  rules:
  - backendRefs:
    - group: kuma.io
      kind: ExternalService
      name: external-service
`, namespace, meshName, gatewayName)

			// when
			err := YamlK8s(routes)(kubernetes.Cluster)

			// then
			Expect(err).ToNot(HaveOccurred())

			Eventually(func(g Gomega) {
				resp, err := client.CollectResponseDirectly("http://"+address, client.WithHeader("host", "external-service.com"))
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(resp.Instance).To(Equal("external-service"))
			}, "30s", "1s").Should(Succeed())

			Expect(k8s.KubectlDeleteFromStringE(
				kubernetes.Cluster.GetTesting(),
				kubernetes.Cluster.GetKubectlOptions(namespace),
				routes,
			)).To(Succeed())
		})
	})

	Context("HTTPS Gateway", Ordered, func() {
		const gatewayName = "kuma-https"
		secret := fmt.Sprintf(`
apiVersion: v1
kind: Secret
metadata:
  name: secret-tls
  namespace: %s
type: kubernetes.io/tls
data:
  tls.crt: >-
    LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUVPekNDQXlPZ0F3SUJBZ0lSQU5RUisvcTNEWk5jLy80ckdXKzR5am93RFFZSktvWklodmNOQVFFTEJRQXcKSFRFYk1Ca0dBMVVFQXhNU2EzVnRZUzFqYjI1MGNtOXNMWEJzWVc1bE1CNFhEVEl5TURFeU5ERXdNekExTmxvWApEVE15TURFeU1qRXdNekExTmxvd0hURWJNQmtHQTFVRUF4TVNhM1Z0WVMxamIyNTBjbTlzTFhCc1lXNWxNSUlCCklqQU5CZ2txaGtpRzl3MEJBUUVGQUFPQ0FROEFNSUlCQ2dLQ0FRRUF0M0E2bzRvT3A3MnZMV1ZmTVM1WWFHRzAKdFFHY0FRVEFvR3diY2VheUFaN2xQbElNZVFPNXdNYlpFUGw3TUMvZWJxQ3NNWjB0THJoQ0tOUHh2QUt5WG05OAoyK2lRVlk5bkNZemlQZGZmZEY1aVk2SDhQL0FxNDVSbDVIYmpmcFNIZ3JRYlRXUUZCbUNsNkJrV3BPTEYwcThOCkozV0RpVHMvdnlkSWF6Q0tOTjRsTlIxVEFSODdXL0c3MHRxVnd2R1FEN1Y0VXFFUDRia05nQVVmNW5iSmtZTSsKeVROSG9remRaUFhyaHlmMkhqYXlzekRQZWhEMThlMkNaeDJhWEMxMVFRSnFHQmp6QWsvU2FOVUpya1poc0lpYgo5SGZZQ3BQMHprTmNYWms0MnJPMVdMOVRaNDZQUjNNVVNYa1A4Q1lkcUxrV1pGc0kwUFZGNk5ZdHA1cEJUUUlECkFRQUJvNElCZERDQ0FYQXdEZ1lEVlIwUEFRSC9CQVFEQWdLa01CTUdBMVVkSlFRTU1Bb0dDQ3NHQVFVRkJ3TUIKTUE4R0ExVWRFd0VCL3dRRk1BTUJBZjh3SFFZRFZSME9CQllFRkdJRDAvUVRpclJUTEtwVmpZc09SVjhjaEZZMQpNSUlCRndZRFZSMFJCSUlCRGpDQ0FRcUNCMkZ1ZEMxa1pYYUNDV3h2WTJGc2FHOXpkSWNFWkVQbEFZY0Vmd0FBCkFZY0VyQkVBQVljRXJCSUFBWWNFd0tnQk00Y1FBQUFBQUFBQUFBQUFBQUFBQUFBQUFZY1EvUUQ5RWpSV0FBQUEKQUFBQUFBQUFBWWNRL1hvUlhLSGdxeEpJUTgyV1lrUGxBWWNRL29BQUFBQUFBQUFBQUFBQUFBQUFBWWNRL29BQQpBQUFBQUFBZ1B6Yi8vdXhJWFljUS9vQUFBQUFBQUFBQVFnai8vdVVIMlljUS9vQUFBQUFBQUFBQVFwMy8vamdCCnA0Y1Evb0FBQUFBQUFBQlFWQUQvL3U4NUVvY1Evb0FBQUFBQUFBQlVtSUQvL254dHZZY1Evb0FBQUFBQUFBQmcKRlhYLy9yZ1V0SWNRL29BQUFBQUFBQURJLzMvLy9tZkJ1WWNRL29BQUFBQUFBQUQyeTI0dWlNY0huVEFOQmdrcQpoa2lHOXcwQkFRc0ZBQU9DQVFFQUxkTVBnaE1sRGdSQW04UHJwL0FxdERGWTRLN3p4Qmhzc2dTNWNnUWtKdnU3CitJVmszQ2o2aXdObUFhdDZCdFJYUmREODUxdlJxRDBzNk90QXBUZXlyaVcrZlgwcWN1UVc1NXVQbTZFM0JEZGcKNU9qZXRhYU9heXppUmRzeTdOU0N2bWtrWURRUVQvTTF6WDBXdlBXTkR0SDhpd2c1aEpoOHFrK3A0Q2M3blAvSAowSlBpaVQ1TEs1bE1aOGZTRHowUHBGeTF0MUd3N3RzTkhBdHN6NkZGaDJOZ1FtdkxpNFJpa1J4SGViRWlZdzlECjhtSzQ3WSsxVnErWFQ3eHd1aTZ0YzBCQXRRSnZVSUQremMvazg5QU55YmpFSFNvMG01d2RyTmpiRzBBb0xxbVAKbHM5UHY0cDNIbjJRMlRaVW5xd250Nk12cm1zYlVpSFhGYXFsOE9FclpBPT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
  tls.key: >-
    LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFb3dJQkFBS0NBUUVBdDNBNm80b09wNzJ2TFdWZk1TNVlhR0cwdFFHY0FRVEFvR3diY2VheUFaN2xQbElNCmVRTzV3TWJaRVBsN01DL2VicUNzTVowdExyaENLTlB4dkFLeVhtOTgyK2lRVlk5bkNZemlQZGZmZEY1aVk2SDgKUC9BcTQ1Umw1SGJqZnBTSGdyUWJUV1FGQm1DbDZCa1dwT0xGMHE4TkozV0RpVHMvdnlkSWF6Q0tOTjRsTlIxVApBUjg3Vy9HNzB0cVZ3dkdRRDdWNFVxRVA0YmtOZ0FVZjVuYkprWU0reVROSG9remRaUFhyaHlmMkhqYXlzekRQCmVoRDE4ZTJDWngyYVhDMTFRUUpxR0JqekFrL1NhTlVKcmtaaHNJaWI5SGZZQ3BQMHprTmNYWms0MnJPMVdMOVQKWjQ2UFIzTVVTWGtQOENZZHFMa1daRnNJMFBWRjZOWXRwNXBCVFFJREFRQUJBb0lCQUhyRHJUck5wa2swZFF4WQpqNENHbDd3ang2QnIxMUFITWpNcXBxTnYxU21vZ1p0WHBlbEhTUVZ2RHM2QmFLUXpKUlc4aWdFYVE2YkV3ZUk1CkZjclJzelhvUHhPZGJSc1Z3Y3R1Y2VzWmtmNTdQRFdacnd2TFc2aTdKQVhtV3hIWHJXa1h5RDNlOWszeVdKWWcKVkRzOVdVOUt2KzdzZ245Ukc3UitRY1VhMHlQVmM5MFN1RU0vengxVG9ZR09QL3ZJUjNwVlkwVW9jNHFWMWt6ZwpKUUJvVWxBcGZJU0w2TWpEWE5Eekk1QjRkNWpyMWVWTmJZNEdxbnlSNDlOMXRnRjRjdFFYSzNvYUJTY1BSaHl2Ci8rRXJOVDZiNDFUeUsyVFRYQkNMdDM0Ynl0REtxc3kxVXIvdE1lOWE5NnVxU1ZIUkFjbzU3YjZ3MGhZV0pvUlAKVjRFVDlnRUNnWUVBd01MSDR4ZUN1QjBraE9ERTJsbU9ZekJDQ1Rpd3p0bkltVlRXN0hnS1B2V0FTR1pVRS91dgpHQWZackVId0o0NERvT2pnTkgyN09qL2ZxZTNIeDZ0VXNBTlJ3THF2dkhCcldLdWZRS090THFHTzN6MUlCWGtWCmFtV1QyazJJMTVmakVVbnVCcGZkc1ROUEdFNHFWeU53V2szdUpNNytsQlFQYkRqYWsvaFNLQzBDZ1lFQTg1NTkKditsSXV0UHpIUmYwaGcxSUtzZWJ5cWRJYm41RytQOVNEUXV1aHc3UFhKQnhXS2JsUmd6MTFwdW1BUi96U3RnQQpDcUpncUJhbmQzV1JRWVliZnBiN0VpK0FoSkRpMVp3ZGRGVGsybUE5YUtsVlVOdXl5VHg5cjNCTjlzV1lxRGVnCm4rdGJmL3lyNnBLaHBRd25VMFVtMW5OYVdhbG0xczBuM2dyVUVhRUNnWUFocW1NcXdGSnVRWGk5VkZ4TkhsTUYKODhtMHZwZnlxSXFtYlBEVWYrcWFNRnBsU3Fub2k0NTdEZlB3Wjl1L3JNZnBkSUtqNkVtbzFMc0ZmS2Zsc1lDcQo5UWwwTmFhM3JKS3krOVptZmEramMwZjJxVWRJM1dybUdET0lidjQxV1N1cE8xWTlCSTBOZzc2T3FpZ3U2OXVWCmlnTExudk5MZlcxc0kwblppZ2NmU1FLQmdCQmY1Y25oYno4SGdmN0JubkRvTWFLV2VoVTcrelZhRFlFdEFDSGEKV0NmQnloUkpyU1N0U3huVFF5N2lsVnpiL2VsWTdWL0puRCtRRGorTVNuQWlDSFVReHQxcERmVmJHN1FKNHp6dgplOVpsdzVybVR0SzVnYUhmQy8rZng4Mi9hRXhlT05DbTdDYUZJRFVMR0F4VTdjdStDU2MrNTZMQkxTVmc4cjRNCjhrWWhBb0dCQUtMaDVnblYxaEhYcWNZckZpWjVpWVY5L3BJUXFMdkI4VXNHVUljRm1Vb3hnZ0JKcXpRUUExaDAKQXE5cW4rTStBUlNNRUk0WFVmUjhqQnVTNk1sWmh6czRyeXRUckt0dWdmSFpLNEdGQjFuYzNPSlR5QmpTWUFmZAp5a2t4UUtrcWd4aERDZUFveG1qVkRUUDhLRGgzZ2hySUFBWWozSDFzS0Q0K1dmRXZyRHByCi0tLS0tRU5EIFJTQSBQUklWQVRFIEtFWS0tLS0tCg==
`, namespace)

		gateway := fmt.Sprintf(`
apiVersion: gateway.networking.k8s.io/v1beta1
kind: Gateway
metadata:
  name: %s
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  gatewayClassName: kuma
  listeners:
  - name: proxy
    port: 8090
    hostname: 'test-server-1.com'
    protocol: HTTPS
    tls:
      certificateRefs:
      - name: secret-tls
  - name: proxy-wildcard
    port: 8091
    protocol: HTTPS
    tls:
      certificateRefs:
      - name: secret-tls
`, gatewayName, namespace, meshName)

		var ip string

		BeforeAll(func() {
			Expect(YamlK8s(secret)(kubernetes.Cluster)).To(Succeed())
			Expect(YamlK8s(gateway)(kubernetes.Cluster)).To(Succeed())
			ip = GatewayIP(gatewayName)
			Expect(WaitPodsAvailable(namespace, gatewayName)(kubernetes.Cluster)).To(Succeed())
		})
		E2EAfterAll(func() {
			Expect(k8s.RunKubectlE(kubernetes.Cluster.GetTesting(), kubernetes.Cluster.GetKubectlOptions(namespace), "delete", "gateway", gatewayName)).To(Succeed())
		})

		It("should route the traffic using TLS", func() {
			// given
			route := fmt.Sprintf(`
apiVersion: gateway.networking.k8s.io/v1beta1
kind: HTTPRoute
metadata:
  name: test-server-paths
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  parentRefs:
  - name: %s
  rules:
  - backendRefs:
    - name: test-server-1
      port: 80
    matches:
    - path:
        type: PathPrefix
        value: /
`, namespace, meshName, gatewayName)

			// when
			err := YamlK8s(route)(kubernetes.Cluster)

			// then
			Expect(err).ToNot(HaveOccurred())

			Eventually(func(g Gomega) {
				resp, err := client.CollectResponseDirectly("https://"+net.JoinHostPort(ip, "8090"), client.WithHeader("host", "test-server-1.com"))
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(resp.Instance).To(Equal("test-server-1"))
			}, "30s", "1s").Should(Succeed())

			Eventually(func(g Gomega) {
				resp, err := client.CollectResponseDirectly("https://" + net.JoinHostPort(ip, "8091"))
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(resp.Instance).To(Equal("test-server-1"))
			}, "30s", "1s").Should(Succeed())

			Expect(k8s.KubectlDeleteFromStringE(
				kubernetes.Cluster.GetTesting(),
				kubernetes.Cluster.GetKubectlOptions(namespace),
				route,
			)).To(Succeed())
		})

		It("should manage Kuma Secret", func() {
			// given converted Kuma secret
			convertedSecretName := fmt.Sprintf("gapi-%s-secret-tls", namespace)
			var kumaSecret string
			Eventually(func(g Gomega) {
				out, err := kubernetes.Cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "secret", "-m", meshName, convertedSecretName, "-o", "json")
				g.Expect(err).ToNot(HaveOccurred())
				kumaSecret = out
			}, "30s", "1s").Should(Succeed())

			// when original secret is changed
			secret = fmt.Sprintf(`
apiVersion: v1
kind: Secret
metadata:
  name: secret-tls
  namespace: %s
type: kubernetes.io/tls
data:
  tls.crt: >-
    LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURLRENDQWhDZ0F3SUJBZ0lRZUxvNDFxKzJMcWpNWjNubnZSMWd2akFOQmdrcWhraUc5dzBCQVFzRkFEQVhNUlV3RXdZRFZRUURFd3gwWlhOMExtdDFiV0V1YVc4d0hoY05Nakl3TXpJME1UTTFNRE0xV2hjTk16SXdNekl4TVRNMU1ETTFXakFYTVJVd0V3WURWUVFERXd4MFpYTjBMbXQxYldFdWFXOHdnZ0VpTUEwR0NTcUdTSWIzRFFFQkFRVUFBNElCRHdBd2dnRUtBb0lCQVFDbW41aVNPV0h6Y1YrWEtxQ1Z4QTRoaUhBd2UrbnZNWkxxbmcwWkZxY1FiNkVZTWw3bXMzQU5VaXlRUVNURVZyWkdYTXJDZWY0N0VYdWZ1NlVFeGsyNkcrK2FTcjFpKzZlZE5ELzAyZklQL1JwYmZDMXlqdnhjRGJ6eWE4SUlING5DUE5LNElPVmpBZmtVRVpmSk1aVGRORUZ5MUN0S3Nod2hRY3BPd3d6em5uNmhRM284U3d3SFhlSGxlNGFCdnMvTnlpT1FTbFNXTzVxcUszdHRHSWxaRmMvbGJJWVU0N2Rqd2tMVFBNVFpXTW9BdjZvSEhVdzdvSkRDR2lGaGpzQjA3UFRMaS9udTUrekhKaG1jWUJTSC9CUkprMHpXMTlLS09odVhlQ2lLRCtJV0U5ZEN1bjZWSmpUb3B6WllNZzdBYnYwMmdkRStjYkVCejRZcTZyR0hBZ01CQUFHamNEQnVNQTRHQTFVZER3RUIvd1FFQXdJQ3BEQVRCZ05WSFNVRUREQUtCZ2dyQmdFRkJRY0RBVEFQQmdOVkhSTUJBZjhFQlRBREFRSC9NQjBHQTFVZERnUVdCQlIzb1loRDMxK0c1TUFEdVlUR2tXbzlXczJJNERBWEJnTlZIUkVFRURBT2dneDBaWE4wTG10MWJXRXVhVzh3RFFZSktvWklodmNOQVFFTEJRQURnZ0VCQUNhTUZZTWR5REt0WlhLN1hJRVJ2cmpxVlQ4VXlqTVp5RHFwVzNUeW9rSVliMGNmOHNuaEVHQUZlelhiNmZUTnd4TThEaUtFWEZnK0RMSnNJWEwzcTYyK3BWNjdOTjI1R0c1eklGSit4bG9YU1NWbXRMWEhvUVpJSGNLaUJnenVWaFo3d3NMdmo1UjYyTll4d2RpS2piZ0VSblRlaldpQzRlRUJsM0hHSk51NTV6RjI1cFRZdXZocGhwSnZmYUhzdWh2bndIWE1WbXhDWGErZFF3czV3T3dqalNIYkFMOER3Rk9pd0JBTXMrNEI2QXZQNzBYN0NkdURXV0tmUzdmVE1VN3lvNGtOT1JGRDc5TXBzWEN2QVlQenJwYjY4cDdJdjlXQTlSU0dLSXhrdjFuaFNnbjhtVWZvaHBQVVU1L2RuMzdxQnFZTUtTbG9uVlZWYUNUOWU0Yz0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
  tls.key: >-
    LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcEFJQkFBS0NBUUVBcHArWWtqbGg4M0ZmbHlxZ2xjUU9JWWh3TUh2cDd6R1M2cDROR1JhbkVHK2hHREplCjVyTndEVklza0VFa3hGYTJSbHpLd25uK094RjduN3VsQk1aTnVodnZta3E5WXZ1bm5UUS85Tm55RC8wYVczd3QKY283OFhBMjg4bXZDQ0IrSndqelN1Q0RsWXdINUZCR1h5VEdVM1RSQmN0UXJTckljSVVIS1RzTU04NTUrb1VONgpQRXNNQjEzaDVYdUdnYjdQemNvamtFcFVsanVhcWl0N2JSaUpXUlhQNVd5R0ZPTzNZOEpDMHp6RTJWaktBTCtxCkJ4MU1PNkNRd2hvaFlZN0FkT3oweTR2NTd1ZnN4eVlabkdBVWgvd1VTWk5NMXRmU2lqb2JsM2dvaWcvaUZoUFgKUXJwK2xTWTA2S2MyV0RJT3dHNzlOb0hSUG5HeEFjK0dLdXF4aHdJREFRQUJBb0lCQUdNbHNtN0lNRzNndDRYRworcmxEYVRrdzY3a2Q4dXkrN2ZJbnpCbHlya1NNZUNwaXhxKzJkR1dvMFJXaGZkUksyTGx6dTc4UFFtVTVtUHRLCmQvNG9WZFg1aTVDZkNxU2NwSGRad1BqY3V6b2lYSTIxallHT2JjSUU5cnExdmthQkpjTHIyR055UjZ5clh1QS8KTzdlZmhqbytQdmVxSW55WEVVQUUydklWQkY3dHFoYVpIalA3S1dZa0U1cjRlSGMrM0NSbVB5VWFSTDhsdXpRNwowc1hLZEJUMmZ5WnRTc0NadXM3aEt3dHN3K3V6dFY5QjN0VjA0V3BONDhKMkMzQjFoanNjU2xWeS9UNHo5Z1ZBCkJWU21aMENnVUY2VmU2QW1zNmJDaVcxR3ZWOUVuTnBSUkpXb1BQNDRleldyU1ZYUkxFcThkVmtySWlISUkrQW0KeUVwT2JkRUNnWUVBemFuRDRFOTZZb3E1bys0UDlYbERTdTZld3JEZDloc3pmYWZTcXovcGx6SW1zQi8veEFLNQo4VE40T2trR0dPRC82cDBLR3EzdU9XN2Z5c1ArZDZBbWZLNXhZcENvUnA4ZXNYNUNSSmxKS1pBY09YWHFGaHFwCklYUEdQcDRFQ0tYRGJCTitnQTZSZm9zVzdaYTI3b0hHdDNqYitwSXFaeDRCMENDSC84VDR3ejhDZ1lFQXoyZTQKRS9EV09FUGVJeElyb3pTeFo4VG00dU9hTFlucEY5NkxHc2Q3cGNveTJiVW5vMTUxMy84bGplaTlGRytxKzNRUwp2TVgzeGlaMmc1OURqby9FRmdDeHVFRURHZGw5a1FLemJ4bEVPamRNUnAvNDVNV1VySm5lUklhWklBRGExVmt0CkludmZndnl3NXlra2t3SWFlalhnc2dFNGp4WUVaNm1nZkZQWko3a0NnWUVBakMrcjFMcFlNZE5kdHVBUEFNUW4KbW13TXk2akRvMzNuR3ovSjJmRTJ5RmpuQmliSnNGSXJiTDRvdFpJUkZlUklqU04rUDdGUE1OYml0TlBrSUthSgpsWE5TMWx6RVYxOGZETjJEVGo4dUg2YWJsbzlKZ01lcmdhSG8vOFcxK2k4RGhpZkRrb1picG1Zb3VzcUE1eEtPCjRZRUFjVXd3bXhsWkl3VUpyczRVd3dFQ2dZQXBaY0ZuTVlZQW13Tkdxc1ROQWFKN1hPRGMzcU1TZmRscHEwREcKcXBSeWhnWmFULzlHYTM5Sm8yckNoWGJnRWwzbGJNaWtwenNLY1BqczBxZ3dWMS9ES0laUWlhRnQwbXh1dWtSSQpZNW1ycVFmdmZOUzREUHZjNjZWaXRoN3dOVnQ0aENFdkpkeDZENmZicSttaDhpU0l5aUk4UldRZG96NWoxb2F5CjZpV0krUUtCZ1FDdmVUdDNDWFpuT2FOVXpHNzYxc293WFh4ZGpCMFlvUExjTXJGMFZyaGtNZlZrYnR5NTM1bDgKV2lzOVdkNVhVSW1LV21jWExzUlZGczRqVi9ZWCtLcHN6TmdSdWVmVEoycDhWL1JCMEZ3aDduV3ZlQzNuOHZOZwpONkFWd0VBczdERVhxZFZYek5xNmwzZHlOeFgxTDNNVlBlVmJQVEVmWlVOdktqNmVkMjJ0ZkE9PQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=
`, namespace)
			err := YamlK8s(secret)(kubernetes.Cluster)

			// then copied secret is also changed
			Expect(err).ToNot(HaveOccurred())
			Eventually(func(g Gomega) {
				out, err := kubernetes.Cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "secret", "-m", meshName, convertedSecretName, "-o", "json")
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(out).ToNot(MatchJSON(kumaSecret))
			}, "30s", "1s").Should(Succeed())

			// when original secret is removed
			err = k8s.RunKubectlE(kubernetes.Cluster.GetTesting(), kubernetes.Cluster.GetKubectlOptions(namespace), "delete", "secret", "secret-tls")

			// then copied secret is removed
			Expect(err).ToNot(HaveOccurred())
			Eventually(func(g Gomega) {
				_, err := kubernetes.Cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "secret", "-m", meshName, convertedSecretName)
				g.Expect(err).To(HaveOccurred())
			}, "30s", "1s").Should(Succeed())
		})
	})
}
