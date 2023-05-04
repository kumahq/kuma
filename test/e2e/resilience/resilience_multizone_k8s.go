package resilience

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func ResilienceMultizoneK8s() {
	var global, zone1 *K8sCluster

	BeforeAll(func() {
		// Global
		global = NewK8sCluster(NewTestingT(), Kuma1, Silent)
		Expect(NewClusterSetup().
			Install(Kuma(core.Global, WithCtlOpts(map[string]string{"--set": "controlPlane.terminationGracePeriodSeconds=5"}))).
			Setup(global)).To(Succeed())
		Expect(global.VerifyKuma()).To(Succeed())

		globalCP := global.GetKuma()

		// Cluster 1
		zone1 = NewK8sCluster(NewTestingT(), Kuma2, Silent)

		Expect(NewClusterSetup().
			Install(Kuma(core.Zone, WithGlobalAddress(globalCP.GetKDSServerAddress()), WithCtlOpts(map[string]string{"--set": "controlPlane.terminationGracePeriodSeconds=5"}))).
			Install(NamespaceWithSidecarInjection(TestNamespace)).
			Setup(zone1)).To(Succeed())
		Expect(zone1.VerifyKuma()).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(zone1.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(zone1.DeleteKuma()).To(Succeed())
		Expect(zone1.DismissCluster()).To(Succeed())

		Expect(global.DeleteKuma()).To(Succeed())
		Expect(global.DismissCluster()).To(Succeed())
	})

	It("should see global entities in zone after a zone restart", func() {
		// Create a mesh
		Expect(YamlK8s(`
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: before-zone-restart
`)(global)).To(Succeed())

		// The mesh should make it to zone
		Eventually(func() error {
			return zone1.GetKumactlOptions().RunKumactl("get", "mesh", "before-zone-restart")
		}, "30s", "1s").Should(Succeed())

		// Stop the zone
		Expect(zone1.StopControlPlane()).ToNot(HaveOccurred())

		// Create a mesh while the zone cp is down
		Expect(YamlK8s(`
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: when-zone-is-down
`)(global)).To(Succeed())

		// Start the zone back
		Expect(zone1.RestartControlPlane()).ToNot(HaveOccurred())
		Eventually(func() (string, error) {
			return global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zones")
		}, "30s", "1s").Should(ContainSubstring("Online"))

		// Create a mesh now that the remote zone is back up
		Expect(YamlK8s(`
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: after-zone-restart
`)(global)).To(Succeed())

		// All 3 meshes should be in the zone cp
		Eventually(func() (string, error) {
			return zone1.GetKumactlOptions().RunKumactlAndGetOutput("get", "meshes")
		}, "30s", "1s").Should(And(
			ContainSubstring("before-zone-restart"),
			ContainSubstring("when-zone-is-down"),
			ContainSubstring("after-zone-restart"),
		))
	})

	It("should see global entities in zone after a global restart", func() {
		// Create a mesh
		Expect(YamlK8s(`
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: before-global-restart
`)(global)).To(Succeed())

		// Check the mesh makes it to the zone cp
		Eventually(func() error {
			return zone1.GetKumactlOptions().RunKumactl("get", "mesh", "before-global-restart")
		}, "30s", "1s").Should(Succeed())

		// Stop global
		Expect(global.StopControlPlane()).To(Succeed())

		// The mesh is still present in zone1
		Expect(zone1.GetKumactlOptions().RunKumactl("get", "mesh", "before-global-restart")).To(Succeed())

		// Start back global
		Expect(global.RestartControlPlane()).To(Succeed())
		Eventually(func() (string, error) {
			return global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zones")
		}, "30s", "1s").Should(ContainSubstring("Online"))

		// Create a mesh now that global is backup
		Expect(YamlK8s(`
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: after-global-restart
`)(global)).To(Succeed())

		// Check the zone has both Meshes
		Eventually(func() (string, error) {
			return zone1.GetKumactlOptions().RunKumactlAndGetOutput("get", "meshes")
		}, "30s", "1s").Should(And(
			ContainSubstring("before-global-restart"),
			ContainSubstring("after-global-restart"),
		))
	})

	It("should see zone entities in global after a zone restart", func() {
		// Run an app
		Expect(testserver.Install(testserver.WithName("kds-before-zone-restart"))(zone1)).To(Succeed())

		// Check the dp goes to global
		Eventually(func() (string, error) {
			return global.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplanes")
		}, "30s", "1s").Should(ContainSubstring("kds-before-zone-restart"))

		// Stop the zone CP
		Expect(zone1.StopControlPlane()).To(Succeed())

		// The global should still has the dp
		out, err := global.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplanes")
		Expect(err).ToNot(HaveOccurred())
		Expect(out).To(ContainSubstring("kds-before-zone-restart"))

		// Start again the zone CP
		Expect(zone1.RestartControlPlane()).To(Succeed())
		Eventually(func() (string, error) {
			return global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zones")
		}, "30s", "1s").Should(ContainSubstring("Online"))

		// Start a new app
		Expect(testserver.Install(testserver.WithName("kds-after-zone-restart"))(zone1)).To(Succeed())

		// Check all 2 dps are in the local zone
		Eventually(func() (string, error) {
			return zone1.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplanes")
		}, "30s", "1s").Should(And(
			ContainSubstring("kds-before-zone-restart"),
			ContainSubstring("kds-after-zone-restart"),
		))

		// Check all 2 dps are in the global zone
		Eventually(func() (string, error) {
			return global.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplanes")
		}, "30s", "1s").Should(And(
			ContainSubstring("kds-before-zone-restart"),
			ContainSubstring("kds-after-zone-restart"),
		))
	})

	It("should see zone entities in global after a global restart", func() {
		// Start an app
		Expect(testserver.Install(testserver.WithName("kds-before-global-restart"))(zone1)).To(Succeed())

		// Check the dp gets to global
		Eventually(func() (string, error) {
			return global.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplanes")
		}, "30s", "1s").Should(ContainSubstring("kds-before-global-restart"))

		// Stop the global CP
		Expect(global.StopControlPlane()).To(Succeed())

		// Start an app while the global CP is down
		Expect(testserver.Install(testserver.WithName("kds-during-global-restart"))(zone1)).To(Succeed())

		// Check the dp is in the zone
		Eventually(func() (string, error) {
			return zone1.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplanes")
		}, "30s", "1s").Should(ContainSubstring("kds-during-global-restart"))

		// Start back the global CP
		Expect(global.RestartControlPlane()).To(Succeed())
		Eventually(func() (string, error) {
			return global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zones")
		}, "30s", "1s").Should(ContainSubstring("Online"))

		// Start a new app
		Expect(testserver.Install(testserver.WithName("kds-after-global-restart"))(zone1)).To(Succeed())

		// Check all 3 dps are in the local zone
		Eventually(func() (string, error) {
			return zone1.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplanes")
		}, "30s", "1s").Should(And(
			ContainSubstring("kds-before-global-restart"),
			ContainSubstring("kds-during-global-restart"),
			ContainSubstring("kds-after-global-restart"),
		))
		// Check all 3 dps are in the global zone
		Eventually(func() (string, error) {
			return global.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplanes")
		}, "30s", "1s").Should(And(
			ContainSubstring("kds-before-global-restart"),
			ContainSubstring("kds-during-global-restart"),
			ContainSubstring("kds-after-global-restart"),
		))
	})
}
