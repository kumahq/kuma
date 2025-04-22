package transparentproxy

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func TransparentProxy() {
	const mesh = "transparent-proxy"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MTLSMeshUniversal(mesh)).
			Install(TestServerUniversal("test-server", mesh,
				WithTransparentProxy(true),
				WithArgs([]string{"echo", "--instance", "echo-v1"}),
				WithServiceName("test-server"),
			)).
			Install(MeshTrafficPermissionAllowAllUniversal(mesh)).
			Install(DemoClientUniversal("tp-client", mesh, WithTransparentProxy(true))).
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, mesh)
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(mesh)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(mesh)).To(Succeed())
	})

	It("should be able to re-install transparent proxy", func() {
		// given
		Eventually(func(g Gomega) {
			_, err := client.CollectResponses(universal.Cluster, "tp-client", "test-server.mesh")
			g.Expect(err).ToNot(HaveOccurred())
		}).Should(Succeed())

		// when
		Eventually(func(g Gomega) {
			stdout, _, err := universal.Cluster.Exec("", "", "tp-client",
				"/usr/bin/kumactl", "uninstall", "transparent-proxy", "--verbose")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring("transparent proxy cleanup completed successfully"))
		}).Should(Succeed())

		// then
		Eventually(func(g Gomega) {
			failures, err := client.CollectFailure(universal.Cluster, "tp-client", "test-server.mesh")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(failures.Exitcode).To((Or(Equal(6), Equal(28))))
		}).Should(Succeed())

		// and
		Eventually(func(g Gomega) {
			stdout, _, err := universal.Cluster.Exec("", "", "tp-client",
				"/usr/bin/kumactl", "install", "transparent-proxy",
				"--redirect-dns", "--verbose")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring("transparent proxy setup completed successfully"))
		}).Should(Succeed())

		// then
		Eventually(func(g Gomega) {
			_, err := client.CollectResponses(universal.Cluster, "tp-client", "test-server.mesh")
			g.Expect(err).ToNot(HaveOccurred())
		}).Should(Succeed())
	})
}
