package externalservices

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/externalservice"
)

func ExternalServicesOnKubernetesWithoutEgress() {
	namespace := "external-service-k8s-no-egress"
	meshDefaulMtlsOn := `
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: default
spec:
  mtls:
    enabledBackend: ca-1
    backends:
      - name: ca-1
        type: builtin
  networking:
    outbound:
      passthrough: %s
  routing:
    zoneEgress: %s
`

	externalService := `
apiVersion: kuma.io/v1alpha1
kind: ExternalService
mesh: default
metadata:
  name: external-service-%s
spec:
  tags:
    kuma.io/service: external-service
    kuma.io/protocol: http
    id: "%s"
  networking:
    address: %s:%d
    tls:
      enabled: %s
`
	es1 := "1"
	es2 := "2"

	var cluster Cluster
	var clientPodName string

	BeforeEach(func() {
		clusters, err := NewK8sClusters(
			[]string{Kuma1},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		cluster = clusters.GetCluster(Kuma1)
		err = NewClusterSetup().
			Install(Kuma(core.Standalone)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(DemoClientK8s("default", namespace)).
			Install(externalservice.Install(externalservice.HttpServer, []string{})).
			Install(externalservice.Install(externalservice.HttpsServer, []string{})).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())

		err = YamlK8s(fmt.Sprintf(meshDefaulMtlsOn, "false", "false"))(cluster)
		Expect(err).ToNot(HaveOccurred())

		clientPodName, err = PodNameOfApp(cluster, "demo-client", namespace)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterEach(func() {
		Expect(cluster.DeleteNamespace(namespace)).ToNot(HaveOccurred())
		Expect(cluster.DeleteNamespace(fmt.Sprintf("%s-%s", externalservice.DeploymentName, "namespace"))).ToNot(HaveOccurred())
		Expect(cluster.DeleteKuma()).ToNot(HaveOccurred())
		Expect(cluster.DismissCluster()).ToNot(HaveOccurred())
	})

	trafficBlocked := func() error {
		_, _, err := cluster.Exec(namespace, clientPodName, "demo-client",
			"curl", "-v", "-m", "3", "--fail", "http://externalservice-http-server.externalservice-namespace:10080")
		return err
	}

	It("should route to external-service", func() {
		// given Mesh with passthrough enabled and no egress
		err := YamlK8s(fmt.Sprintf(meshDefaulMtlsOn, "true", "false"))(cluster)
		Expect(err).ToNot(HaveOccurred())

		// then communication outside of the Mesh works
		Eventually(func(g Gomega) {
			_, stderr, err := cluster.Exec(namespace, clientPodName, "demo-client",
				"curl", "-v", "-m", "3", "--fail", "http://externalservice-http-server.externalservice-namespace:10080")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))
		}, "1m", "3s").Should(Succeed())

		// when passthrough is disabled on the Mesh and no egress
		err = YamlK8s(fmt.Sprintf(meshDefaulMtlsOn, "false", "false"))(cluster)
		Expect(err).ToNot(HaveOccurred())

		// then communication outside of the Mesh works
		Eventually(func(g Gomega) {
			_, stderr, err := cluster.Exec(namespace, clientPodName, "demo-client",
				"curl", "-v", "-m", "3", "--fail", "http://externalservice-http-server.externalservice-namespace:10080")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))
		}, "1m", "3s").Should(Succeed())

		// when apply external service
		err = YamlK8s(fmt.Sprintf(externalService,
			es1, es1,
			"externalservice-http-server.externalservice-namespace.svc.cluster.local", 10080, // .svc.cluster.local is needed, otherwise Kubernetes will resolve this to the real IP
			"false"))(cluster)
		Expect(err).ToNot(HaveOccurred())

		// and passthrough is disabled on the Mesh and zone egress enabled
		err = YamlK8s(fmt.Sprintf(meshDefaulMtlsOn, "false", "true"))(cluster)
		Expect(err).ToNot(HaveOccurred())

		// then accessing the external service is no longer possible
		Eventually(trafficBlocked, "30s", "1s").Should(HaveOccurred())
	})

	It("should route to external-service over tls", func() {
		// given Mesh with passthrough enabled and with zone egress
		err := YamlK8s(fmt.Sprintf(meshDefaulMtlsOn, "true", "true"))(cluster)
		Expect(err).ToNot(HaveOccurred())

		// when apply external service
		err = YamlK8s(fmt.Sprintf(externalService,
			es2, es2,
			"externalservice-https-server.externalservice-namespace.svc.cluster.local", 10080,
			"true"))(cluster)
		Expect(err).ToNot(HaveOccurred())

		Eventually(func() error {
			_, _, err = cluster.Exec(namespace, clientPodName, "demo-client",
				"curl", "-v", "-m", "3", "--fail", "http://external-service.mesh:10080")
			return err
		}, "30s", "1s").Should(HaveOccurred())
	})
}
