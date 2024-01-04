package sync

import (
	"fmt"
	"strings"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/kds/hash"
	"github.com/kumahq/kuma/test/framework"
	. "github.com/kumahq/kuma/test/framework"
)

func SyncLegacy() {
	var zone1, zone2, global *UniversalCluster
	const clusterNameGlobal = "kuma-sync-v2-global"
	const clusterName1 = "kuma-sync-v2-1"
	const clusterName2 = "kuma-sync-v2-2"
	meshName := "sync-v2"

	BeforeAll(func() {
		// Global
		global = NewUniversalCluster(NewTestingT(), clusterNameGlobal, Silent)
		err := NewClusterSetup().
			Install(Kuma(core.Global, framework.KumaDeploymentOptionsFromConfig(framework.Config.KumaCpConfig.Multizone.Global)...)).
			Install(MTLSMeshUniversal(meshName)).
			Setup(global)
		Expect(err).ToNot(HaveOccurred())

		globalCP := global.GetKuma()

		wg := sync.WaitGroup{}
		wg.Add(2)

		// Zone1 cluster
		zone1Options := append(
			[]framework.KumaDeploymentOption{
				WithEnv("KUMA_EXPERIMENTAL_KDS_DELTA_ENABLED", "false"),
				WithGlobalAddress(globalCP.GetKDSServerAddress()),
				WithHDS(false),
			},
			framework.KumaDeploymentOptionsFromConfig(framework.Config.KumaCpConfig.Multizone.UniZone1)...,
		)
		zone1 = NewUniversalCluster(NewTestingT(), clusterName1, Silent)
		go func() {
			defer GinkgoRecover()
			defer wg.Done()
			err := NewClusterSetup().
				Install(Kuma(core.Zone, zone1Options...)).
				Install(IngressUniversal(globalCP.GenerateZoneIngressToken)).
				Install(EgressUniversal(globalCP.GenerateZoneEgressToken)).
				Install(DemoClientUniversal("demo-client", meshName)).
				Setup(zone1)
			Expect(err).ToNot(HaveOccurred())
		}()

		// Zone2 cluster
		zone2Options := append(
			[]framework.KumaDeploymentOption{
				WithEnv("KUMA_EXPERIMENTAL_KDS_DELTA_ENABLED", "false"),
				WithGlobalAddress(globalCP.GetKDSServerAddress()),
				WithHDS(false),
			},
			framework.KumaDeploymentOptionsFromConfig(framework.Config.KumaCpConfig.Multizone.UniZone2)...,
		)
		zone2 = NewUniversalCluster(NewTestingT(), clusterName2, Silent)
		go func() {
			defer GinkgoRecover()
			defer wg.Done()
			err := NewClusterSetup().
				Install(Kuma(core.Zone, zone2Options...)).
				Install(IngressUniversal(globalCP.GenerateZoneIngressToken)).
				Install(EgressUniversal(globalCP.GenerateZoneEgressToken)).
				Install(TestServerUniversal("test-server", meshName)).
				Setup(zone2)
			Expect(err).ToNot(HaveOccurred())
		}()
		wg.Wait()
	})
	E2EAfterAll(func() {
		Expect(zone2.DeleteMeshApps(meshName)).To(Succeed())
		Expect(zone2.DismissCluster()).To(Succeed())
		Expect(zone1.DeleteMeshApps(meshName)).To(Succeed())
		Expect(zone1.DismissCluster()).To(Succeed())
		Expect(global.DismissCluster()).To(Succeed())
	})

	It("should show zones as online", func() {
		Eventually(func(g Gomega) {
			out, err := global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zones")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(strings.Count(out, "Online")).To(Equal(2))
		}, "30s", "1s").Should(Succeed())
	})

	Context("from Global to Zone", func() {
		universalPolicyNamed := func(name string, weight int, meshName string) string {
			return fmt.Sprintf(`
type: TrafficRoute
mesh: %s
name: %s
sources:
  - match:
      kuma.io/service: '*'
destinations:
  - match:
      kuma.io/service: '*'
conf:
  split:
    - weight: %d
      destination:
        kuma.io/service: '*'`, meshName, name, weight)
		}

		policySyncedToZones := func(name string) {
			Eventually(func() (string, error) {
				return zone1.GetKumactlOptions().RunKumactlAndGetOutput("get", "traffic-routes", "-m", meshName)
			}, "30s", "1s").Should(ContainSubstring(name))
			Eventually(func() (string, error) {
				return zone2.GetKumactlOptions().RunKumactlAndGetOutput("get", "traffic-routes", "-m", meshName)
			}, "30s", "1s").Should(ContainSubstring(name))
		}

		It("should sync policy creation", func() {
			// given
			name := "tr-synced"

			// when
			Expect(global.Install(YamlUniversal(universalPolicyNamed(name, 100, meshName)))).To(Succeed())

			// then
			policySyncedToZones(name)
		})

		It("should sync policy update", func() {
			// given
			name := "tr-update"
			Expect(global.Install(YamlUniversal(universalPolicyNamed(name, 100, meshName)))).To(Succeed())
			policySyncedToZones(name)

			// when
			Expect(global.Install(YamlUniversal(universalPolicyNamed(name, 101, meshName)))).To(Succeed())

			// then
			hashedName := hash.SyncedNameInZone(meshName, "tr-update")
			for _, zone := range []*UniversalCluster{zone1, zone2} {
				Eventually(func(g Gomega) {
					output, err := zone.GetKumactlOptions().RunKumactlAndGetOutput("get", "traffic-route", hashedName, "-m", meshName, "-o", "yaml")
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(output).To(ContainSubstring(`weight: 101`))
					g.Expect(output).To(ContainSubstring(`kuma.io/display-name: ` + name))
				}, "30s", "1s").Should(Succeed())
			}
		})

		Context("Deny", func() {
			name := "tr-denied"
			BeforeAll(func() {
				Expect(global.Install(YamlUniversal(universalPolicyNamed(name, 100, meshName)))).To(Succeed())
				policySyncedToZones(name)
			})

			It("should deny creating policy on Universal Zone CP", func() {
				err := zone1.GetKumactlOptions().KumactlApplyFromString(universalPolicyNamed("denied", 100, meshName))
				Expect(err).To(HaveOccurred())
			})

			It("should deny update on Universal Zone CP", func() {
				policyUpdate := universalPolicyNamed(name, 101, meshName)
				err := zone1.GetKumactlOptions().KumactlApplyFromString(policyUpdate)
				Expect(err).To(HaveOccurred())
			})

			It("should deny delete on Universal Zone CP", func() {
				err := zone1.GetKumactlOptions().RunKumactl("delete", "traffic-route", name, "-m", meshName)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("from Remote to Global", func() {
		It("should sync Zone Ingress", func() {
			Eventually(func(g Gomega) {
				out, err := global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zone-ingresses")
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(strings.Count(out, "Online")).To(Equal(2))
			}, "30s", "1s").Should(Succeed())
		})

		It("should sync Zone Egresses", func() {
			Eventually(func(g Gomega) {
				out, err := global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zoneegresses")
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(strings.Count(out, "Online")).To(Equal(2))
			}, "30s", "1s").Should(Succeed())
		})

		It("should sync Dataplane with insight", func() {
			Eventually(func(g Gomega) {
				out, err := global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "dataplanes", "--mesh", meshName)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(strings.Count(out, "Online")).To(Equal(2))
			}, "30s", "1s").Should(Succeed())
		})
	})
}
