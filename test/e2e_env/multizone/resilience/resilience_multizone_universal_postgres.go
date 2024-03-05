package resilience

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/postgres"
)

func ResilienceMultizoneUniversalPostgres() {
	var clusterName1 string
	var clusterName2 string

	var global, zoneUniversal Cluster

	BeforeEach(func() {
		testingT := NewTestingT()
		clusterName1 = "kuma1-" + testingT.Hash()
		clusterName2 = "kuma2-" + testingT.Hash()
		// Global
		global = NewUniversalCluster(testingT, clusterName1, Verbose)

		err := NewClusterSetup().
			Install(postgres.Install(clusterName1)).
			Setup(global)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(Kuma(core.Global,
				WithPostgres(postgres.From(global, clusterName1).GetEnvVars()),
				WithEnv("KUMA_METRICS_ZONE_IDLE_TIMEOUT", "10s"),
			)).
			Setup(global)
		Expect(err).ToNot(HaveOccurred())

		globalCP := global.GetKuma()

		// Cluster 1
		zoneUniversal = NewUniversalCluster(NewTestingT(), clusterName2, Verbose)

		err = NewClusterSetup().
			Install(postgres.Install(clusterName2)).
			Setup(zoneUniversal)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(Kuma(core.Zone,
				WithGlobalAddress(globalCP.GetKDSServerAddress()),
				WithPostgres(postgres.From(zoneUniversal, clusterName2).GetEnvVars()),
				WithEnv("KUMA_METRICS_DATAPLANE_IDLE_TIMEOUT", "10s"),
			)).
			Setup(zoneUniversal)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterEach(func() {
		err := zoneUniversal.DeleteKuma()
		Expect(err).ToNot(HaveOccurred())
		err = zoneUniversal.DismissCluster()
		Expect(err).ToNot(HaveOccurred())

		err = global.DeleteKuma()
		Expect(err).ToNot(HaveOccurred())
		err = global.DismissCluster()
		Expect(err).ToNot(HaveOccurred())
	})

	It("should mark zone as offline after zone control-plane is killed forcefully when global control-plane is down", func() {
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
		_, _, err = zoneUniversal.Exec("", "", AppModeCP, "pkill", "-9", "kuma-cp")
		Expect(err).ToNot(HaveOccurred())

		// and global is restarted
		Expect(kumaCP.ReStart()).Should(Succeed())

		Eventually(global.VerifyKuma, "30s", "1s").ShouldNot(HaveOccurred())

		// then zone is offline
		Eventually(func() (string, error) {
			return global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zones")
		}, "40s", "1s").Should(ContainSubstring("Offline"))
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
		_, _, err = zoneUniversal.Exec("", "", AppModeCP, "pkill", "-9", "kuma-cp")
		Expect(err).ToNot(HaveOccurred())

		// then zone is offline
		Eventually(func() (string, error) {
			return global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zones")
		}, "30s", "1s").Should(ContainSubstring("Offline"))
	})

	It("should mark zone ingress as offline after it is killed forcefully when zone control-plane is down", func() {
		// deploy zone-ingress and wait while it's started
		Expect(IngressUniversal(global.GetKuma().GenerateZoneIngressToken)(zoneUniversal)).To(Succeed())
		Eventually(func() (string, error) {
			return global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zone-ingresses")
		}, "30s", "1s").Should(ContainSubstring("Online"))

		g, ok := zoneUniversal.(*UniversalCluster)
		Expect(ok).To(BeTrue())

		kumaCP := g.GetApp(AppModeCP)
		Expect(kumaCP).ToNot(BeNil())

		// when Zone CP is killed
		_, _, err := zoneUniversal.Exec("", "", AppModeCP, "pkill", "-9", "kuma-cp")
		Expect(err).ToNot(HaveOccurred())

		// and zone-ingress is killed while Zone CP is down
		_, _, err = zoneUniversal.Exec("", "", AppIngress, "pkill", "-9", "envoy")
		Expect(err).ToNot(HaveOccurred())

		// and Zone CP is restarted
		Expect(kumaCP.ReStart()).Should(Succeed())
		Eventually(zoneUniversal.VerifyKuma, "30s", "1s").ShouldNot(HaveOccurred())

		// then zone-ingress is offline
		Eventually(func() (string, error) {
			return zoneUniversal.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zone-ingresses")
		}, "40s", "1s").Should(ContainSubstring("Offline"))
	})

	// Disabled because of flakes: https://github.com/kumahq/kuma/issues/9345
	XIt("should mark zone as offline when zone control-plane is down", func() {
		// given zone connected to global
		Eventually(func() (string, error) {
			return global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zones")
		}, "30s", "1s").Should(ContainSubstring("Online"))

		// when Zone CP is killed
		_, _, err := zoneUniversal.Exec("", "", AppModeCP, "pkill", "-9", "kuma-cp")
		Expect(err).ToNot(HaveOccurred())

		// then zone is offline immediately
		Eventually(func() (string, error) {
			return global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zones")
		}, "10s", "1s").Should(ContainSubstring("Offline"))
	})
}
