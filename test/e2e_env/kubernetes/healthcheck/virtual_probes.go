package healthcheck

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func VirtualProbes() {
	const name = "test-server"
	const namespace = "virtual-probes"
	const mesh = "virtual-probes"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MTLSMeshKubernetes(mesh)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh(mesh),
				testserver.WithName(name),
			)).
			Setup(kubernetes.Cluster)
		Expect(err).To(Succeed())
	})
	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(mesh)).To(Succeed())
	})

	It("should deploy test-server with probes", func() {
		// Sample pod readiness to ensure they stay ready to at least 10sec.
		Expect(WaitNumPods(namespace, 1, name)(kubernetes.Cluster)).To(Succeed())
		Expect(WaitPodsAvailable(namespace, name)(kubernetes.Cluster)).To(Succeed())

		Consistently(func(g Gomega) {
			g.Expect(WaitNumPods(namespace, 1, name)(kubernetes.Cluster)).To(Succeed())
			g.Expect(WaitPodsAvailable(namespace, name)(kubernetes.Cluster)).To(Succeed())
		}, "10s", "1s").Should(Succeed())
	})
}
