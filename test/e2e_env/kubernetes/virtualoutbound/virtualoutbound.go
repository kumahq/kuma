package virtualoutbound

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func VirtualOutbound() {
	namespace := "virtual-outbounds"
	meshName := "virtual-outbounds"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshKubernetes(meshName)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Parallel(
				democlient.Install(democlient.WithNamespace(namespace), democlient.WithMesh(meshName)),
				testserver.Install(
					testserver.WithMesh(meshName),
					testserver.WithNamespace(namespace),
					testserver.WithStatefulSet(),
					testserver.WithReplicas(2),
				),
			)).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugKube(kubernetes.Cluster, meshName, namespace)
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	BeforeEach(func() {
		Expect(DeleteMeshResources(kubernetes.Cluster, meshName, mesh.VirtualOutboundResourceTypeDescriptor)).To(Succeed())
	})

	It("simple virtual outbound", func() {
		virtualOutboundAll := `
apiVersion: kuma.io/v1alpha1
kind: VirtualOutbound
mesh: virtual-outbounds
metadata:
  name: instance
spec:
  selectors:
  - match:
      kuma.io/service: "*"
  conf:
    host: "{{.svc}}.foo"
    port: "8080"
    parameters:
    - name: "svc"
      tagKey: "kuma.io/service"
`
		err := YamlK8s(virtualOutboundAll)(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// Succeed with virtual-outbound
		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(
				kubernetes.Cluster, "demo-client", "test-server_virtual-outbounds_svc_80.foo:8080",
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(ContainSubstring("test-server"))
		}, "30s", "1s").Should(Succeed())
	})

	It("virtual outbounds on statefulSet", func() {
		virtualOutboundAll := `
apiVersion: kuma.io/v1alpha1
kind: VirtualOutbound
mesh: virtual-outbounds
metadata:
  name: instance
spec:
  selectors:
  - match:
      kuma.io/service: "*"
      statefulset.kubernetes.io/pod-name: "*"
  conf:
    host: "{{.svc}}.{{.inst}}"
    port: "8080"
    parameters:
    - name: "svc"
      tagKey: "kuma.io/service"
    - name: "inst"
      tagKey: "statefulset.kubernetes.io/pod-name"
`
		err := YamlK8s(virtualOutboundAll)(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())

		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(
				kubernetes.Cluster, "demo-client", "test-server_virtual-outbounds_svc_80.test-server-0:8080",
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(Equal("test-server-0"))
		}, "30s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(
				kubernetes.Cluster, "demo-client", "test-server_virtual-outbounds_svc_80.test-server-1:8080",
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(ContainSubstring("test-server-1"))
		}, "30s", "1s").Should(Succeed())
	})
}
