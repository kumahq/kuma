package reachableservices

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func ReachableServices() {
	meshName := "reachable-svc"
	namespace := "reachable-svc"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MTLSMeshKubernetes(meshName)).
			Install(MeshTrafficPermissionAllowAllKubernetes(meshName)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(testserver.Install(
				testserver.WithName("client-server"),
				testserver.WithMesh(meshName),
				testserver.WithNamespace(namespace),
				testserver.WithReachableServices("first-test-server_reachable-svc_svc_80"),
			)).
			Install(testserver.Install(
				testserver.WithName("first-test-server"),
				testserver.WithMesh(meshName),
				testserver.WithNamespace(namespace),
			)).
			Install(testserver.Install(
				testserver.WithName("second-test-server"),
				testserver.WithMesh(meshName),
				testserver.WithNamespace(namespace),
			)).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should be able to connect to reachable services", func() {
		Eventually(func(g Gomega) {
			_, err := client.CollectFailure(
				kubernetes.Cluster, "client-server", "first-test-server_reachable-svc_svc_80.mesh",
				client.FromKubernetesPod(namespace, "client-server"),
			)
			g.Expect(err).ToNot(HaveOccurred())
		}, "30s", "1s").Should(Succeed())
	})

	It("should not connect to non reachable service", func() {
		Consistently(func(g Gomega) {
			// when trying to connect to non-reachable services via Kuma DNS
			response, err := client.CollectFailure(
				kubernetes.Cluster, "client-server", "second-test-server_reachable-svc_svc_80.mesh",
				client.FromKubernetesPod(namespace, "client-server"),
			)
			// then it fails because Kuma DP has no such DNS
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Exitcode).To(Or(Equal(6), Equal(28)))
		}).Should(Succeed())

		Consistently(func(g Gomega) {
			// when trying to connect to non-reachable service via Kubernetes DNS
			response, err := client.CollectFailure(
				kubernetes.Cluster, "client-server", "second-test-server",
				client.FromKubernetesPod(namespace, "client-server"),
			)
			// then it fails because we don't encrypt traffic to unknown destination in the mesh
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Exitcode).To(Or(Equal(52), Equal(56)))
		}).Should(Succeed())
	})
}
