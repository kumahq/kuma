package dp

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

var cluster Cluster

var _ = E2EBeforeSuite(func() {
	cluster = NewUniversalCluster(NewTestingT(), Kuma3, Silent)

	Expect(NewClusterSetup().
		Install(Kuma(core.Standalone)).
		Setup(cluster)).To(Succeed())

	E2EDeferCleanup(cluster.DismissCluster)
})

func DpAuthUniversal() {
	It("should not be able to override someone else Dataplane", func() {
		// given other dataplane
		dp := `
type: Dataplane
mesh: default
name: dp-01
networking:
  address: 192.168.0.1
  inbound:
  - port: 8080
    tags:
      kuma.io/service: not-test-server
`
		Expect(YamlUniversal(dp)(cluster)).To(Succeed())

		// when trying to spin up dataplane with same name but token bound to a different service
		dpToken, err := cluster.GetKuma().GenerateDpToken("default", "test-server")
		Expect(err).ToNot(HaveOccurred())
		err = TestServerUniversal("dp-01", "default", dpToken, WithServiceName("test-server"))(cluster)
		Expect(err).ToNot(HaveOccurred())

		// then
		Eventually(func() (string, error) {
			return cluster.GetKuma().GetKumaCPLogs()
		}, "30s", "1s").Should(ContainSubstring("you are trying to override existing dataplane to which you don't have an access"))
	})

	It("should be able to override old Dataplane of same service", func() {
		// given
		dp := `
type: Dataplane
mesh: default
name: dp-02
networking:
  address: 192.168.0.2
  inbound:
  - port: 8080
    tags:
      kuma.io/service: test-server
`
		Expect(YamlUniversal(dp)(cluster)).To(Succeed())

		// when
		dpToken, err := cluster.GetKuma().GenerateDpToken("default", "test-server")
		Expect(err).ToNot(HaveOccurred())
		err = TestServerUniversal("dp-02", "default", dpToken, WithServiceName("test-server"))(cluster)
		Expect(err).ToNot(HaveOccurred())

		// then
		Eventually(func() (string, error) {
			return cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplanes", "-oyaml")
		}, "30s", "1s").ShouldNot(ContainSubstring("192.168.0.2"))
	})
}
