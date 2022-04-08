package compatibility

import (
	"github.com/gruntwork-io/terratest/modules/retry"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

func UniversalCompatibility() {
	var cluster Cluster

	BeforeEach(func() {
		cluster = NewUniversalCluster(NewTestingT(), Kuma3, Silent)

		err := NewClusterSetup().
			Install(Kuma(core.Standalone)).
			Install(TestServerUniversal("test-server", "default",
				WithArgs([]string{"echo", "--instance", "universal1"}),
				WithDPVersion("1.1.6"))).
			Install(DemoClientUniversal(AppModeDemoClient, "default",
				WithDPVersion("1.1.6"),
				WithTransparentProxy(true)),
			).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterEach(func() {
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	It("client should access server", func() {
		retry.DoWithRetry(cluster.GetTesting(), "check communication between services", DefaultRetries, DefaultTimeout, func() (string, error) {
			_, _, err := cluster.Exec("", "", "demo-client", "curl", "-v", "-m", "3", "--fail", "test-server.mesh")
			return "", err
		})
	})
}
