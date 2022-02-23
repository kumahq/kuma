package resilience

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/postgres"
)

func ResilienceStandaloneUniversal() {
	var universal Cluster

	BeforeEach(func() {
		clusters, err := NewUniversalClusters([]string{Kuma1}, Silent)
		Expect(err).ToNot(HaveOccurred())

		universal = clusters.GetCluster(Kuma1)

		err = NewClusterSetup().
			Install(postgres.Install(Kuma1)).
			Setup(universal)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(Kuma(core.Standalone,
				WithPostgres(postgres.From(universal, Kuma1).GetEnvVars()),
				WithEnv("KUMA_METRICS_DATAPLANE_IDLE_TIMEOUT", "10s"),
			)).
			Setup(universal)
		Expect(err).ToNot(HaveOccurred())

		demoClientToken, err := universal.GetKuma().GenerateDpToken("default", "demo-client")
		Expect(err).ToNot(HaveOccurred())

		Expect(
			DemoClientUniversal(AppModeDemoClient, "default", demoClientToken, WithKumactlFlow())(universal),
		).To(Succeed())
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
		_, _, err := universal.Exec("", "", AppModeCP, "pkill", "-9", "kuma-cp")
		Expect(err).ToNot(HaveOccurred())

		// and DPP is killed while Kuma CP is down
		_, _, err = universal.Exec("", "", AppModeDemoClient, "pkill", "-9", "envoy")
		Expect(err).ToNot(HaveOccurred())

		// and Kuma CP is restarted
		Expect(kumaCP.ReStart()).Should(Succeed())

		Eventually(universal.VerifyKuma, "30s", "1s").ShouldNot(HaveOccurred())

		// then DPP is offline
		Eventually(func() (string, error) {
			return universal.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "dataplanes")
		}, "40s", "1s").Should(ContainSubstring("Offline"))
	})
}
