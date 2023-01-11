package externalservices

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func ExternalServicesOnKubernetesWithoutEgress() {
	const meshDefaulMtlsOn = `
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

	const externalService = `
apiVersion: kuma.io/v1alpha1
kind: ExternalService
mesh: default
metadata:
  name: external-service-1
spec:
  tags:
    kuma.io/service: external-service
    kuma.io/protocol: http
  networking:
    address: external-service.external-services.svc.cluster.local:80 # .svc.cluster.local is needed, otherwise Kubernetes will resolve this to the real IP
`
	const externalServicesNamespace = "external-services"

	var cluster *K8sCluster
	var clientPodName string

	BeforeEach(func() {
		cluster = NewK8sCluster(NewTestingT(), Kuma1, Silent)
		err := NewClusterSetup().
			Install(Kuma(core.Standalone)).
			Install(YamlK8s(fmt.Sprintf(meshDefaulMtlsOn, "false", "false"))).
			Install(NamespaceWithSidecarInjection(TestNamespace)).
			Install(DemoClientK8s("default", TestNamespace)).
			Install(Namespace(externalServicesNamespace)).
			Install(testserver.Install(
				testserver.WithNamespace(externalServicesNamespace),
				testserver.WithName("external-service"),
			)).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())

		clientPodName, err = PodNameOfApp(cluster, "demo-client", TestNamespace)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterEach(func() {
		Expect(cluster.TriggerDeleteNamespace(externalServicesNamespace)).To(Succeed())
		Expect(cluster.TriggerDeleteNamespace(TestNamespace)).To(Succeed())
		cluster.WaitNamespaceDelete(externalServicesNamespace)
		cluster.WaitNamespaceDelete(TestNamespace)

		Expect(cluster.DeleteKuma()).To(Succeed())
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	trafficBlocked := func(g Gomega) {
		_, _, err := cluster.Exec(TestNamespace, clientPodName, "demo-client",
			"curl", "-v", "-m", "3", "--fail", "external-service.external-services.svc.cluster.local:80")
		g.Expect(err).To(HaveOccurred())
	}

	trafficAllowed := func(g Gomega) {
		_, stderr, err := cluster.Exec(TestNamespace, clientPodName, "demo-client",
			"curl", "-v", "-m", "3", "--fail", "external-service.external-services.svc.cluster.local:80")
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))
	}

	It("should route to external-service", func() {
		// given the mesh with passthrough enabled and no egress disabled
		Expect(cluster.Install(YamlK8s(fmt.Sprintf(meshDefaulMtlsOn, "true", "false")))).To(Succeed())

		// then communication outside of the Mesh works, because it goes through passthrough
		Eventually(trafficAllowed, "30s", "1s").Should(Succeed())

		// when passthrough is disabled
		Expect(cluster.Install(YamlK8s(fmt.Sprintf(meshDefaulMtlsOn, "false", "false")))).To(Succeed())

		// then accessing the external service is no longer possible
		Eventually(trafficBlocked, "30s", "1s").Should(Succeed())

		// when apply external service
		Expect(cluster.Install(YamlK8s(externalService))).To(Succeed())

		// then communication outside of the Mesh works, because it goes directly from the client to a service
		Eventually(trafficAllowed, "30s", "1s").Should(Succeed())

		// when passthrough is disabled on the Mesh and zone egress enabled
		Expect(cluster.Install(YamlK8s(fmt.Sprintf(meshDefaulMtlsOn, "false", "true")))).To(Succeed())

		// then accessing the external service is no longer possible, because zone egress is not deployed
		Eventually(trafficBlocked, "30s", "1s").Should(Succeed())
	})
}
