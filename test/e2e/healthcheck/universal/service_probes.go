package universal

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

func ServiceProbes() {
	var cluster Cluster

	BeforeEach(func() {
		cluster = NewUniversalCluster(NewTestingT(), Kuma3, Silent)

		err := NewClusterSetup().
			Install(Kuma(core.Standalone)).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())

		echoServerToken, err := cluster.GetKuma().GenerateDpToken("default", "test-server")
		Expect(err).ToNot(HaveOccurred())
		demoClientToken, err := cluster.GetKuma().GenerateDpToken("default", "dp-demo-client")
		Expect(err).ToNot(HaveOccurred())

		err = TestServerUniversal("test-server", "default", echoServerToken,
			WithArgs([]string{"echo", "--instance", "universal-1"}),
			ProxyOnly(),
			ServiceProbe())(cluster)
		Expect(err).ToNot(HaveOccurred())
		err = DemoClientUniversal("dp-demo-client", "default", demoClientToken, ServiceProbe())(cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}
		Expect(cluster.DeleteKuma()).To(Succeed())
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	It("should update dataplane.inbound.health", func() {
		Eventually(func() (string, error) {
			output, err := cluster.GetKumactlOptions().RunKumactlAndGetOutputV(Verbose, "get", "dataplane", "test-server", "-oyaml")
			if err != nil {
				return "", err
			}
			return output, nil
		}, "30s", "500ms").Should(ContainSubstring("health: {}"))

		Eventually(func() (string, error) {
			output, err := cluster.GetKumactlOptions().RunKumactlAndGetOutputV(Verbose, "get", "dataplane", "dp-demo-client", "-oyaml")
			if err != nil {
				return "", err
			}
			return output, nil
		}, "30s", "500ms").Should(ContainSubstring("ready: true"))
	})
}
