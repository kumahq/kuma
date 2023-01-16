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
		}, "60s", "1s").Should(Succeed(), "could not get a LoadBalancer IP of the Gateway")
		return ip
	}

	Describe("GatewayClass parametersRef", Ordered, func() {
		gatewayName := "kuma-ha"
		haGatewayClass := `
apiVersion: gateway.networking.k8s.io/v1alpha2
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
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: Gateway
metadata:
  name: %s
  namespace: %s
  annotations:
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
  annotations:
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
  annotations:
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
		})

		It("should route the traffic to test-server by header", func() {
			// given
			routes := fmt.Sprintf(`
apiVersion: gateway.networking.k8s.io/v1beta1
kind: HTTPRoute
metadata:
  name: test-server-1
  namespace: %s
  annotations:
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
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: HTTPRoute
metadata:
  name: test-server-2
  namespace: %s
  annotations:
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
		})

		It("should route to external service", func() {
			// given
			routes := fmt.Sprintf(`
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: HTTPRoute
metadata:
  name: external-service
  namespace: %s
  annotations:
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
  tls.crt: "MIIEOzCCAyOgAwIBAgIRANQR+/q3DZNc//4rGW+4yjowDQYJKoZIhvcNAQELBQAwHTEbMBkGA1UEAxMSa3VtYS1jb250cm9sLXBsYW5lMB4XDTIyMDEyNDEwMzA1NloXDTMyMDEyMjEwMzA1NlowHTEbMBkGA1UEAxMSa3VtYS1jb250cm9sLXBsYW5lMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAt3A6o4oOp72vLWVfMS5YaGG0tQGcAQTAoGwbceayAZ7lPlIMeQO5wMbZEPl7MC/ebqCsMZ0tLrhCKNPxvAKyXm982+iQVY9nCYziPdffdF5iY6H8P/Aq45Rl5HbjfpSHgrQbTWQFBmCl6BkWpOLF0q8NJ3WDiTs/vydIazCKNN4lNR1TAR87W/G70tqVwvGQD7V4UqEP4bkNgAUf5nbJkYM+yTNHokzdZPXrhyf2HjayszDPehD18e2CZx2aXC11QQJqGBjzAk/SaNUJrkZhsIib9HfYCpP0zkNcXZk42rO1WL9TZ46PR3MUSXkP8CYdqLkWZFsI0PVF6NYtp5pBTQIDAQABo4IBdDCCAXAwDgYDVR0PAQH/BAQDAgKkMBMGA1UdJQQMMAoGCCsGAQUFBwMBMA8GA1UdEwEB/wQFMAMBAf8wHQYDVR0OBBYEFGID0/QTirRTLKpVjYsORV8chFY1MIIBFwYDVR0RBIIBDjCCAQqCB2FudC1kZXaCCWxvY2FsaG9zdIcEZEPlAYcEfwAAAYcErBEAAYcErBIAAYcEwKgBM4cQAAAAAAAAAAAAAAAAAAAAAYcQ/QD9EjRWAAAAAAAAAAAAAYcQ/XoRXKHgqxJIQ82WYkPlAYcQ/oAAAAAAAAAAAAAAAAAAAYcQ/oAAAAAAAAAgPzb//uxIXYcQ/oAAAAAAAAAAQgj//uUH2YcQ/oAAAAAAAAAAQp3//jgBp4cQ/oAAAAAAAABQVAD//u85EocQ/oAAAAAAAABUmID//nxtvYcQ/oAAAAAAAABgFXX//rgUtIcQ/oAAAAAAAADI/3///mfBuYcQ/oAAAAAAAAD2y24uiMcHnTANBgkqhkiG9w0BAQsFAAOCAQEALdMPghMlDgRAm8Prp/AqtDFY4K7zxBhssgS5cgQkJvu7+IVk3Cj6iwNmAat6BtRXRdD851vRqD0s6OtApTeyriW+fX0qcuQW55uPm6E3BDdg5OjetaaOayziRdsy7NSCvmkkYDQQT/M1zX0WvPWNDtH8iwg5hJh8qk+p4Cc7nP/H0JPiiT5LK5lMZ8fSDz0PpFy1t1Gw7tsNHAtsz6FFh2NgQmvLi4RikRxHebEiYw9D8mK47Y+1Vq+XT7xwui6tc0BAtQJvUID+zc/k89ANybjEHSo0m5wdrNjbG0AoLqmPls9Pv4p3Hn2Q2TZUnqwnt6MvrmsbUiHXFaql8OErZA=="
  tls.key: "MIIEowIBAAKCAQEAt3A6o4oOp72vLWVfMS5YaGG0tQGcAQTAoGwbceayAZ7lPlIMeQO5wMbZEPl7MC/ebqCsMZ0tLrhCKNPxvAKyXm982+iQVY9nCYziPdffdF5iY6H8P/Aq45Rl5HbjfpSHgrQbTWQFBmCl6BkWpOLF0q8NJ3WDiTs/vydIazCKNN4lNR1TAR87W/G70tqVwvGQD7V4UqEP4bkNgAUf5nbJkYM+yTNHokzdZPXrhyf2HjayszDPehD18e2CZx2aXC11QQJqGBjzAk/SaNUJrkZhsIib9HfYCpP0zkNcXZk42rO1WL9TZ46PR3MUSXkP8CYdqLkWZFsI0PVF6NYtp5pBTQIDAQABAoIBAHrDrTrNpkk0dQxYj4CGl7wjx6Br11AHMjMqpqNv1SmogZtXpelHSQVvDs6BaKQzJRW8igEaQ6bEweI5FcrRszXoPxOdbRsVwctucesZkf57PDWZrwvLW6i7JAXmWxHXrWkXyD3e9k3yWJYgVDs9WU9Kv+7sgn9RG7R+QcUa0yPVc90SuEM/zx1ToYGOP/vIR3pVY0Uoc4qV1kzgJQBoUlApfISL6MjDXNDzI5B4d5jr1eVNbY4GqnyR49N1tgF4ctQXK3oaBScPRhyv/+ErNT6b41TyK2TTXBCLt34bytDKqsy1Ur/tMe9a96uqSVHRAco57b6w0hYWJoRPV4ET9gECgYEAwMLH4xeCuB0khODE2lmOYzBCCTiwztnImVTW7HgKPvWASGZUE/uvGAfZrEHwJ44DoOjgNH27Oj/fqe3Hx6tUsANRwLqvvHBrWKufQKOtLqGO3z1IBXkVamWT2k2I15fjEUnuBpfdsTNPGE4qVyNwWk3uJM7+lBQPbDjak/hSKC0CgYEA8559v+lIutPzHRf0hg1IKsebyqdIbn5G+P9SDQuuhw7PXJBxWKblRgz11pumAR/zStgACqJgqBand3WRQYYbfpb7Ei+AhJDi1ZwddFTk2mA9aKlVUNuyyTx9r3BN9sWYqDegn+tbf/yr6pKhpQwnU0Um1nNaWalm1s0n3grUEaECgYAhqmMqwFJuQXi9VFxNHlMF88m0vpfyqIqmbPDUf+qaMFplSqnoi457DfPwZ9u/rMfpdIKj6Emo1LsFfKflsYCq9Ql0Naa3rJKy+9Zmfa+jc0f2qUdI3WrmGDOIbv41WSupO1Y9BI0Ng76Oqigu69uVigLLnvNLfW1sI0nZigcfSQKBgBBf5cnhbz8Hgf7BnnDoMaKWehU7+zVaDYEtACHaWCfByhRJrSStSxnTQy7ilVzb/elY7V/JnD+QDj+MSnAiCHUQxt1pDfVbG7QJ4zzve9Zlw5rmTtK5gaHfC/+fx82/aExeONCm7CaFIDULGAxU7cu+CSc+56LBLSVg8r4M8kYhAoGBAKLh5gnV1hHXqcYrFiZ5iYV9/pIQqLvB8UsGUIcFmUoxggBJqzQQA1h0Aq9qn+M+ARSMEI4XUfR8jBuS6MlZhzs4rytTrKtugfHZK4GFB1nc3OJTyBjSYAfdykkxQKkqgxhDCeAoxmjVDTP8KDh3ghrIAAYj3H1sKD4+WfEvrDpr"
`, namespace)

		gateway := fmt.Sprintf(`
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: Gateway
metadata:
  name: %s
  namespace: %s
  annotations:
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
		})
		E2EAfterAll(func() {
			Expect(k8s.RunKubectlE(kubernetes.Cluster.GetTesting(), kubernetes.Cluster.GetKubectlOptions(namespace), "delete", "gateway", gatewayName)).To(Succeed())
		})

		It("should route the traffic using TLS", func() {
			// given
			route := fmt.Sprintf(`
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: HTTPRoute
metadata:
  name: test-server-paths
  namespace: %s
  annotations:
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
  tls.crt: "MIIDKDCCAhCgAwIBAgIQeLo41q+2LqjMZ3nnvR1gvjANBgkqhkiG9w0BAQsFADAXMRUwEwYDVQQDEwx0ZXN0Lmt1bWEuaW8wHhcNMjIwMzI0MTM1MDM1WhcNMzIwMzIxMTM1MDM1WjAXMRUwEwYDVQQDEwx0ZXN0Lmt1bWEuaW8wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCmn5iSOWHzcV+XKqCVxA4hiHAwe+nvMZLqng0ZFqcQb6EYMl7ms3ANUiyQQSTEVrZGXMrCef47EXufu6UExk26G++aSr1i+6edND/02fIP/RpbfC1yjvxcDbzya8IIH4nCPNK4IOVjAfkUEZfJMZTdNEFy1CtKshwhQcpOwwzznn6hQ3o8SwwHXeHle4aBvs/NyiOQSlSWO5qqK3ttGIlZFc/lbIYU47djwkLTPMTZWMoAv6oHHUw7oJDCGiFhjsB07PTLi/nu5+zHJhmcYBSH/BRJk0zW19KKOhuXeCiKD+IWE9dCun6VJjTopzZYMg7Abv02gdE+cbEBz4Yq6rGHAgMBAAGjcDBuMA4GA1UdDwEB/wQEAwICpDATBgNVHSUEDDAKBggrBgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBR3oYhD31+G5MADuYTGkWo9Ws2I4DAXBgNVHREEEDAOggx0ZXN0Lmt1bWEuaW8wDQYJKoZIhvcNAQELBQADggEBACaMFYMdyDKtZXK7XIERvrjqVT8UyjMZyDqpW3TyokIYb0cf8snhEGAFezXb6fTNwxM8DiKEXFg+DLJsIXL3q62+pV67NN25GG5zIFJ+xloXSSVmtLXHoQZIHcKiBgzuVhZ7wsLvj5R62NYxwdiKjbgERnTejWiC4eEBl3HGJNu55zF25pTYuvhphpJvfaHsuhvnwHXMVmxCXa+dQws5wOwjjSHbAL8DwFOiwBAMs+4B6AvP70X7CduDWWKfS7fTMU7yo4kNORFD79MpsXCvAYPzrpb68p7Iv9WA9RSGKIxkv1nhSgn8mUfohpPUU5/dn37qBqYMKSlonVVVaCT9e4c="
  tls.key: "MIIEpAIBAAKCAQEApp+Ykjlh83FflyqglcQOIYhwMHvp7zGS6p4NGRanEG+hGDJe5rNwDVIskEEkxFa2RlzKwnn+OxF7n7ulBMZNuhvvmkq9YvunnTQ/9NnyD/0aW3wtco78XA288mvCCB+JwjzSuCDlYwH5FBGXyTGU3TRBctQrSrIcIUHKTsMM855+oUN6PEsMB13h5XuGgb7PzcojkEpUljuaqit7bRiJWRXP5WyGFOO3Y8JC0zzE2VjKAL+qBx1MO6CQwhohYY7AdOz0y4v57ufsxyYZnGAUh/wUSZNM1tfSijobl3goig/iFhPXQrp+lSY06Kc2WDIOwG79NoHRPnGxAc+GKuqxhwIDAQABAoIBAGMlsm7IMG3gt4XG+rlDaTkw67kd8uy+7fInzBlyrkSMeCpixq+2dGWo0RWhfdRK2Llzu78PQmU5mPtKd/4oVdX5i5CfCqScpHdZwPjcuzoiXI21jYGObcIE9rq1vkaBJcLr2GNyR6yrXuA/O7efhjo+PveqInyXEUAE2vIVBF7tqhaZHjP7KWYkE5r4eHc+3CRmPyUaRL8luzQ70sXKdBT2fyZtSsCZus7hKwtsw+uztV9B3tV04WpN48J2C3B1hjscSlVy/T4z9gVABVSmZ0CgUF6Ve6Ams6bCiW1GvV9EnNpRRJWoPP44ezWrSVXRLEq8dVkrIiHII+AmyEpObdECgYEAzanD4E96Yoq5o+4P9XlDSu6ewrDd9hszfafSqz/plzImsB//xAK58TN4OkkGGOD/6p0KGq3uOW7fysP+d6AmfK5xYpCoRp8esX5CRJlJKZAcOXXqFhqpIXPGPp4ECKXDbBN+gA6RfosW7Za27oHGt3jb+pIqZx4B0CCH/8T4wz8CgYEAz2e4E/DWOEPeIxIrozSxZ8Tm4uOaLYnpF96LGsd7pcoy2bUno1513/8ljei9FG+q+3QSvMX3xiZ2g59Djo/EFgCxuEEDGdl9kQKzbxlEOjdMRp/45MWUrJneRIaZIADa1VktInvfgvyw5ykkkwIaejXgsgE4jxYEZ6mgfFPZJ7kCgYEAjC+r1LpYMdNdtuAPAMQnmmwMy6jDo33nGz/J2fE2yFjnBibJsFIrbL4otZIRFeRIjSN+P7FPMNbitNPkIKaJlXNS1lzEV18fDN2DTj8uH6ablo9JgMergaHo/8W1+i8DhifDkoZbpmYousqA5xKO4YEAcUwwmxlZIwUJrs4UwwECgYApZcFnMYYAmwNGqsTNAaJ7XODc3qMSfdlpq0DGqpRyhgZaT/9Ga39Jo2rChXbgEl3lbMikpzsKcPjs0qgwV1/DKIZQiaFt0mxuukRIY5mrqQfvfNS4DPvc66Vith7wNVt4hCEvJdx6D6fbq+mh8iSIyiI8RWQdoz5j1oay6iWI+QKBgQCveTt3CXZnOaNUzG761sowXXxdjB0YoPLcMrF0VrhkMfVkbty535l8Wis9Wd5XUImKWmcXLsRVFs4jV/YX+KpszNgRuefTJ2p8V/RB0Fwh7nWveC3n8vNgN6AVwEAs7DEXqdVXzNq6l3dyNxX1L3MVPeVbPTEfZUNvKj6ed22tfA=="
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

	Context("Upstream validation", func() {
		It("should validate Gateway", func() {
			gateway := fmt.Sprintf(`
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: Gateway
metadata:
  name: kuma-validate
  namespace: %s
  annotations:
    kuma.io/mesh: %s
spec:
  gatewayClassName: kuma
  listeners:
  - name: proxy
    port: 10080
    protocol: TCP
    hostname: xyz.io
`, namespace, meshName)

			err := k8s.KubectlApplyFromStringE(kubernetes.Cluster.GetTesting(), kubernetes.Cluster.GetKubectlOptions(), gateway)
			Expect(err).To(HaveOccurred())
		})
	})
}
