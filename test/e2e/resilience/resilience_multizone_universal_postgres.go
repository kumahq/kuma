package resilience

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/postgres"
)

func ResilienceMultizoneUniversalPostgres() {
	var global, remote_1 Cluster
	var optsGlobal, optsRemote1 []DeployOptionsFunc

	BeforeEach(func() {
		clusters, err := NewUniversalClusters(
			[]string{Kuma1, Kuma2},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		// Global
		global = clusters.GetCluster(Kuma1)
		optsGlobal = []DeployOptionsFunc{}

		err = NewClusterSetup().
			Install(postgres.Install(Kuma1)).
			Setup(global)
		Expect(err).ToNot(HaveOccurred())

		optsGlobal = []DeployOptionsFunc{
			WithPostgres(postgres.From(global, Kuma1).GetEnvVars()),
		}

		err = NewClusterSetup().
			Install(Kuma(core.Global, optsGlobal...)).
			Setup(global)
		Expect(err).ToNot(HaveOccurred())
		err = global.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		globalCP := global.GetKuma()

		// Cluster 1
		remote_1 = clusters.GetCluster(Kuma2)

		err = NewClusterSetup().
			Install(postgres.Install(Kuma2)).
			Setup(remote_1)
		Expect(err).ToNot(HaveOccurred())

		optsRemote1 = []DeployOptionsFunc{
			WithGlobalAddress(globalCP.GetKDSServerAddress()),
			WithPostgres(postgres.From(remote_1, Kuma2).GetEnvVars()),
		}

		err = NewClusterSetup().
			Install(Kuma(core.Zone, optsRemote1...)).
			Setup(remote_1)
		Expect(err).ToNot(HaveOccurred())

		err = remote_1.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterEach(func() {
		err := remote_1.DeleteKuma(optsRemote1...)
		Expect(err).ToNot(HaveOccurred())
		err = remote_1.DismissCluster()
		Expect(err).ToNot(HaveOccurred())

		err = global.DeleteKuma(optsGlobal...)
		Expect(err).ToNot(HaveOccurred())
		err = global.DismissCluster()
		Expect(err).ToNot(HaveOccurred())
	})

	PIt("should mark zone as offline after zone control-plane is killed forcefully when global control-plane is down", func() {
		// given zone connected to global
		Eventually(func() (string, error) {
			return global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zones")
		}, "30s", "1s").Should(ContainSubstring("Online"))

		g, ok := global.(*UniversalCluster)
		Expect(ok).To(BeTrue())

		kumaCP := g.GetApp(AppModeCP)
		Expect(kumaCP).ToNot(BeNil())

		// when global is killed
		_, _, err := global.Exec("", "", AppModeCP, "pkill", "-9", "kuma-cp")
		Expect(err).ToNot(HaveOccurred())

		// and zone is killed while global is down
		_, _, err = remote_1.Exec("", "", AppModeCP, "pkill", "-9", "kuma-cp")
		Expect(err).ToNot(HaveOccurred())

		// and global is restarted
		Expect(kumaCP.ReStart()).Should(Succeed())

		Eventually(global.VerifyKuma, "30s", "1s").ShouldNot(HaveOccurred())

		// then zone is offline
		Eventually(func() (string, error) {
			return global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zones")
		}, "30s", "1s").Should(ContainSubstring("Offline"))
	})

	It("should mark zone as offline after global control-plane is ungracefully restarted", func() {
		// given zone connected to global
		Eventually(func() (string, error) {
			return global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zones")
		}, "30s", "1s").Should(ContainSubstring("Online"))

		g, ok := global.(*UniversalCluster)
		Expect(ok).To(BeTrue())

		kumaCP := g.GetApp(AppModeCP)
		Expect(kumaCP).ToNot(BeNil())

		// when global is killed
		_, _, err := global.Exec("", "", AppModeCP, "pkill", "-9", "kuma-cp")
		Expect(err).ToNot(HaveOccurred())

		// and global is restarted
		Expect(kumaCP.ReStart()).Should(Succeed())

		Eventually(global.VerifyKuma, "30s", "1s").ShouldNot(HaveOccurred())

		time.Sleep(10 * time.Second) // ZoneInsightFlushInterval

		// and zone is killed
		_, _, err = remote_1.Exec("", "", AppModeCP, "pkill", "-9", "kuma-cp")
		Expect(err).ToNot(HaveOccurred())

		// then zone is offline
		Eventually(func() (string, error) {
			return global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zones")
		}, "30s", "1s").Should(ContainSubstring("Offline"))
	})
}
