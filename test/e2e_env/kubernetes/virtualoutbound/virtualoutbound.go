package virtualoutbound

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/env"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func VirtualOutbound() {
	namespace := "virtual-outbounds"
	meshName := "virtual-outbounds"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshKubernetes(meshName)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(DemoClientK8s(meshName, namespace)).
			Install(testserver.Install(
				testserver.WithMesh(meshName),
				testserver.WithNamespace(namespace),
				testserver.WithStatefulSet(true),
				testserver.WithReplicas(2),
			)).
			Setup(env.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterAll(func() {
		Expect(env.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(meshName))
	})

	BeforeEach(func() {
		Expect(DeleteMeshResources(env.Cluster, meshName, mesh.VirtualOutboundResourceTypeDescriptor)).To(Succeed())
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
		err := YamlK8s(virtualOutboundAll)(env.Cluster)
		Expect(err).ToNot(HaveOccurred())
		// when client sends requests to server
		clientPodName, err := PodNameOfApp(env.Cluster, "demo-client", namespace)
		Expect(err).ToNot(HaveOccurred())

		// Succeed with virtual-outbound
		Eventually(func(g Gomega) {
			stdout, stderr, err := env.Cluster.Exec(namespace, clientPodName, "demo-client",
				"curl", "-v", "-m", "3", "--fail", "test-server_virtual-outbounds_svc_80.foo:8080")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))
			g.Expect(stdout).To(ContainSubstring(`"instance":"test-server`))
		}).Should(Succeed())
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
		err := YamlK8s(virtualOutboundAll)(env.Cluster)
		Expect(err).ToNot(HaveOccurred())
		// when client sends requests to server
		clientPodName, err := PodNameOfApp(env.Cluster, "demo-client", namespace)
		Expect(err).ToNot(HaveOccurred())

		Eventually(func(g Gomega) {
			stdout, stderr, err := env.Cluster.Exec(namespace, clientPodName, "demo-client",
				"curl", "-v", "-m", "3", "--fail", "test-server_virtual-outbounds_svc_80.test-server-0:8080")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))
			g.Expect(stdout).To(ContainSubstring(`"instance":"test-server-0"`))
		}, "30s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			stdout, stderr, err := env.Cluster.Exec(namespace, clientPodName, "demo-client",
				"curl", "-v", "-m", "3", "--fail", "test-server_virtual-outbounds_svc_80.test-server-1:8080")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))
			g.Expect(stdout).To(ContainSubstring(`"instance":"test-server-1"`))
		}, "30s", "1s").Should(Succeed())
	})
}
