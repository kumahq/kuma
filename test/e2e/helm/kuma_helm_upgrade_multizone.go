package helm

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gruntwork-io/terratest/modules/random"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	meshtimeout "github.com/kumahq/kuma/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/api"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func UpgradingWithHelmChartMultizone() {
	namespace := "helm-upgrade-ns"
	var global, zoneK8s, zoneUniversal Cluster
	var globalCP ControlPlane

	BeforeEach(func() {
		global = NewK8sCluster(NewTestingT(), Kuma1, Silent).
			WithTimeout(6 * time.Second).
			WithRetries(60)
		zoneK8s = NewK8sCluster(NewTestingT(), Kuma2, Silent).
			WithTimeout(6 * time.Second).
			WithRetries(60)
		zoneUniversal = NewUniversalCluster(NewTestingT(), Kuma3, Silent)
	})

	E2EAfterEach(func() {
		grp := sync.WaitGroup{}
		grp.Add(3)
		go func() {
			defer grp.Done()
			Expect(zoneUniversal.DismissCluster()).To(Succeed())
		}()
		go func() {
			defer grp.Done()
			Expect(zoneK8s.DeleteNamespace(namespace)).To(Succeed())
			Expect(zoneK8s.DeleteKuma()).To(Succeed())
			Expect(zoneK8s.DismissCluster()).To(Succeed())
		}()
		go func() {
			defer grp.Done()
			Expect(global.DeleteKuma()).To(Succeed())
			Expect(global.DismissCluster()).To(Succeed())
		}()
		grp.Wait()
	})
	DescribeTable("upgrade helm multizone",
		func(version string) {
			releaseName := fmt.Sprintf("kuma-%s", strings.ToLower(random.UniqueId()))
			By("Install global with version: " + version)
			err := NewClusterSetup().
				Install(Kuma(core.Global,
					WithInstallationMode(HelmInstallationMode),
					WithHelmChartPath(Config.HelmChartName),
					WithHelmReleaseName(releaseName),
					WithHelmChartVersion(version),
					WithoutHelmOpt("global.image.tag"),
				)).
				Setup(global)
			Expect(err).ToNot(HaveOccurred())

			globalCP = global.GetKuma()
			Expect(globalCP).ToNot(BeNil())

			By("Install zone with version: " + version)
			err = NewClusterSetup().
				Install(Kuma(core.Zone,
					WithInstallationMode(HelmInstallationMode),
					WithHelmChartPath(Config.HelmChartName),
					WithHelmReleaseName(releaseName),
					WithHelmChartVersion(version),
					WithGlobalAddress(globalCP.GetKDSServerAddress()),
					WithHelmOpt("ingress.enabled", "true"),
					WithoutHelmOpt("global.image.tag"),
				)).
				Setup(zoneK8s)
			Expect(err).ToNot(HaveOccurred())

			By("Sync policies from Global to Zone")
			// when apply policy on Global
			Expect(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: mt1
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Mesh
  from:
    - targetRef:
        kind: Mesh
      default:
        idleTimeout: 20s
        http:
          requestTimeout: 2s
          maxStreamDuration: 20s`, Config.KumaNamespace, "default"))(global)).To(Succeed())

			Eventually(func(g Gomega) (int, error) {
				return NumberOfResources(zoneK8s, meshtimeout.MeshTimeoutResourceTypeDescriptor)
			}, "30s", "1s").Should(Equal(3), "meshtimeouts are not synced to zone")

			By("Sync DPPs from Zone to Global")
			// when start test server on Zone
			err = NewClusterSetup().
				Install(NamespaceWithSidecarInjection(namespace)).
				Install(testserver.Install(testserver.WithNamespace(namespace))).Setup(zoneK8s)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func(g Gomega) (int, error) {
				return NumberOfResources(global, mesh.DataplaneResourceTypeDescriptor)
			}, "30s", "1s").Should(Equal(1), "dpp should be synced to global")

			By("upgrade global")
			err = global.(*K8sCluster).UpgradeKuma(core.Global,
				WithHelmReleaseName(releaseName),
				WithHelmChartPath(Config.HelmChartPath),
				ClearNoHelmOpts(),
			)
			Expect(err).ToNot(HaveOccurred())

			By("deploy a new universal zone with latest version")
			err = NewClusterSetup().
				Install(Kuma(core.Zone, WithGlobalAddress(global.GetKuma().GetKDSServerAddress()))).
				Install(IngressUniversal(global.GetKuma().GenerateZoneIngressToken)).
				Setup(zoneUniversal)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func(g Gomega) (int, error) {
				return NumberOfResources(zoneK8s, mesh.ZoneIngressResourceTypeDescriptor)
			}, "30s", "1s").Should(Equal(2), "have remote and local zoneIngress")

			By("upgrade Zone")
			// when
			err = zoneK8s.(*K8sCluster).UpgradeKuma(core.Zone,
				WithHelmReleaseName(releaseName),
				WithHelmChartPath(Config.HelmChartPath),
				ClearNoHelmOpts(),
			)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func(g Gomega) {
				result := &system.ZoneInsightResource{}
				api.FetchResource(g, global, result, "", "kuma-2")
				g.Expect(len(result.Spec.Subscriptions)).To(BeNumerically(">", 1))
				newZoneConnected := false
				for _, sub := range result.Spec.Subscriptions {
					if sub.Version.KumaCp.Version != version {
						newZoneConnected = true
						break
					}
				}
				g.Expect(newZoneConnected).To(BeTrue())
			}, "30s", "100ms").Should(Succeed())

			// then
			Consistently(func(g Gomega) {
				policiesGlobal, err := NumberOfResources(global, meshtimeout.MeshTimeoutResourceTypeDescriptor)
				g.Expect(err).ToNot(HaveOccurred())
				policiesK8sZone, err := NumberOfResources(zoneK8s, meshtimeout.MeshTimeoutResourceTypeDescriptor)
				g.Expect(err).ToNot(HaveOccurred())
				policiesUniversalZone, err := NumberOfResources(zoneUniversal, meshtimeout.MeshTimeoutResourceTypeDescriptor)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(policiesGlobal).To(And(Equal(policiesUniversalZone), Equal(policiesK8sZone), Equal(3)))

				dppsGlobal, err := NumberOfResources(global, mesh.DataplaneResourceTypeDescriptor)
				g.Expect(err).ToNot(HaveOccurred())
				dppsK8sZone, err := NumberOfResources(zoneK8s, mesh.DataplaneResourceTypeDescriptor)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(dppsGlobal).To(And(Equal(dppsK8sZone), Equal(1)))
				// Dpps don't get copied to other zones
				dppsUniversalZone, err := NumberOfResources(zoneUniversal, mesh.DataplaneResourceTypeDescriptor)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(dppsUniversalZone).To(Equal(0))

				zoneIngressesGlobal, err := NumberOfResources(global, mesh.ZoneIngressResourceTypeDescriptor)
				g.Expect(err).ToNot(HaveOccurred())
				zoneIngressesK8sZone, err := NumberOfResources(zoneK8s, mesh.ZoneIngressResourceTypeDescriptor)
				g.Expect(err).ToNot(HaveOccurred())
				zoneIngressesUniversalZone, err := NumberOfResources(zoneUniversal, mesh.ZoneIngressResourceTypeDescriptor)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(zoneIngressesGlobal).To(And(Equal(2), Equal(zoneIngressesUniversalZone), Equal(zoneIngressesK8sZone)))
			}, "5s", "100ms").Should(Succeed())
		},
		EntryDescription("from version: %s"),
		SupportedVersionEntries(),
	)
}
