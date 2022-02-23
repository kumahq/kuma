package ownership

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

func MultizoneUniversal() {
	var global, zoneUniversal Cluster
	BeforeEach(func() {
		clusters, err := NewUniversalClusters([]string{Kuma1, Kuma2}, Silent)
		Expect(err).ToNot(HaveOccurred())

		// Global
		global = clusters.GetCluster(Kuma1)
		Expect(Kuma(core.Global)(global)).To(Succeed())

		// Cluster 1
		zoneUniversal = clusters.GetCluster(Kuma2)
		Expect(Kuma(core.Zone,
			WithGlobalAddress(global.GetKuma().GetKDSServerAddress()))(zoneUniversal),
		).To(Succeed())
	})

	E2EAfterEach(func() {
		Expect(zoneUniversal.DeleteKuma()).To(Succeed())
		Expect(zoneUniversal.DismissCluster()).To(Succeed())

		Expect(global.DeleteKuma()).To(Succeed())
		Expect(global.DismissCluster()).To(Succeed())
	})

	installZoneIngress := func() {
		ingressToken, err := global.GetKuma().GenerateZoneIngressToken(Kuma2)
		Expect(err).ToNot(HaveOccurred())
		Expect(IngressUniversal(ingressToken)(zoneUniversal)).To(Succeed())
	}

	installZoneEgress := func() {
		egressToken, err := global.GetKuma().GenerateZoneEgressToken(Kuma2)
		Expect(err).ToNot(HaveOccurred())
		Expect(EgressUniversal(egressToken)(zoneUniversal)).To(Succeed())
	}

	installDataplane := func() {
		token, err := global.GetKuma().GenerateDpToken("default", AppModeDemoClient)
		Expect(err).ToNot(HaveOccurred())
		Expect(DemoClientUniversal(AppModeDemoClient, "default", token)(zoneUniversal)).To(Succeed())
	}

	has := func(resourceURI string) func() bool {
		return func() bool {
			cmd := []string{"curl", "-v", "-m", "3", "--fail", "localhost:5681/" + resourceURI}
			stdout, _, err := global.ExecWithRetries("", "", AppModeCP, cmd...)
			Expect(err).ToNot(HaveOccurred())
			return strings.Contains(stdout, `"total": 1`)
		}
	}

	killKumaDP := func(appname string) {
		_, _, err := zoneUniversal.Exec("", "", appname, "pkill", "-9", "envoy")
		Expect(err).ToNot(HaveOccurred())
	}

	killZone := func() {
		_, _, err := zoneUniversal.Exec("", "", AppModeCP, "pkill", "-9", "kuma-cp")
		Expect(err).ToNot(HaveOccurred())
		Eventually(func() (string, error) {
			return global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zones")
		}, "30s", "1s").Should(ContainSubstring("Offline"))
		Expect(global.GetKumactlOptions().RunKumactl("delete", "zone", Kuma2)).To(Succeed())
	}

	It("should delete ZoneInsight when Zone is deleted", func() {
		Eventually(has("zones"), "30s", "1s").Should(BeTrue())
		Eventually(has("zone-insights"), "30s", "1s").Should(BeTrue())

		killZone()

		Eventually(has("zones"), "30s", "1s").Should(BeFalse())
		Eventually(has("zone-insights"), "30s", "1s").Should(BeFalse())
	})

	It("should delete ZoneIngressInsights when ZoneIngress is deleted", func() {
		installZoneIngress()

		Eventually(has("zone-ingresses"), "30s", "1s").Should(BeTrue())
		Eventually(has("zone-ingress-insights"), "30s", "1s").Should(BeTrue())

		killKumaDP(AppIngress)

		Eventually(has("zone-ingresses"), "30s", "1s").Should(BeFalse())
		Eventually(has("zone-ingress-insights"), "30s", "1s").Should(BeFalse())
	})

	It("should delete ZoneEgressInsights when ZoneEgress is deleted", func() {
		installZoneEgress()

		Eventually(has("zoneegresses"), "30s", "1s").Should(BeTrue())
		Eventually(has("zoneegressinsights"), "30s", "1s").Should(BeTrue())

		killKumaDP(AppEgress)

		Eventually(has("zoneegresses"), "30s", "1s").Should(BeFalse())
		Eventually(has("zoneegressinsights"), "30s", "1s").Should(BeFalse())
	})

	It("should delete DataplaneInsight when Dataplane is deleted", func() {
		installDataplane()

		Eventually(has("dataplanes"), "30s", "1s").Should(BeTrue())
		Eventually(has("dataplane-insights"), "30s", "1s").Should(BeTrue())

		killKumaDP(AppModeDemoClient)

		Eventually(has("dataplanes"), "30s", "1s").Should(BeFalse())
		Eventually(has("dataplane-insights"), "30s", "1s").Should(BeFalse())
	})
}
