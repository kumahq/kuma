package jobs

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func Jobs() {
	It("should terminate jobs without mTLS", func() {
		const namespace = "jobs"
		const mesh = "jobs"

		E2EDeferCleanup(func() {
			Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
			Expect(kubernetes.Cluster.DeleteMesh(mesh)).To(Succeed())
		})

		err := NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(MeshKubernetes(mesh)).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh(mesh),
			)).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// when
		err = kubernetes.Cluster.Install(DemoClientJobK8s(namespace, mesh, "test-server_jobs_svc_80.mesh"))

		// then CP terminates the job by sending /quitquitquit to Envoy Admin and verifies connection using self-signed certs
		Expect(err).ToNot(HaveOccurred())

		// and Dataplane object is deleted
		Eventually(func(g Gomega) {
			out, err := kubernetes.Cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplanes", "--mesh", mesh)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(out).ToNot(ContainSubstring("demo-job-client"))
		}, "30s", "1s").Should(Succeed())
	})

	It("should terminate jobs with mTLS", func() {
		const namespace = "jobs-mtls"
		const mesh = "jobs-mtls"

		E2EDeferCleanup(func() {
			Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
			Expect(kubernetes.Cluster.DeleteMesh(mesh)).To(Succeed())
		})

		err := NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(MTLSMeshKubernetes(mesh)).
			Install(MeshTrafficPermissionAllowAllKubernetes(mesh)).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh(mesh),
			)).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// when
		err = kubernetes.Cluster.Install(DemoClientJobK8s(namespace, mesh, "test-server_jobs-mtls_svc_80.mesh"))

		// then CP terminates the job by sending /quitquitquit to Envoy Admin and verifies connection using mTLS certs
		Expect(err).ToNot(HaveOccurred())
	})
}
