package compatibility

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/versions"
)

func UniversalCompatibility() {
	var cluster Cluster

	BeforeEach(func() {
		cluster = NewUniversalCluster(NewTestingT(), "kuma-compat", Silent)

		oldestUpgradble := versions.OldestUpgradableToBuildVersion(Config.SupportedVersions())
		err := NewClusterSetup().
			Install(Kuma(core.Zone)).
			Install(TestServerUniversal("test-server", "default",
				WithArgs([]string{"echo", "--instance", "universal1"}),
				WithDPVersion(oldestUpgradble))).
			Install(DemoClientUniversal(AppModeDemoClient, "default",
				WithDPVersion(oldestUpgradble),
				WithTransparentProxy(true)),
			).
			SetupWithRetries(cluster, 3)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterEach(func() {
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	It("client should access server", func() {
		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(cluster, "demo-client", "test-server.mesh")
			g.Expect(err).ToNot(HaveOccurred())
		}, "20s", "250ms").Should(Succeed())
	})
}
