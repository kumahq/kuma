package ownership

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
)

func MultizoneUniversal() {
	const clusterName1 = "kuma-owner1"
	const clusterName2 = "kuma-owner2"

	var global, zoneUniversal Cluster
	BeforeEach(func() {
		// Global
		global = NewUniversalCluster(NewTestingT(), clusterName1, Silent)
		Expect(Kuma(core.Global)(global)).To(Succeed())

		// Cluster 1
		zoneUniversal = NewUniversalCluster(NewTestingT(), clusterName2, Silent)
		Expect(Kuma(core.Zone,
			WithGlobalAddress(global.GetKuma().GetKDSServerAddress()))(zoneUniversal),
		).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(global, "default")
		DebugUniversal(zoneUniversal, "default")
	})

	E2EAfterEach(func() {
		Expect(zoneUniversal.DeleteKuma()).To(Succeed())
		Expect(zoneUniversal.DismissCluster()).To(Succeed())

		Expect(global.DeleteKuma()).To(Succeed())
		Expect(global.DismissCluster()).To(Succeed())
	})

	installZoneIngress := func() {
		Expect(IngressUniversal(global.GetKuma().GenerateZoneIngressToken)(zoneUniversal)).To(Succeed())
	}

	installZoneEgress := func() {
		Expect(EgressUniversal(global.GetKuma().GenerateZoneEgressToken)(zoneUniversal)).To(Succeed())
	}

	installDataplane := func() {
		Expect(DemoClientUniversal(AppModeDemoClient, "default")(zoneUniversal)).To(Succeed())
	}

	has := func(resourceURI string) func() bool {
		return func() bool {
			stdout, _, err := client.CollectResponse(global, AppModeCP, "localhost:5681/"+resourceURI)
			Expect(err).ToNot(HaveOccurred())
			return strings.Contains(stdout, `"total": 1`)
		}
	}

	killKumaDP := func(appname string) {
		Expect(zoneUniversal.(*UniversalCluster).Kill(appname, "envoy")).To(Succeed())
	}

	killZone := func() {
		Expect(zoneUniversal.(*UniversalCluster).Kill(AppModeCP, "kuma-cp run")).To(Succeed())
		Eventually(func() (string, error) {
			return global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zones")
		}, "30s", "1s").Should(ContainSubstring("Offline"))
		Expect(global.GetKumactlOptions().RunKumactl("delete", "zone", clusterName2)).To(Succeed())
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
