package reachableservices

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func ReachableServices() {
	meshName := "reachable-svc"
	namespace := "reachable-svc"

	var clientPodName string

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MTLSMeshKubernetes(meshName)).
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

		clientPodName, err = PodNameOfApp(kubernetes.Cluster, "client-server", namespace)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(meshName))
	})

	It("should be able to connect to reachable services", func() {
		Eventually(func(g Gomega) {
			// when
			_, stderr, err := kubernetes.Cluster.Exec(namespace, clientPodName, "client-server",
				"curl", "-v", "-m", "3", "--fail", "first-test-server_reachable-svc_svc_80.mesh")

			// then
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))
		}).Should(Succeed())
	})

	It("should not connect to non reachable service", func() {
		Consistently(func(g Gomega) {
			// when trying to connect to non-reachable services via Kuma DNS
			_, _, err := kubernetes.Cluster.Exec(namespace, clientPodName, "client-server",
				"curl", "-v", "second-test-server_reachable-svc_svc_80.mesh")

			// then it fails because Kuma DP has no such DNS
			g.Expect(err).To(HaveOccurred())
		}).Should(Succeed())

		Consistently(func(g Gomega) {
			// when trying to connect to non-reachable service via Kubernetes DNS
			_, _, err := kubernetes.Cluster.Exec(namespace, clientPodName, "client-server",
				"curl", "-v", "second-test-server")

			// then it fails because we don't encrypt traffic to unknown destination in the mesh
			g.Expect(err).To(HaveOccurred())
		}).Should(Succeed())
	})
}
