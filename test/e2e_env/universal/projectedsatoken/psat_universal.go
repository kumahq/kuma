package projectedsatoken

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

func ProjectedServiceAccountToken() {
	var universal Cluster

	BeforeEach(func() {
		universal = NewUniversalCluster(NewTestingT(), "kuma-psat", Silent)
		Expect(NewClusterSetup().
			Install(Kuma(core.Zone,
				WithEnv("KUMA_DP_SERVER_AUTHN_ENABLE_RELOADABLE_TOKENS", "true"),
			)).
			Install(DemoClientUniversal("demo-client", "default")).
			Setup(universal)).To(Succeed())
	})

	E2EAfterEach(func() {
		Expect(universal.DismissCluster()).To(Succeed())
	})

	It("should connect to restarted control plane with new token without dp restart", func() {
		// given
		uniCluster := universal.(*UniversalCluster)
		kumaCP := universal.(*UniversalCluster).GetApp(AppModeCP)

		// then should have dataplane in control plane
		Eventually(func() bool {
			stdout, _, err := uniCluster.Exec("", "", "kuma-cp", "kumactl", "get", "dataplanes")
			if err != nil {
				return false
			}
			return strings.Contains(stdout, "demo-client")
		}, "60s", "1s").Should(BeTrue())

		// when restart control-plane
		Expect(kumaCP.ReStart()).Should(Succeed())

		// then should not have demo-client in dataplanes
		Eventually(func() bool {
			stdout, _, err := uniCluster.Exec("", "", "kuma-cp", "kumactl", "get", "dataplanes")
			if err != nil {
				return true
			}
			return strings.Contains(stdout, "demo-client")
		}, "60s", "1s").Should(BeFalse())

		// when new token generated
		token, err := universal.GetKuma().GenerateDpToken("default", "demo-client")
		Expect(err).ToNot(HaveOccurred())

		// and set token
		_, _, err = universal.Exec("", "", "demo-client", "printf", "\""+token+"\"", ">", "/kuma/token-demo-client")
		Expect(err).ToNot(HaveOccurred())

		// then there should be dataplane for demo-client
		Eventually(func() bool {
			stdout, _, err := uniCluster.Exec("", "", "kuma-cp", "kumactl", "get", "dataplanes")
			if err != nil {
				return false
			}
			return strings.Contains(stdout, "demo-client")
		}, "60s", "1s").Should(BeTrue())
	})
}
