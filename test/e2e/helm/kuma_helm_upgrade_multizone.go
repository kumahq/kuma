package helm

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gruntwork-io/terratest/modules/random"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/pkg/config/core"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/system"
	meshtimeout "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	. "github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/api"
	"github.com/kumahq/kuma/v3/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/v3/test/framework/versions"
)

// UpgradingZoneWithHelmChart exercises upgrading a Kubernetes Zone from an old
// published Helm chart version to the current one, against a Universal Global
// on the current version throughout. Global itself can no longer be version-pinned
// or upgraded here: Global-on-Kubernetes was removed (#17270), Global is always
// Universal now, and the test framework has no mechanism to run an old Universal
// kuma-cp binary, only old Helm charts/images for Kubernetes clusters.
func UpgradingZoneWithHelmChart() {
	namespace := "helm-upgrade-ns"
	var global, zoneK8s, zoneUniversal Cluster
	var globalCP ControlPlane

	BeforeEach(func() {
		global = NewUniversalCluster(NewTestingT(), Kuma1, Silent)
		zoneK8s = NewK8sCluster(NewTestingT(), Kuma2, Silent).
			WithTimeout(6 * time.Second).
			WithRetries(60)
		zoneUniversal = NewUniversalCluster(NewTestingT(), Kuma3, Silent)

		Expect(NewClusterSetup().
			Install(Kuma(core.Global)).
			Setup(global)).To(Succeed())
		globalCP = global.GetKuma()
		Expect(globalCP).ToNot(BeNil())
	})

	AfterEachFailure(func() {
		DebugUniversal(global, "default")
		DebugKube(zoneK8s, "default", namespace)
		DebugUniversal(zoneUniversal, "default")
	})

	E2EAfterEach(func() {
		ControlPlaneAssertions(global)
		ControlPlaneAssertions(zoneK8s)
		ControlPlaneAssertions(zoneUniversal)
		grp := sync.WaitGroup{}
		grp.Add(3)
		go func() {
			defer GinkgoRecover()
			defer grp.Done()
			Expect(zoneUniversal.DismissCluster()).To(Succeed())
		}()
		go func() {
			defer GinkgoRecover()
			defer grp.Done()
			Expect(zoneK8s.DeleteNamespace(namespace)).To(Succeed())
			Expect(zoneK8s.DeleteKuma()).To(Succeed())
			Expect(zoneK8s.DismissCluster()).To(Succeed())
		}()
		go func() {
			defer GinkgoRecover()
			defer grp.Done()
			Expect(global.DeleteKuma()).To(Succeed())
			Expect(global.DismissCluster()).To(Succeed())
		}()
		grp.Wait()
	})
	DescribeTable("upgrade zone",
		func(version string) {
			releaseName := fmt.Sprintf("kuma-%s", strings.ToLower(random.UniqueID()))

			By("Install zone with version: " + version)
			err := NewClusterSetup().
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
			Expect(YamlUniversal(fmt.Sprintf(`
type: MeshTimeout
name: mt1
mesh: %s
spec:
  targetRef:
    kind: Mesh
  rules:
    - default:
        idleTimeout: 20s
        http:
          requestTimeout: 2s
          maxStreamDuration: 20s`, "default"))(global)).To(Succeed())

			Eventually(func(g Gomega) (int, error) {
				return NumberOfResources(zoneK8s, meshtimeout.MeshTimeoutResourceTypeDescriptor)
			}, "30s", "1s").Should(Equal(5), "meshtimeouts are not synced to zone")

			By("Sync DPPs from Zone to Global")
			err = NewClusterSetup().
				Install(NamespaceWithSidecarInjection(namespace)).
				Install(testserver.Install(testserver.WithNamespace(namespace))).Setup(zoneK8s)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func(g Gomega) (int, error) {
				return NumberOfResources(global, mesh.DataplaneResourceTypeDescriptor)
			}, "30s", "1s").Should(Equal(1), "dpp should be synced to global")

			By("deploy a new universal zone with latest version")
			err = NewClusterSetup().
				Install(Kuma(core.Zone, WithGlobalAddress(global.GetKuma().GetKDSServerAddress()))).
				Install(IngressUniversal(global.GetKuma().GenerateZoneIngressToken)).
				Setup(zoneUniversal)
			Expect(err).ToNot(HaveOccurred())

			// kumactl on zone CPs older than 2.11.0 exposes /zone-ingresses, renamed to
			// /zoneingresses afterwards; query the old path directly to avoid the mismatch.
			if versions.IsVersionLessThan(version, "2.11.0") {
				Eventually(func(g Gomega) (int, error) {
					return NumberOfResourcesByPath(zoneK8s, "/zone-ingresses")
				}, "30s", "1s").Should(Equal(2), "have remote and local zoneIngress")
			} else {
				Eventually(func(g Gomega) (int, error) {
					return NumberOfResources(zoneK8s, mesh.ZoneIngressResourceTypeDescriptor)
				}, "30s", "1s").Should(Equal(2), "have remote and local zoneIngress")
			}

			// Scale down ingress before upgrading so the pod never runs against a
			// mixed-version CP: an old replica could hand it enable_reuse_port=false,
			// then the upgraded CP flips it to true, which Envoy rejects indefinitely.
			By("scale down zone ingress before upgrade")
			Expect(zoneK8s.(*K8sCluster).StopZoneIngress()).To(Succeed())

			By("upgrade Zone")
			err = zoneK8s.(*K8sCluster).UpgradeKuma(core.Zone,
				WithHelmReleaseName(releaseName),
				WithHelmChartPath(Config.HelmChartPath),
				ClearNoHelmOpts(),
				WithHelmOpt("ingress.replicas", "0"),
			)
			Expect(err).ToNot(HaveOccurred())

			By("wait for upgraded zone CP to connect to global")
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
			}, "60s", "1s").Should(Succeed())

			By("start zone ingress after upgrade")
			Expect(zoneK8s.(*K8sCluster).StartZoneIngress()).To(Succeed())

			// The old ingress pod can remain visible to global for up to ~40s after the
			// upgrade (K8s graceful termination plus CP deregistration delay).
			Eventually(func(g Gomega) {
				zoneIngressesGlobal, err := NumberOfResources(global, mesh.ZoneIngressResourceTypeDescriptor)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(zoneIngressesGlobal).To(Equal(2))
			}, "3m", "1s").Should(Succeed())

			Eventually(func(g Gomega) {
				zoneIngressesK8sZone, err := NumberOfResources(zoneK8s, mesh.ZoneIngressResourceTypeDescriptor)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(zoneIngressesK8sZone).To(Equal(2))
				zoneIngressesUniversalZone, err := NumberOfResources(zoneUniversal, mesh.ZoneIngressResourceTypeDescriptor)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(zoneIngressesUniversalZone).To(Equal(2))
			}, "3m", "1s").Should(Succeed())

			Consistently(func(g Gomega) {
				policiesGlobal, err := NumberOfResources(global, meshtimeout.MeshTimeoutResourceTypeDescriptor)
				g.Expect(err).ToNot(HaveOccurred())
				policiesK8sZone, err := NumberOfResources(zoneK8s, meshtimeout.MeshTimeoutResourceTypeDescriptor)
				g.Expect(err).ToNot(HaveOccurred())
				policiesUniversalZone, err := NumberOfResources(zoneUniversal, meshtimeout.MeshTimeoutResourceTypeDescriptor)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(policiesGlobal).To(And(Equal(policiesUniversalZone), Equal(policiesK8sZone), Equal(5)))

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
			}, "30s", "1s").Should(Succeed())
		},
		EntryDescription("from version: %s"),
		SupportedVersionEntries(),
	)
}
