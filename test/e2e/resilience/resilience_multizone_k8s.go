package resilience

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func ResilienceMultizoneK8s() {
	var global, zone1 *K8sCluster
	var optsGlobal, optsZone1 = KumaK8sDeployOpts, KumaK8sDeployOpts
	defer GinkgoRecover()
	Skip("Skipped until https://github.com/kumahq/kuma/issues/1001 is fixed and this tests run reliably")

	BeforeEach(func() {
		clusters, err := NewK8sClusters(
			[]string{Kuma1, Kuma2},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		// Global
		global = clusters.GetCluster(Kuma1).(*K8sCluster)
		err = NewClusterSetup().
			Install(Kuma(core.Global, optsGlobal...)).
			Setup(global)
		Expect(err).ToNot(HaveOccurred())
		err = global.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		globalCP := global.GetKuma()

		// Cluster 1
		zone1 = clusters.GetCluster(Kuma2).(*K8sCluster)
		optsZone1 = append(optsZone1, WithGlobalAddress(globalCP.GetKDSServerAddress()))

		err = NewClusterSetup().
			Install(Kuma(core.Zone, optsZone1...)).
			Install(NamespaceWithSidecarInjection(TestNamespace)).
			Setup(zone1)
		Expect(err).ToNot(HaveOccurred())
		err = zone1.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterEach(func() {
		err := zone1.DeleteKuma(optsZone1...)
		Expect(err).ToNot(HaveOccurred())
		err = zone1.DismissCluster()
		Expect(err).ToNot(HaveOccurred())

		err = global.DeleteKuma(optsGlobal...)
		Expect(err).ToNot(HaveOccurred())
		err = global.DismissCluster()
		Expect(err).ToNot(HaveOccurred())
	})

	It("should see global entities in zone after a zone restart", func() {
		// Create a mesh
		err := YamlK8s(`
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: my-resilience-mesh-before
`)(global)
		Expect(err).ToNot(HaveOccurred())

		// The mesh should make it to zone
		Eventually(func() string {
			output, err2 := zone1.GetKumactlOptions().RunKumactlAndGetOutput("get", "meshes")
			Expect(err2).ToNot(HaveOccurred())
			return output
		}).Should(ContainSubstring("my-resilience-mesh-before"))

		// Stop the zone
		Expect(zone1.ShutdownCP()).ToNot(HaveOccurred())

		// Create a mesh while the zone cp is down
		err = YamlK8s(`
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: my-resilience-mesh-when-down
`)(global)
		Expect(err).ToNot(HaveOccurred())

		// Start the zone back
		Expect(zone1.StartCP()).ToNot(HaveOccurred())
		// TODO remove once https://github.com/kumahq/kuma/issues/1001 is fixed
		time.Sleep(time.Second * 10)

		// Create a mesh now that the remote zone is backup
		err = YamlK8s(`
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: my-resilience-mesh-after-restart
`)(global)
		Expect(err).ToNot(HaveOccurred())

		// All 3 meshes should be in the zone cp
		Eventually(func() string {
			output, err2 := zone1.GetKumactlOptions().RunKumactlAndGetOutput("get", "meshes")
			Expect(err2).ToNot(HaveOccurred())
			return output
		}, "30s", "1s").Should(And(
			ContainSubstring("my-resilience-mesh-before"),
			ContainSubstring("my-resilience-mesh-when-down"),
			ContainSubstring("my-resilience-mesh-after-restart"),
		))
	})

	It("should see global entities in zone after a global restart", func() {
		// Create a mesh
		err := YamlK8s(`
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: my-resilience-mesh-before
`)(global)
		Expect(err).ToNot(HaveOccurred())

		// Check the mesh makes it to the zone cp
		Eventually(func() string {
			output, err2 := zone1.GetKumactlOptions().RunKumactlAndGetOutput("get", "meshes")
			Expect(err2).ToNot(HaveOccurred())
			return output
		}, "30s", "1s").Should(ContainSubstring("my-resilience-mesh-before"))

		// Stop global
		Expect(global.ShutdownCP()).ToNot(HaveOccurred())

		// The mesh is still present in zone1
		out, err := zone1.GetKumactlOptions().RunKumactlAndGetOutput("get", "meshes")
		Expect(err).ToNot(HaveOccurred())
		Expect(out).Should(ContainSubstring("my-resilience-mesh-before"))

		// Start back global
		Expect(global.StartCP()).ToNot(HaveOccurred())
		// TODO remove once https://github.com/kumahq/kuma/issues/1001 is fixed
		time.Sleep(time.Second * 10)

		// Create a mesh now that global is backup
		err = YamlK8s(`
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: my-resilience-mesh-after-restart
`)(global)
		Expect(err).ToNot(HaveOccurred())

		// Check the zone has both Meshes
		Eventually(func() string {
			output, err2 := zone1.GetKumactlOptions().RunKumactlAndGetOutput("get", "meshes")
			Expect(err2).ToNot(HaveOccurred())
			return output
		}, "2m", "1s").Should(And(
			ContainSubstring("my-resilience-mesh-before"),
			ContainSubstring("my-resilience-mesh-after-restart"),
		))
	})

	It("should see zone entities in global after a zone restart", func() {
		// Run an app
		err := testserver.Install(testserver.WithName("kds-before-restart"))(zone1)
		Expect(err).ToNot(HaveOccurred())

		// Check the dp goes to global
		Eventually(func() string {
			output, err2 := global.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplanes")
			Expect(err2).ToNot(HaveOccurred())
			return output
		}, "30s", "1s").Should(ContainSubstring("kds-before-restart"))

		// Stop the zone CP
		Expect(zone1.ShutdownCP()).ToNot(HaveOccurred())

		// The global should still has the dp
		out, err := global.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplanes")
		Expect(err).ToNot(HaveOccurred())
		Expect(out).To(ContainSubstring("kds-before-restart"))

		// Start again the zone CP
		Expect(zone1.StartCP()).ToNot(HaveOccurred())
		// TODO remove once https://github.com/kumahq/kuma/issues/1001 is fixed
		time.Sleep(time.Second * 10)

		// Start a new app
		err = testserver.Install(testserver.WithName("kds-after-restart"))(zone1)
		Expect(err).ToNot(HaveOccurred())

		// Check all 2 dps are in the local zone
		Eventually(func() string {
			output, err2 := zone1.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplanes")
			Expect(err2).ToNot(HaveOccurred())
			return output
		}, "30s", "1s").Should(And(
			ContainSubstring("kds-before-restart"),
			ContainSubstring("kds-after-restart"),
		))

		// Check all 2 dps are in the global zone
		Eventually(func() string {
			output, err2 := global.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplanes")
			Expect(err2).ToNot(HaveOccurred())
			return output
		}, "30s", "1s").Should(And(
			ContainSubstring("kds-before-restart"),
			ContainSubstring("kds-after-restart"),
		))
	})

	It("should see zone entities in global after a global restart", func() {
		// Start an app
		err := testserver.Install(testserver.WithName("kds-before-restart"))(zone1)
		Expect(err).ToNot(HaveOccurred())

		// Check the dp gets to global
		Eventually(func() string {
			output, err2 := global.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplanes")
			Expect(err2).ToNot(HaveOccurred())
			return output
		}, "30s", "1s").Should(ContainSubstring("kds-before-restart"))

		// Stop the global CP
		Expect(global.ShutdownCP()).ToNot(HaveOccurred())

		// Start an app while the global CP is down
		err = testserver.Install(testserver.WithName("kds-during-restart"))(zone1)
		Expect(err).ToNot(HaveOccurred())

		// Check the dp is in the zone
		Eventually(func() string {
			output, err2 := zone1.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplanes")
			Expect(err2).ToNot(HaveOccurred())
			return output
		}, "30s", "1s").Should(ContainSubstring("kds-during-restart"))

		// Start back the global CP
		Expect(global.StartCP()).ToNot(HaveOccurred())
		// TODO remove once https://github.com/kumahq/kuma/issues/1001 is fixed
		time.Sleep(time.Second * 10)

		// Start a new app
		err = testserver.Install(testserver.WithName("kds-after-restart"))(zone1)
		Expect(err).ToNot(HaveOccurred())

		// Check all 3 dps are in the local zone
		Eventually(func() string {
			output, err2 := zone1.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplanes")
			Expect(err2).ToNot(HaveOccurred())
			return output
		}, "30s", "1s").Should(And(
			ContainSubstring("kds-before-restart"),
			ContainSubstring("kds-during-restart"),
			ContainSubstring("kds-after-restart"),
		))
		// Check all 3 dps are in the global zone
		Eventually(func() string {
			output, err2 := global.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplanes")
			Expect(err2).ToNot(HaveOccurred())
			return output
		}, "30s", "1s").Should(And(
			ContainSubstring("kds-before-restart"),
			ContainSubstring("kds-during-restart"),
			ContainSubstring("kds-after-restart"),
		))
	})
}
