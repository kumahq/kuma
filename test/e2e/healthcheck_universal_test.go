package e2e_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

var _ = Describe("Test application HealthCheck on Universal", func() {

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

		echoServerToken, err := cluster.GetKuma().GenerateDpToken("default", "echo-server_kuma-test_svc_8080")
		Expect(err).ToNot(HaveOccurred())
		demoClientToken, err := cluster.GetKuma().GenerateDpToken("default", "demo-client")
		Expect(err).ToNot(HaveOccurred())

		err = EchoServerUniversal("universal", "default", echoServerToken, ProxyOnly())(cluster)
		Expect(err).ToNot(HaveOccurred())
		err = DemoClientUniversal("default", demoClientToken)(cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}
		Expect(cluster.DeleteKuma(deployOptsFuncs...)).To(Succeed())
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	It("should update dataplane.inbound.health", func() {
		Eventually(func() (string, error) {
			output, err := cluster.GetKumactlOptions().RunKumactlAndGetOutputV(Verbose, "get", "dataplane", "dp-echo-server", "-oyaml")
			if err != nil {
				return "", err
			}
			return output, nil
		}, "10s", "1s").Should(ContainSubstring("health: {}"))

		Eventually(func() (string, error) {
			output, err := cluster.GetKumactlOptions().RunKumactlAndGetOutputV(Verbose, "get", "dataplane", "dp-demo-client", "-oyaml")
			if err != nil {
				return "", err
			}
			return output, nil
		}, "10s", "1s").Should(ContainSubstring("ready: true"))
	})

})
