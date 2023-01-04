package healthcheck

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/kubernetes/env"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
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
			Setup(env.Cluster)
		Expect(err).To(Succeed())
	})
	E2EAfterAll(func() {
		Expect(env.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(mesh))
	})

	It("should deploy test-server with probes", func() {
		// Sample pod readiness to ensure they stay ready to at least 10sec.
		Expect(WaitNumPods(namespace, 1, name)(env.Cluster)).To(Succeed())
		Expect(WaitPodsAvailable(namespace, name)(env.Cluster)).To(Succeed())

		Consistently(func(g Gomega) {
			g.Expect(WaitNumPods(namespace, 1, name)(env.Cluster)).To(Succeed())
			g.Expect(WaitPodsAvailable(namespace, name)(env.Cluster)).To(Succeed())
		}, "10s", "1s").Should(Succeed())
	})
}
