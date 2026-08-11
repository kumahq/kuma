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
	meshzoneaddress_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshzoneaddress/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/system"
	meshtimeout "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	. "github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/api"
	"github.com/kumahq/kuma/v3/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/v3/test/framework/deployments/zoneproxy"
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
					WithHelmOpt("meshes[0].name", "default"),
					WithHelmOpt("meshes[0].ingress.enabled", "true"),
					// The chart defaults the ingress Service to LoadBalancer,
					// which never gets an address on k3d.
					WithHelmOpt("meshes[0].ingress.service.type", "NodePort"),
					WithoutHelmOpt("global.image.tag"),
				)).
				Install(WaitNumPods(Config.KumaNamespace, 1, meshZoneIngressApp)).
				Install(WaitPodsAvailable(Config.KumaNamespace, meshZoneIngressApp)).
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
			}, "30s", "1s").Should(Equal(3), "meshtimeouts are not synced to zone")

			By("Sync DPPs from Zone to Global")
			err = NewClusterSetup().
				Install(NamespaceWithSidecarInjection(namespace)).
				Install(testserver.Install(testserver.WithNamespace(namespace))).Setup(zoneK8s)
			Expect(err).ToNot(HaveOccurred())

			// The zone's own mesh zone ingress is a Dataplane too, so it counts
			// alongside the test server.
			Eventually(func(g Gomega) (int, error) {
				return NumberOfResources(global, mesh.DataplaneResourceTypeDescriptor)
			}, "60s", "1s").Should(Equal(2), "dpps should be synced to global")

			By("deploy a new universal zone with latest version")
			err = NewClusterSetup().
				Install(Kuma(core.Zone, WithGlobalAddress(global.GetKuma().GetKDSServerAddress()))).
				// A mesh zone proxy is a Dataplane in the default mesh, so the
				// mesh has to reach this zone over KDS before deploying it.
				Install(func(c Cluster) error {
					return WaitForMesh("default", []Cluster{c})
				}).
				Install(zoneproxy.Install(
					zoneproxy.WithMesh("default"),
					zoneproxy.WithIngress(),
				)).
				Setup(zoneUniversal)
			Expect(err).ToNot(HaveOccurred())

			// Mesh zone ingresses publish a MeshZoneAddress instead of a
			// ZoneIngress, and global mirrors each zone's onto the others, so
			// every cluster ends up seeing both.
			Eventually(func(g Gomega) (int, error) {
				return NumberOfResources(zoneK8s, meshzoneaddress_api.MeshZoneAddressResourceTypeDescriptor)
			}, "60s", "1s").Should(Equal(2), "have remote and local mesh zone ingress")

			// Scale down ingress before upgrading so the pod never runs against a
			// mixed-version CP: an old replica could hand it enable_reuse_port=false,
			// then the upgraded CP flips it to true, which Envoy rejects indefinitely.
			By("scale down zone ingress before upgrade")
			Expect(zoneK8s.(*K8sCluster).ScaleApp(Config.KumaNamespace, meshZoneIngressApp, 0)).To(Succeed())

			By("upgrade Zone")
			err = zoneK8s.(*K8sCluster).UpgradeKuma(core.Zone,
				WithHelmReleaseName(releaseName),
				WithHelmChartPath(Config.HelmChartPath),
				ClearNoHelmOpts(),
				WithHelmOpt("meshes[0].ingress.deployment.replicas", "0"),
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
			Expect(zoneK8s.(*K8sCluster).ScaleApp(Config.KumaNamespace, meshZoneIngressApp, 1)).To(Succeed())

			// The scaled-down ingress drops its MeshZoneAddress, and global can
			// keep the old pod's Dataplane visible for up to ~40s after the
			// upgrade (K8s graceful termination plus CP deregistration delay).
			Eventually(func(g Gomega) {
				addressesGlobal, err := NumberOfResources(global, meshzoneaddress_api.MeshZoneAddressResourceTypeDescriptor)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(addressesGlobal).To(Equal(2))
			}, "3m", "1s").Should(Succeed())

			Eventually(func(g Gomega) {
				addressesK8sZone, err := NumberOfResources(zoneK8s, meshzoneaddress_api.MeshZoneAddressResourceTypeDescriptor)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(addressesK8sZone).To(Equal(2))
				addressesUniversalZone, err := NumberOfResources(zoneUniversal, meshzoneaddress_api.MeshZoneAddressResourceTypeDescriptor)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(addressesUniversalZone).To(Equal(2))
			}, "3m", "1s").Should(Succeed())

			// Restarting the ingress replaces its Dataplane, and the one from
			// before the upgrade stays visible for a while, so wait for the
			// counts to settle before asserting they stay put.
			Eventually(func(g Gomega) {
				dppsK8sZone, err := NumberOfResources(zoneK8s, mesh.DataplaneResourceTypeDescriptor)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(dppsK8sZone).To(Equal(2))
				dppsGlobal, err := NumberOfResources(global, mesh.DataplaneResourceTypeDescriptor)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(dppsGlobal).To(Equal(3))
			}, "3m", "1s").Should(Succeed())

			Consistently(func(g Gomega) {
				policiesGlobal, err := NumberOfResources(global, meshtimeout.MeshTimeoutResourceTypeDescriptor)
				g.Expect(err).ToNot(HaveOccurred())
				policiesK8sZone, err := NumberOfResources(zoneK8s, meshtimeout.MeshTimeoutResourceTypeDescriptor)
				g.Expect(err).ToNot(HaveOccurred())
				policiesUniversalZone, err := NumberOfResources(zoneUniversal, meshtimeout.MeshTimeoutResourceTypeDescriptor)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(policiesGlobal).To(And(Equal(policiesUniversalZone), Equal(policiesK8sZone), Equal(3)))

				dppsK8sZone, err := NumberOfResources(zoneK8s, mesh.DataplaneResourceTypeDescriptor)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(dppsK8sZone).To(Equal(2))
				dppsUniversalZone, err := NumberOfResources(zoneUniversal, mesh.DataplaneResourceTypeDescriptor)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(dppsUniversalZone).To(Equal(1))
				dppsGlobal, err := NumberOfResources(global, mesh.DataplaneResourceTypeDescriptor)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(dppsGlobal).To(Equal(3))

				addressesGlobal, err := NumberOfResources(global, meshzoneaddress_api.MeshZoneAddressResourceTypeDescriptor)
				g.Expect(err).ToNot(HaveOccurred())
				addressesK8sZone, err := NumberOfResources(zoneK8s, meshzoneaddress_api.MeshZoneAddressResourceTypeDescriptor)
				g.Expect(err).ToNot(HaveOccurred())
				addressesUniversalZone, err := NumberOfResources(zoneUniversal, meshzoneaddress_api.MeshZoneAddressResourceTypeDescriptor)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(addressesGlobal).To(And(Equal(2), Equal(addressesUniversalZone), Equal(addressesK8sZone)))
			}, "30s", "1s").Should(Succeed())
		},
		EntryDescription("from version: %s"),
		// Only the last 2.14.x release is a supported upgrade path onto this
		// (3.0-line) master; older LTS releases predate APIs (e.g. MeshTimeout
		// 'rules') that this branch's KDS payloads now assume unconditionally,
		// and are not a real upgrade path straight to a new major version.
		SupportedVersionEntriesAtLeast("2.14.0"),
	)
}
