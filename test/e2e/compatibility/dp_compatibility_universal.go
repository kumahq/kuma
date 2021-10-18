package compatibility

import (
	"github.com/gruntwork-io/terratest/modules/retry"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

func UniversalCompatibility() {
	var cluster Cluster
	var deployOptsFuncs []KumaDeploymentOption

	BeforeEach(func() {
		cluster = NewUniversalCluster(NewTestingT(), Kuma3, Silent)
		deployOptsFuncs = KumaUniversalDeployOpts

		err := NewClusterSetup().
			Install(Kuma(core.Standalone, deployOptsFuncs...)).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())
		err = cluster.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		testServerToken, err := cluster.GetKuma().GenerateDpToken("default", "test-server")
		Expect(err).ToNot(HaveOccurred())
		demoClientToken, err := cluster.GetKuma().GenerateDpToken("default", "demo-client")
		Expect(err).ToNot(HaveOccurred())

		err = TestServerUniversal("test-server", "default", testServerToken,
			WithArgs([]string{"echo", "--instance", "universal1"}),
			WithDPVersion("1.1.6"))(cluster)
		Expect(err).ToNot(HaveOccurred())
		err = DemoClientUniversal(AppModeDemoClient, "default", demoClientToken, WithDPVersion("1.1.6"), WithTransparentProxy(true))(cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}
		Expect(cluster.DeleteKuma(deployOptsFuncs...)).To(Succeed())
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	It("client should access server", func() {
		retry.DoWithRetry(cluster.GetTesting(), "check communication between services", DefaultRetries, DefaultTimeout, func() (string, error) {
			_, _, err := cluster.Exec("", "", "demo-client", "curl", "-v", "-m", "3", "--fail", "test-server.mesh")
			return "", err
		})
	})
}
