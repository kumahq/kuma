package externalservices

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func PermissiveMTLS() {
	meshName := "perm-external-services"
	namespace := "perm-external-services"
	clientNamespace := "perm-client-external-services"

	mesh := `
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: perm-external-services
spec:
  mtls:
    enabledBackend: ca-1
    backends:
      - name: ca-1
        type: builtin
        mode: PERMISSIVE
  networking:
    outbound:
      passthrough: false
  routing:
    zoneEgress: true
`

	tlsExternalService := `
apiVersion: kuma.io/v1alpha1
kind: ExternalService
mesh: perm-external-services
metadata:
  name: perm-tls-external-service
spec:
  tags:
    kuma.io/service: perm-tls-external-service
    kuma.io/protocol: http
  networking:
    address: perm-tls-external-service.perm-external-services.svc.cluster.local:80 # .svc.cluster.local is needed, otherwise Kubernetes will resolve this to the real IP
    tls:
      enabled: true
`

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(YamlK8s(mesh)).
			Install(YamlK8s(tlsExternalService)).
			Install(Namespace(namespace)).
			Install(NamespaceWithSidecarInjection(clientNamespace)).
			Install(Parallel(
				democlient.Install(democlient.WithNamespace(clientNamespace), democlient.WithMesh(meshName)),
				testserver.Install(
					testserver.WithNamespace(namespace),
					testserver.WithEchoArgs("--tls", "--crt=/kuma/server.crt", "--key=/kuma/server.key"),
					testserver.WithName("perm-tls-external-service"),
					testserver.WithoutProbes(), // not compatible with TLS
				),
			)).
			Install(MeshTrafficPermissionAllowAllKubernetes(meshName)).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugKube(kubernetes.Cluster, meshName, namespace, clientNamespace)
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(clientNamespace)).To(Succeed())
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should access external service", func() {
		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				kubernetes.Cluster, "demo-client", "http://perm-tls-external-service.mesh",
				client.FromKubernetesPod(clientNamespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
		}, "30s", "1s").Should(Succeed())
	})
}
