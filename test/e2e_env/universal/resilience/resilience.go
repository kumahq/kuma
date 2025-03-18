package resilience

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/postgres"
)

func ResilienceUniversal() {
	const clusterName = "kuma-resilience"
	var universal Cluster

	BeforeEach(func() {
		universal = NewUniversalCluster(NewTestingT(), clusterName, Silent)

		Expect(postgres.Install(clusterName)(universal)).To(Succeed())

		err := NewClusterSetup().
			Install(Kuma(core.Zone,
				WithPostgres(postgres.From(universal, clusterName).GetEnvVars()),
				WithEnv("KUMA_METRICS_DATAPLANE_IDLE_TIMEOUT", "10s"),
			)).
			Install(DemoClientUniversal(AppModeDemoClient, "default", WithKumactlFlow())).
			Setup(universal)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal, "default")
	})

	E2EAfterEach(func() {
		Expect(universal.DeleteKuma()).To(Succeed())
		Expect(universal.DismissCluster()).To(Succeed())
	})

	It("should mark data plane proxy as offline after it is killed forcefully when control-plane is down", func() {
		// given DPP connected to Kuma CP
		Eventually(func() (string, error) {
			return universal.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "dataplanes")
		}, "30s", "1s").Should(ContainSubstring("Online"))

		g, ok := universal.(*UniversalCluster)
		Expect(ok).To(BeTrue())

		kumaCP := g.GetApp(AppModeCP)
		Expect(kumaCP).ToNot(BeNil())

		// when Kuma CP is killed
		Expect(g.Kill(AppModeCP, "kuma-cp run")).To(Succeed())

		out, _, err := g.Exec("", "", AppModeDemoClient, "ps", "aux")
		Logf("ps aux: %q", out)
		Expect(err).ToNot(HaveOccurred())
		// and DPP is killed while Kuma CP is down
		Expect(g.Kill(AppModeDemoClient, "kuma-dp")).To(Succeed())

		// and Kuma CP is restarted
		Expect(kumaCP.ReStart()).Should(Succeed())

		Eventually(universal.VerifyKuma, "30s", "1s").ShouldNot(HaveOccurred())

		// then DPP is offline
		Eventually(func() (string, error) {
			return universal.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "dataplanes")
		}, "40s", "1s").Should(ContainSubstring("Offline"))
	})
}
