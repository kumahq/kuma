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
	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(mesh)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(mesh)).To(Succeed())
	})

	// This test is currently disabled as it basically never worked as expected
	// as when installing tproxy again, it was just adding the same set of rules
	// after the existing ones (with small exception to DNS ones, but it's
	// irrelevant here). So the tests were passing only because rules form
	// the second installation have been just ignored.
	//
	// When mechanism to uninstall tproxy will be implemented (ref.
	// https://github.com/kumahq/kuma/issues/6093), we will adapt this test
	// to run the uninstaller step in between.
	XIt("should be able to re-install transparent proxy", func() {
		// given
		Eventually(func(g Gomega) {
			_, err := client.CollectResponses(universal.Cluster, "tp-client", "test-server.mesh")
			g.Expect(err).ToNot(HaveOccurred())
		}).Should(Succeed())

		// when

		// This logic is currently non-existent
		// TODO: remove above comment after implementing uninstaller logic
		//       (ref. https://github.com/kumahq/kuma/issues/6093)
		Eventually(func(g Gomega) {
			stdout, _, err := universal.Cluster.Exec("", "", "tp-client",
				"/usr/bin/kumactl", "uninstall", "transparent-proxy",
				"--kuma-dp-user", "kuma-dp", "--verbose")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring("Transparent proxy uninstalled successfully"))
		}).Should(Succeed())

		// and
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
