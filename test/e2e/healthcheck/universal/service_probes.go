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
			Install(TestServerUniversal("test-server", "default",
				WithArgs([]string{"echo", "--instance", "universal-1"}),
				ProxyOnly(),
				ServiceProbe()),
			).
			Install(DemoClientUniversal("dp-demo-client", "default", ServiceProbe())).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterEach(func() {
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
