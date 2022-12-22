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

	BeforeAll(func() {
		clusters, err := NewK8sClusters(
			[]string{Kuma1},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		cluster = clusters.GetCluster(Kuma1)
		err = NewClusterSetup().
			Install(Kuma(core.Standalone)).
			Install(NamespaceWithSidecarInjection(TestNamespace)).
			Install(DemoClientK8s("default", TestNamespace)).
			Install(externalservice.Install(externalservice.HttpServer, []string{})).
			Install(externalservice.Install(externalservice.HttpsServer, []string{})).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())

		err = YamlK8s(fmt.Sprintf(meshDefaulMtlsOn, "false", "false"))(cluster)
		Expect(err).ToNot(HaveOccurred())

		clientPodName, err = PodNameOfApp(cluster, "demo-client", TestNamespace)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterAll(func() {
		err := cluster.DeleteNamespace(TestNamespace)
		Expect(err).ToNot(HaveOccurred())

		err = cluster.DeleteKuma()
		Expect(err).ToNot(HaveOccurred())

		err = cluster.DismissCluster()
		Expect(err).ToNot(HaveOccurred())
	})

	trafficBlocked := func() error {
		_, _, err := cluster.Exec(TestNamespace, clientPodName, "demo-client",
			"curl", "-v", "-m", "3", "--fail", "http://externalservice-http-server.externalservice-namespace:10080")
		return err
	}

	It("should route to external-service", func() {
		// given Mesh with passthrough enabled and no egress
		err := YamlK8s(fmt.Sprintf(meshDefaulMtlsOn, "true", "false"))(cluster)
		Expect(err).ToNot(HaveOccurred())

		// then communication outside of the Mesh works
		_, stderr, err := cluster.ExecWithRetries(TestNamespace, clientPodName, "demo-client",
			"curl", "-v", "-m", "3", "--fail", "http://externalservice-http-server.externalservice-namespace:10080")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))

		// when passthrough is disabled on the Mesh and no egress
		err = YamlK8s(fmt.Sprintf(meshDefaulMtlsOn, "false", "false"))(cluster)
		Expect(err).ToNot(HaveOccurred())

		// then communication outside of the Mesh works
		_, stderr, err = cluster.ExecWithRetries(TestNamespace, clientPodName, "demo-client",
			"curl", "-v", "-m", "3", "--fail", "http://externalservice-http-server.externalservice-namespace:10080")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))

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
			_, _, err = cluster.Exec(TestNamespace, clientPodName, "demo-client",
				"curl", "-v", "-m", "3", "--fail", "http://external-service.mesh:10080")
			return err
		}, "30s", "1s").Should(HaveOccurred())
	})
}
