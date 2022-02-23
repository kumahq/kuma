package resilience

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/postgres"
)

func LeaderElectionPostgres() {
	var standalone1, standalone2 Cluster

	BeforeEach(func() {
		clusters, err := NewUniversalClusters(
			[]string{Kuma1, Kuma2},
			Silent)
		Expect(err).ToNot(HaveOccurred())
		standalone1 = clusters.GetCluster(Kuma1)
		standalone2 = clusters.GetCluster(Kuma2)

		err = NewClusterSetup().
			Install(postgres.Install(Kuma1)).
			Setup(standalone1)
		Expect(err).ToNot(HaveOccurred())
		postgresInstance := postgres.From(standalone1, Kuma1)

		// Standalone 1
		err = NewClusterSetup().
			Install(Kuma(core.Standalone, WithPostgres(postgresInstance.GetEnvVars()))).
			Setup(standalone1)

		Expect(err).ToNot(HaveOccurred())

		// Standalone 2
		err = NewClusterSetup().
			Install(Kuma(core.Standalone, WithPostgres(postgresInstance.GetEnvVars()))).
			Setup(standalone2)

		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterEach(func() {
		err := standalone1.DeleteKuma()
		Expect(err).ToNot(HaveOccurred())
		err = standalone1.DismissCluster()
		Expect(err).ToNot(HaveOccurred())

		err = standalone2.DeleteKuma()
		Expect(err).ToNot(HaveOccurred())
		err = standalone2.DismissCluster()
		Expect(err).ToNot(HaveOccurred())
	})

	It("should elect only one leader and drop the leader on DB disconnect", func() {
		// given two instances of the control plane connected to one postgres, only one is a leader
		Eventually(func() (string, error) {
			return standalone1.GetKuma().GetMetrics()
		}, "30s", "1s").Should(ContainSubstring(`leader{zone="Standalone"} 1`))

		metrics, err := standalone2.GetKuma().GetMetrics()
		Expect(err).ToNot(HaveOccurred())
		Expect(metrics).To(ContainSubstring(`leader{zone="Standalone"} 0`))

		// when CP 1 is killed
		_, _, err = standalone1.Exec("", "", AppModeCP, "pkill", "-9", "kuma-cp")
		Expect(err).ToNot(HaveOccurred())

		// then CP 2 is leader
		Eventually(func() (string, error) {
			return standalone2.GetKuma().GetMetrics()
		}, "30s", "1s").Should(ContainSubstring(`leader{zone="Standalone"} 1`))

		// when postgres is down
		err = standalone1.DeleteDeployment(postgres.AppPostgres + Kuma1)
		Expect(err).ToNot(HaveOccurred())

		// then CP 2 is not a leader anymore
		Eventually(func() (string, error) {
			return standalone2.GetKuma().GetMetrics()
		}, "30s", "1s").Should(ContainSubstring(`leader{zone="Standalone"} 0`))
	})
}
