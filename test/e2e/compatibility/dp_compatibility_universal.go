package compatibility

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	. "github.com/kumahq/kuma/test/framework/client"
)

func UniversalCompatibility() {
	var cluster Cluster

	BeforeEach(func() {
		cluster = NewUniversalCluster(NewTestingT(), Kuma3, Silent)

		err := NewClusterSetup().
			Install(Kuma(core.Standalone)).
			Install(TestServerUniversal("test-server", "default",
				WithArgs([]string{"echo", "--instance", "universal1"}),
				WithDPVersion("1.5.0"))).
			Install(DemoClientUniversal(AppModeDemoClient, "default",
				WithDPVersion("1.5.0"),
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
			_, err := CollectResponse(cluster, "demo-client", "test-server.mesh")
			g.Expect(err).ToNot(HaveOccurred())
		}, "20s", "250ms").Should(Succeed())
	})
}
