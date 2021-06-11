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
	var deployOptsFuncs []DeployOptionsFunc

	BeforeEach(func() {
		cluster = NewUniversalCluster(NewTestingT(), Kuma3, Silent)
		deployOptsFuncs = KumaUniversalDeployOpts

		err := NewClusterSetup().
			Install(Kuma(core.Standalone, deployOptsFuncs...)).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())
		err = cluster.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		echoServerToken, err := cluster.GetKuma().GenerateDpToken("default", "echo-server_kuma-test_svc_8080")
		Expect(err).ToNot(HaveOccurred())
		demoClientToken, err := cluster.GetKuma().GenerateDpToken("default", "demo-client")
		Expect(err).ToNot(HaveOccurred())

		err = EchoServerUniversal(AppModeEchoServer, "default", "universal", echoServerToken, WithDPVersion("1.1.6"))(cluster)
		Expect(err).ToNot(HaveOccurred())
		err = DemoClientUniversal(AppModeDemoClient, "default", demoClientToken, WithDPVersion("1.1.6"))(cluster)
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
			_, _, err := cluster.Exec("", "", "demo-client", "curl", "-v", "-m", "3", "--fail", "localhost:4001")
			return "", err
		})
	})
}
