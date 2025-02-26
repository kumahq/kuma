package resilience

import (
	"fmt"
	"strings"

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

	AfterEachFailure(func() {
		DebugUniversal(zone1, "default")
		DebugUniversal(zone2, "default")
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
		var leader, follower Cluster

		// ensure only one control plane instance is elected as leader
		Eventually(func(g Gomega) {
			cp1Metrics, err := zone1.GetKuma().GetMetrics()
			g.Expect(err).ToNot(HaveOccurred())

			cp2Metrics, err := zone2.GetKuma().GetMetrics()
			g.Expect(err).ToNot(HaveOccurred())

			// Identify the current leader dynamically
			if strings.Contains(cp1Metrics, `leader{zone="kuma-leader1"} 1`) {
				g.Expect(cp1Metrics).To(ContainSubstring(`leader{zone="kuma-leader1"} 1`))
				g.Expect(cp2Metrics).To(ContainSubstring(`leader{zone="kuma-leader2"} 0`))
				leader, follower = zone1, zone2
			} else {
				g.Expect(cp2Metrics).To(ContainSubstring(`leader{zone="kuma-leader2"} 1`))
				g.Expect(cp1Metrics).To(ContainSubstring(`leader{zone="kuma-leader1"} 0`))
				leader, follower = zone2, zone1
			}
		}, "10s", "100ms").Should(Succeed())

		// Kill the current leader
		_, _, err := leader.Exec("", "", AppModeCP, "pkill", "kuma-cp")
		Expect(err).ToNot(HaveOccurred())

		// Verify that the other instance takes over as leader
		Eventually(func() (string, error) {
			return follower.GetKuma().GetMetrics()
		}, "30s", "1s").Should(ContainSubstring(fmt.Sprintf(`leader{zone="%s"} 1`, follower.Name())))

		// Shut down PostgreSQL
		err = zone1.DeleteDeployment(postgres.AppPostgres + clusterName1)
		Expect(err).ToNot(HaveOccurred())

		// Verify that the remaining control plane instance loses leadership
		Eventually(func() (string, error) {
			return follower.GetKuma().GetMetrics()
		}, "30s", "1s").Should(ContainSubstring(fmt.Sprintf(`leader{zone="%s"} 0`, follower.Name())))
	})
}
