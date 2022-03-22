package gatewayapi

import (
	"net"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	client "github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func GatewayAPICRDs(cluster Cluster) error {
	out, err := k8s.RunKubectlAndGetOutputE(
		cluster.GetTesting(),
		cluster.GetKubectlOptions(),
		"kustomize", "github.com/kubernetes-sigs/gateway-api/config/crd?ref=v0.4.0",
	)
	if err != nil {
		return err
	}

	return k8s.KubectlApplyFromStringE(cluster.GetTesting(), cluster.GetKubectlOptions(), out)
}

const gatewayClass = `
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: GatewayClass
metadata:
  name: kuma
spec:
  controllerName: gateways.kuma.io/controller
`

var cluster *K8sCluster

var _ = E2EBeforeSuite(func() {
	if Config.IPV6 {
		return // KIND which is used for IPV6 tests does not support load balancer that is used in this test.
	}

	cluster = NewK8sCluster(NewTestingT(), Kuma1, Silent)

	err := NewClusterSetup().
		Install(GatewayAPICRDs).
		Install(Kuma(config_core.Standalone,
			WithCtlOpts(map[string]string{"--experimental-meshgateway": "true"}),
			WithEnv("KUMA_EXPERIMENTAL_GATEWAY_API", "true"),
		)).
		Install(NamespaceWithSidecarInjection(TestNamespace)).
		Install(testserver.Install(
			testserver.WithName("test-server-1"),
			testserver.WithNamespace(TestNamespace),
			testserver.WithArgs("echo", "--instance", "test-server-1"),
		)).
		Install(testserver.Install(
			testserver.WithName("test-server-2"),
			testserver.WithNamespace(TestNamespace),
			testserver.WithArgs("echo", "--instance", "test-server-2"),
		)).
		Install(YamlK8s(gatewayClass)).
		Setup(cluster)
	Expect(err).ToNot(HaveOccurred())

	E2EDeferCleanup(func() {
		Expect(cluster.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(cluster.DeleteKuma()).To(Succeed())
		Expect(cluster.DismissCluster()).To(Succeed())
	})
})

func GatewayAPI() {
	if Config.IPV6 {
		return // KIND which is used for IPV6 tests does not support load balancer that is used in this test.
	}
	Context("HTTP Gateway", func() {
		gateway := `
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: Gateway
metadata:
  name: kuma
  namespace: kuma-test
spec:
  gatewayClassName: kuma
  listeners:
  - name: proxy
    port: 8080
    protocol: HTTP`

		GatewayAddress := func() string {
			var ip string
			Eventually(func(g Gomega) {
				out, err := k8s.RunKubectlAndGetOutputE(
					cluster.GetTesting(),
					cluster.GetKubectlOptions(TestNamespace),
					"get", "gateway", "kuma", "-ojsonpath={.status.addresses[0].value}",
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(out).ToNot(BeEmpty())
				ip = out
			}, "60s", "1s").Should(Succeed(), "could not get a LoadBalancer IP of the Gateway")
			return net.JoinHostPort(ip, "8080")
		}

		var address string

		BeforeEach(func() {
			err := k8s.RunKubectlE(cluster.GetTesting(), cluster.GetKubectlOptions(), "delete", "gateway", "--all")
			Expect(err).ToNot(HaveOccurred())
			Expect(YamlK8s(gateway)(cluster)).To(Succeed())
			address = GatewayAddress()
		})

		It("should route the traffic to test-server by path", func() {
			// given
			route := `
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: HTTPRoute
metadata:
  name: test-server-paths
  namespace: kuma-test
spec:
  parentRefs:
  - name: kuma
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
        value: /2`

			// when
			err := YamlK8s(route)(cluster)

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
			routes := `
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: HTTPRoute
metadata:
  name: test-server-1
  namespace: kuma-test
spec:
  parentRefs:
  - name: kuma
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
  namespace: kuma-test
spec:
  parentRefs:
  - name: kuma
  hostnames:
  - "test-server-2.com"
  rules:
  - backendRefs:
    - name: test-server-2
      port: 80
`

			// when
			err := YamlK8s(routes)(cluster)

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
	})
}
