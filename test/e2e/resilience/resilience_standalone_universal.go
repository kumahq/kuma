package resilience

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/postgres"
)

func ResilienceStandaloneUniversal() {
	var universalCluster Cluster
	var optsUniversal = KumaUniversalDeployOpts

	BeforeEach(func() {
		clusters, err := NewUniversalClusters([]string{Kuma1}, Silent)
		Expect(err).ToNot(HaveOccurred())

		universalCluster = clusters.GetCluster(Kuma1)

		err = postgres.Install(Kuma1)(universalCluster)
		Expect(err).ToNot(HaveOccurred())

		optsUniversal = []DeployOptionsFunc{
			WithPostgres(postgres.From(universalCluster, Kuma1).GetEnvVars()),
		}

		err = Kuma(core.Standalone, optsUniversal...)(universalCluster)
		Expect(err).ToNot(HaveOccurred())

		demoClientToken, err := universalCluster.GetKuma().GenerateDpToken("default", "demo-client")
		Expect(err).ToNot(HaveOccurred())

		err = universalCluster.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		err = DemoClientUniversal(AppModeDemoClient, "default", demoClientToken, WithTransparentProxy(true))(universalCluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterEach(func() {
		Expect(universalCluster.DeleteKuma(optsUniversal...)).To(Succeed())
		Expect(universalCluster.DismissCluster()).To(Succeed())
	})

	It("should mark dataplane as offline after kuma-dp is killed forcefully while kuma-cp is down", func() {
		// given data plane proxy is connected to kuma-cp
		Eventually(func() (string, error) {
			return universalCluster.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "dataplanes")
		}, "30s", "1s").Should(ContainSubstring("Online"))

		u, ok := universalCluster.(*UniversalCluster)
		Expect(ok).To(BeTrue())

		kumaCP := u.GetApp(AppModeCP)
		Expect(kumaCP).ToNot(BeNil())

		// when kuma-cp is killed
		_, _, err := universalCluster.Exec("", "", AppModeCP, "pkill", "-9", "kuma-cp")
		Expect(err).ToNot(HaveOccurred())

		// and data plane proxy is killed while kuma-cp is down
		_, _, err = universalCluster.Exec("", "", "demo-client", "pkill", "-9", "envoy")
		Expect(err).ToNot(HaveOccurred())

		// and kuma-cp is restarted
		Expect(kumaCP.ReStart()).Should(Succeed())

		Eventually(universalCluster.VerifyKuma, "30s", "1s").ShouldNot(HaveOccurred())

		// then data plane proxy is offline
		Eventually(func() (string, error) {
			return universalCluster.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "dataplanes")
		}, "40s", "1s").Should(ContainSubstring("Offline"))
	})
}
