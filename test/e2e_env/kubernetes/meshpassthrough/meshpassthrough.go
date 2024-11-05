package meshpassthrough

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/plugins/policies/meshpassthrough/api/v1alpha1"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

const curlRecvError = 56

func MeshPassthrough() {
	const meshName = "mesh-passthrough"
	const mesNamespace = "mesh-passthrough-mes"
	const namespace = "mesh-passthrough"
	var anotherEsIP string
	var notAccessibleEsIP string

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshKubernetes(meshName)).
			Install(Namespace(mesNamespace)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Parallel(
				democlient.Install(
					democlient.WithNamespace(namespace),
					democlient.WithMesh(meshName),
				),
				testserver.Install(
					testserver.WithNamespace(mesNamespace),
					testserver.WithName("external-service"),
				),
				testserver.Install(
					testserver.WithNamespace(mesNamespace),
					testserver.WithName("another-external-service"),
				),
				testserver.Install(
					testserver.WithNamespace(mesNamespace),
					testserver.WithName("not-accessible-external-service"),
				),
			)).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())
		anotherEsIP, err = PodIPOfApp(kubernetes.Cluster, "another-external-service", mesNamespace)
		Expect(err).ToNot(HaveOccurred())
		notAccessibleEsIP, err = PodIPOfApp(kubernetes.Cluster, "not-accessible-external-service", mesNamespace)
		Expect(err).ToNot(HaveOccurred())
	})

	curlAddress := func(address string) string {
		if Config.IPV6 {
			return fmt.Sprintf("[%s]", address)
		}
		return address
	}

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
    proxyTypes: ["Sidecar"]
    tags:
      kuma.io/service: demo-client_mesh-passthrough_svc
  default:
    passthroughMode: None
`, Config.KumaNamespace, meshName)

		// when
		err := kubernetes.Cluster.Install(YamlK8s(meshPassthrough))

		// then
		Expect(err).ToNot(HaveOccurred())
		Eventually(func(g Gomega) {
			resp, err := client.CollectFailure(
				kubernetes.Cluster, "demo-client", "external-service.mesh-passthrough-mes.svc.cluster.local:80",
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Exitcode).To(Equal(curlRecvError))
		}, "30s", "1s").MustPassRepeatedly(3).Should(Succeed())
	})

	It("should control traffic to domains and IPs", func() {
		// given
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
    proxyTypes: ["Sidecar"]
    tags:
      kuma.io/service: demo-client_mesh-passthrough_svc
  default:
    passthroughMode: None
`, Config.KumaNamespace, meshName)

		// when
		err := kubernetes.Cluster.Install(YamlK8s(meshPassthrough))

		// then
		Expect(err).ToNot(HaveOccurred())
		Eventually(func(g Gomega) {
			resp, err := client.CollectFailure(
				kubernetes.Cluster, "demo-client", "external-service.mesh-passthrough-mes.svc.cluster.local:80",
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Exitcode).To(Equal(curlRecvError))
		}, "30s", "1s").MustPassRepeatedly(3).Should(Succeed())

		meshPassthrough = fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1 
kind: MeshPassthrough
metadata:
  name: allow-specified
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: MeshSubset
    proxyTypes: ["Sidecar"]
    tags:
      kuma.io/service: demo-client_mesh-passthrough_svc
  default:
    passthroughMode: Matched
    appendMatch:
    - type: Domain
      value: external-service.mesh-passthrough-mes.svc.cluster.local
      port: 80
      protocol: http
    - type: IP
      value: %s
      port: 80
      protocol: http
`, Config.KumaNamespace, meshName, anotherEsIP)

		// when
		err = kubernetes.Cluster.Install(YamlK8s(meshPassthrough))

		// then
		Expect(err).ToNot(HaveOccurred())
		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				kubernetes.Cluster, "demo-client", "external-service.mesh-passthrough-mes.svc.cluster.local:80",
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
		}, "30s", "1s").Should(Succeed())

		// and ip address
		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				kubernetes.Cluster, "demo-client", curlAddress(anotherEsIP),
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
		}, "30s", "1s").Should(Succeed())

		// and shouldn't access not-accessible-external-service
		Eventually(func(g Gomega) {
			resp, err := client.CollectFailure(
				kubernetes.Cluster, "demo-client", curlAddress(notAccessibleEsIP),
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.ResponseCode).To(Equal(503))
		}, "30s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			resp, err := client.CollectFailure(
				kubernetes.Cluster, "demo-client", "not-accessible-external-service.mesh-passthrough-mes.svc.cluster.local:80",
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.ResponseCode).To(Equal(503))
		}, "30s", "1s").Should(Succeed())
	})
}
