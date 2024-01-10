package resilience

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/postgres"
)

func LeaderElectionPostgres() {
	const clusterName1 = "kuma-leader1"
	const clusterName2 = "kuma-leader2"
	var zone1, zone2 Cluster

	BeforeEach(func() {
		zone1 = NewUniversalCluster(NewTestingT(), clusterName1, Silent)
		zone2 = NewUniversalCluster(NewTestingT(), clusterName2, Silent)

		err := NewClusterSetup().
			Install(postgres.Install(clusterName1)).
			Setup(zone1)
		Expect(err).ToNot(HaveOccurred())
		postgresInstance := postgres.From(zone1, clusterName1)

		err = NewClusterSetup().
			Install(Kuma(core.Zone, WithPostgres(postgresInstance.GetEnvVars()))).
			Setup(zone1)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(Kuma(core.Zone, WithPostgres(postgresInstance.GetEnvVars()))).
			Setup(zone2)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterEach(func() {
		err := zone1.DeleteKuma()
		Expect(err).ToNot(HaveOccurred())
		err = zone1.DismissCluster()
		Expect(err).ToNot(HaveOccurred())

		err = zone2.DeleteKuma()
		Expect(err).ToNot(HaveOccurred())
		err = zone2.DismissCluster()
		Expect(err).ToNot(HaveOccurred())
	})

	It("should elect only one leader and drop the leader on DB disconnect", func() {
		// given two instances of the control plane connected to one postgres, only one is a leader
		Eventually(func() (string, error) {
			return zone1.GetKuma().GetMetrics()
		}, "30s", "1s").Should(ContainSubstring(`leader{zone="kuma-leader1"} 1`))

		metrics, err := zone2.GetKuma().GetMetrics()
		Expect(err).ToNot(HaveOccurred())
		Expect(metrics).To(ContainSubstring(`leader{zone="kuma-leader2"} 0`))

		// when CP 1 is killed
		_, _, err = zone1.Exec("", "", AppModeCP, "pkill", "kuma-cp")
		Expect(err).ToNot(HaveOccurred())

		// then CP 2 is leader
		Eventually(func() (string, error) {
			return zone2.GetKuma().GetMetrics()
		}, "30s", "1s").Should(ContainSubstring(`leader{zone="kuma-leader2"} 1`))

		// when postgres is down
		err = zone1.DeleteDeployment(postgres.AppPostgres + clusterName1)
		Expect(err).ToNot(HaveOccurred())

		// then CP 2 is not a leader anymore
		Eventually(func() (string, error) {
			return zone2.GetKuma().GetMetrics()
		}, "30s", "1s").Should(ContainSubstring(`leader{zone="kuma-leader2"} 0`))
	})
}
