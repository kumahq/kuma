package kdsauth

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/pkg/config/core"
	"github.com/kumahq/kuma/v3/pkg/tokens/builtin/zone"
	. "github.com/kumahq/kuma/v3/test/framework"
)

func ZoneToken() {
	const globalName = "kds-auth-global"
	const validTokenZone = "kds-auth-valid"
	const otherZoneTokenZone = "kds-auth-other"
	const noTokenZone = "kds-auth-none"

	var global *UniversalCluster
	var zones []*UniversalCluster
	zoneByName := map[string]*UniversalCluster{}

	setupZone := func(name string, opts ...KumaDeploymentOption) {
		cluster := NewUniversalCluster(NewTestingT(), name, Silent)
		zones = append(zones, cluster)
		zoneByName[name] = cluster
		opts = append(opts, WithGlobalAddress(global.GetKuma().GetKDSServerAddress()))
		Expect(NewClusterSetup().Install(Kuma(core.Zone, opts...)).Setup(cluster)).To(Succeed())
	}

	BeforeAll(func() {
		global = NewUniversalCluster(NewTestingT(), globalName, Silent)
		err := NewClusterSetup().
			Install(Kuma(core.Global, WithEnv("KUMA_MULTIZONE_GLOBAL_KDS_AUTH_TYPE", "zoneToken"))).
			Setup(global)
		Expect(err).ToNot(HaveOccurred())

		var token string
		Eventually(func(g Gomega) {
			token, err = global.GetKuma().GenerateZoneToken(validTokenZone, []string{zone.CPScope})
			g.Expect(err).ToNot(HaveOccurred())
		}, "30s", "1s").Should(Succeed())

		setupZone(validTokenZone, WithEnv("KUMA_MULTIZONE_ZONE_KDS_AUTH_TOKEN_INLINE", token))
		setupZone(otherZoneTokenZone, WithEnv("KUMA_MULTIZONE_ZONE_KDS_AUTH_TOKEN_INLINE", token))
		setupZone(noTokenZone)
	})

	AfterEachFailure(func() {
		DebugUniversal(global, "default")
		for _, cluster := range zones {
			DebugUniversal(cluster, "default")
		}
	})

	E2EAfterAll(func() {
		for _, cluster := range append(zones, global) {
			Expect(cluster.DeleteKuma()).To(Succeed())
			Expect(cluster.DismissCluster()).To(Succeed())
		}
	})

	inspectZones := func() (string, error) {
		return global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zones")
	}

	// the Zone CP logs the status its KDS stream was rejected with, a zone absent
	// because it crashed or is still starting does not have it
	cpLogs := func(name string) string {
		var logs strings.Builder
		for _, log := range zoneByName[name].GetKumaCPLogs() {
			logs.WriteString(log)
		}
		return logs.String()
	}

	It("should connect the zone with a token issued for it", func() {
		Eventually(func(g Gomega) {
			out, err := inspectZones()
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(out).To(And(ContainSubstring(validTokenZone), ContainSubstring("Online")))
		}, "30s", "1s").Should(Succeed())
	})

	DescribeTable("should reject a zone that does not authenticate",
		func(zoneName string, rejection string) {
			Eventually(func(g Gomega) {
				g.Expect(cpLogs(zoneName)).To(ContainSubstring(rejection))
			}, "30s", "1s").Should(Succeed())

			Consistently(func(g Gomega) {
				out, err := inspectZones()
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(out).ToNot(ContainSubstring(zoneName))
			}, "10s", "1s").Should(Succeed())
		},
		Entry("without a token", noTokenZone, "Zone CP did not provide a zone token"),
		Entry("with a token of another zone", otherZoneTokenZone,
			`token is signed for "`+validTokenZone+`" zone, but connected CP advertised as "`+otherZoneTokenZone+`"`),
	)
}
