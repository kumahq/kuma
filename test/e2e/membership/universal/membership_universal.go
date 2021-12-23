package universal

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

func MembershipUniversal() {
	var cluster Cluster
	var deployOptsFuncs []KumaDeploymentOption

	E2EBeforeSuite(func() {
		cluster = NewUniversalCluster(NewTestingT(), Kuma3, Silent)
		deployOptsFuncs = KumaUniversalDeployOpts

		err := NewClusterSetup().
			Install(Kuma(core.Standalone, deployOptsFuncs...)).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())
		err = cluster.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterSuite(func() {
		Expect(cluster.DeleteKuma(deployOptsFuncs...)).To(Succeed())
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	It("should take into account membership when dp is connecting to the CP", func() {
		mesh := `
type: Mesh
name: default
dataplaneProxyMembership:
  allowed:
  - tags:
      kuma.io/service: demo-client
  denied:
  - tags:
      kuma.io/service: test-server
`
		Expect(YamlUniversal(mesh)(cluster)).To(Succeed())

		// when demo client is trying to connect
		demoClientToken, err := cluster.GetKuma().GenerateDpToken("default", "demo-client")
		Expect(err).ToNot(HaveOccurred())
		err = DemoClientUniversal(AppModeDemoClient, "default", demoClientToken)(cluster)
		Expect(err).ToNot(HaveOccurred())

		// then it's allowed
		Eventually(func() (string, error) {
			return cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplanes")
		}, "30s", "1s").ShouldNot(ContainSubstring("demo-client"))

		// when test server is trying to connect
		testServerToken, err := cluster.GetKuma().GenerateDpToken("default", "test-server")
		Expect(err).ToNot(HaveOccurred())
		err = TestServerUniversal("test-server", "default", testServerToken, WithArgs([]string{"echo", "--instance", "echo-v1"}))(cluster)
		Expect(err).ToNot(HaveOccurred())

		// then it's not allowed
		Eventually(func() (string, error) {
			return cluster.GetKuma().GetKumaCPLogs()
		}, "30s", "1s").Should(ContainSubstring("dataplane cannot be a member of mesh"))
		dataplanes, err := cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplanes")
		Expect(err).ToNot(HaveOccurred())
		Expect(dataplanes).ToNot(ContainSubstring("test-server"))
	})
}
