package externalservices

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
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
      passthrough: false
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

	BeforeEach(func() {
		cluster = NewK8sCluster(NewTestingT(), Kuma1, Silent)
		err := NewClusterSetup().
			Install(Kuma(core.Zone)).
			Install(YamlK8s(fmt.Sprintf(meshDefaulMtlsOn, "false"))).
			Install(MeshTrafficPermissionAllowAllKubernetes("default")).
			Install(NamespaceWithSidecarInjection(TestNamespace)).
			Install(democlient.Install(democlient.WithNamespace(TestNamespace), democlient.WithMesh("default"))).
			Install(Namespace(externalServicesNamespace)).
			Install(testserver.Install(
				testserver.WithNamespace(externalServicesNamespace),
				testserver.WithName("external-service"),
			)).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterEach(func() {
		Expect(cluster.TriggerDeleteNamespace(externalServicesNamespace)).To(Succeed())
		Expect(cluster.TriggerDeleteNamespace(TestNamespace)).To(Succeed())
		Expect(cluster.WaitNamespaceDelete(externalServicesNamespace)).To(Succeed())
		Expect(cluster.WaitNamespaceDelete(TestNamespace)).To(Succeed())

		Expect(cluster.DeleteKuma()).To(Succeed())
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	It("should block the traffic when zone egress is not deployed", func() {
		// when external service is defined without passthrough and zone routing
		Expect(cluster.Install(YamlK8s(externalService))).To(Succeed())

		// then communication outside the Mesh works, because it goes directly from the client to a service
		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				cluster, "demo-client", "external-service.external-services.svc.cluster.local:80",
				client.FromKubernetesPod(TestNamespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
		}, "30s", "1s").Should(Succeed())

		// when passthrough is disabled on the Mesh and zone egress enabled
		Expect(cluster.Install(YamlK8s(fmt.Sprintf(meshDefaulMtlsOn, "true")))).To(Succeed())

		// then accessing the external service is no longer possible, because zone egress is not deployed
		Eventually(func(g Gomega) {
			response, err := client.CollectFailure(
				cluster, "demo-client", "external-service.external-services.svc.cluster.local:80",
				client.FromKubernetesPod(TestNamespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Exitcode).To(Or(Equal(56), Equal(7), Equal(28)))
		}, "30s", "1s").Should(Succeed())
	})
}
