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
			Install(DemoClientUniversal("tp-client", mesh, WithTransparentProxy(true))).
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())
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
				"/usr/bin/kumactl", "install", "transparent-proxy",
				"--kuma-dp-user", "kuma-dp", "--verbose")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring("Transparent proxy set up successfully"))
		}).Should(Succeed())

		// then
		Eventually(func(g Gomega) {
			_, err := client.CollectResponses(universal.Cluster, "tp-client", "test-server.mesh")
			g.Expect(err).ToNot(HaveOccurred())
		}).Should(Succeed())
	})
}
