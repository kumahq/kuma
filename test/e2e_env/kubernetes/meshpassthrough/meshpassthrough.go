package meshpassthrough

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/plugins/policies/meshpassthrough/api/v1alpha1"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func MeshPassthrough() {
	const meshName = "mesh-passthrough"
	const mesNamespace = "mesh-passthrough-mes"
	const namespace = "mesh-passthrough"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshKubernetes(meshName)).
			Install(Namespace(mesNamespace)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(democlient.Install(
				democlient.WithNamespace(namespace),
				democlient.WithMesh(meshName),
			)).
			Install(testserver.Install(
				testserver.WithNamespace(mesNamespace),
				testserver.WithName("external-service"),
			)).
			Install(testserver.Install(
				testserver.WithNamespace(mesNamespace),
				testserver.WithName("another-external-service"),
			)).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugKube(kubernetes.Cluster, meshName, namespace)
	})

	E2EAfterEach(func() {
		Expect(DeleteMeshResources(kubernetes.Cluster, meshName, v1alpha1.MeshPassthroughResourceTypeDescriptor)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should control passthrough cluster", func() {
		// given
		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				kubernetes.Cluster, "demo-client", "external-service.mesh-passthrough-mes.svc.cluster.local:80",
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
		}, "30s", "1s").Should(Succeed())

		meshPassthrough := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1 
kind: MeshPassthrough
metadata:
  name: disable-passthrough
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: MeshSubset
    tags:
      kuma.io/service: demo-client_mesh-passthrough_svc
  default:
    enabled: false
`, Config.KumaNamespace, meshName)

		// when
		err := kubernetes.Cluster.Install(YamlK8s(meshPassthrough))


		time.Sleep(1000*time.Hour)
		// then
		Expect(err).ToNot(HaveOccurred())
		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				kubernetes.Cluster, "demo-client", "external-service.mesh-passthrough-mes.svc.cluster.local:80",
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).To(HaveOccurred())
		}, "30s", "1s").MustPassRepeatedly(3).Should(Succeed())
	})
}
