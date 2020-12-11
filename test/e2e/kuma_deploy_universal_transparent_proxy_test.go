package e2e_test

import (
	"github.com/gruntwork-io/terratest/modules/retry"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

var _ = Describe("Test Universal Transparent Proxy deployment", func() {

	const iterations = 100

	var cluster Cluster
	var deployOptsFuncs []DeployOptionsFunc

	BeforeEach(func() {
		cluster = NewUniversalCluster(NewTestingT(), Kuma1, Verbose)
		deployOptsFuncs = []DeployOptionsFunc{}

		err := NewClusterSetup().
			Install(Kuma(core.Standalone, deployOptsFuncs...)).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())
		err = cluster.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		echoServerToken, err := cluster.GetKuma().GenerateDpToken("echo-server_kuma-test_svc_80")
		Expect(err).ToNot(HaveOccurred())
		demoClientToken, err := cluster.GetKuma().GenerateDpToken("demo-client")
		Expect(err).ToNot(HaveOccurred())

		err = EchoServerUniversal("universal", echoServerToken, WithTransparentProxy(true))(cluster)
		Expect(err).ToNot(HaveOccurred())
		err = DemoClientUniversal(demoClientToken, WithTransparentProxy(true))(cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}
		Expect(cluster.DeleteKuma(deployOptsFuncs...)).To(Succeed())
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	It("should access the service using .mesh", func() {

		for i := 0; i < iterations; i++ {
			retry.DoWithRetry(cluster.GetTesting(), "check service access", DefaultRetries, DefaultTimeout, func() (string, error) {
				// when client sends requests to server
				_, _, err := cluster.Exec("", "", "demo-client",
					"curl", "-v", "-m", "3", "--fail",
					"echo-server_kuma-test_svc_80.mesh",
				)
				if err != nil {
					return "", err
				}

				return "", nil
			})
		}
	})
})
