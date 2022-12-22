package externalservices

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/kubernetes/env"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/externalservice"
)

func ExternalServicesWithoutEgress() {
	meshName := "external-services-no-egress"
	clientNamespace := "client-external-services-no-egress"

	var clientPodName string
	meshDefaulMtlsOn := `
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: %s
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

	trafficBlocked := func() error {
		_, _, err := env.Cluster.Exec(clientNamespace, clientPodName, "demo-client",
			"curl", "-v", "-m", "3", "--fail", "http://externalservice-http-server.externalservice-namespace:10080")
		return err
	}

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(YamlK8s(fmt.Sprintf(meshDefaulMtlsOn, meshName, "true", "false"))).
			Install(NamespaceWithSidecarInjection(clientNamespace)).
			Install(externalservice.Install(externalservice.HttpServer, []string{})).
			Install(externalservice.Install(externalservice.HttpsServer, []string{})).
			Install(DemoClientK8s(meshName, clientNamespace)).
			Setup(env.Cluster)
		Expect(err).ToNot(HaveOccurred())

		clientPodName, err = PodNameOfApp(env.Cluster, "demo-client", clientNamespace)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterAll(func() {
		Expect(env.Cluster.TriggerDeleteNamespace(clientNamespace)).To(Succeed())
		Expect(env.Cluster.TriggerDeleteNamespace("externalservice-namespace")).To(Succeed())
		Expect(env.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should route to external-service", func() {
		// given default Mesh with passthrough enabled and no egress

		// then communication outside of the Mesh works
		_, stderr, err := env.Cluster.ExecWithRetries(clientNamespace, clientPodName, "demo-client",
			"curl", "-v", "-m", "3", "--fail", "http://externalservice-http-server.externalservice-namespace:10080")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))

		// when passthrough is disabled on the Mesh and no egress
		err = YamlK8s(fmt.Sprintf(meshDefaulMtlsOn, meshName, "false", "false"))(env.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// then communication outside of the Mesh works
		_, stderr, err = env.Cluster.ExecWithRetries(clientNamespace, clientPodName, "demo-client",
			"curl", "-v", "-m", "3", "--fail", "http://externalservice-http-server.externalservice-namespace:10080")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))

		// when apply external service
		err = YamlK8s(fmt.Sprintf(externalService,
			"1", "1",
			"externalservice-http-server.externalservice-namespace.svc.cluster.local", 10080, // .svc.cluster.local is needed, otherwise Kubernetes will resolve this to the real IP
			"false"))(env.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// and passthrough is disabled on the Mesh and zone egress enabled
		err = YamlK8s(fmt.Sprintf(meshDefaulMtlsOn, meshName, "false", "true"))(env.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// then accessing the external service is no longer possible
		Eventually(trafficBlocked, "30s", "1s").Should(HaveOccurred())
	})

	It("should route to external-service over tls", func() {
		// given Mesh with passthrough enabled and with zone egress
		err := YamlK8s(fmt.Sprintf(meshDefaulMtlsOn, meshName, "true", "true"))(env.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// when apply external service
		err = YamlK8s(fmt.Sprintf(externalService,
			"2", "2",
			"externalservice-https-server.externalservice-namespace.svc.cluster.local", 10080,
			"true"))(env.Cluster)
		Expect(err).ToNot(HaveOccurred())

		Eventually(func() error {
			_, _, err = env.Cluster.Exec(clientNamespace, clientPodName, "demo-client",
				"curl", "-v", "-m", "3", "--fail", "http://external-service.mesh:10080")
			return err
		}, "30s", "1s").Should(HaveOccurred())
	})
}
