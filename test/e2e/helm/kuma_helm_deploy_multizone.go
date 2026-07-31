package helm

import (
	"fmt"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/random"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/pkg/config/core"
	. "github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/client"
	"github.com/kumahq/kuma/v3/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/v3/test/framework/deployments/testserver"
)

// ZoneWithHelmChartAndUniversalGlobal verifies the Helm chart's multizone Zone
// install path against a Universal Global, the only supported Global topology
// now that Global-on-Kubernetes has been removed (see #17270, #17272).
func ZoneWithHelmChartAndUniversalGlobal() {
	var globalCluster, zoneCluster Cluster

	BeforeAll(func() {
		globalCluster = NewUniversalCluster(NewTestingT(), Kuma1, Silent)
		zoneCluster = NewK8sCluster(NewTestingT(), Kuma2, Silent).
			WithTimeout(6 * time.Second).
			WithRetries(60)

		releaseName := fmt.Sprintf(
			"kuma-%s",
			strings.ToLower(random.UniqueID()),
		)

		err := NewClusterSetup().
			Install(Kuma(core.Global)).
			Setup(globalCluster)
		Expect(err).ToNot(HaveOccurred())

		global := globalCluster.GetKuma()
		Expect(global).ToNot(BeNil())

		err = NewClusterSetup().
			Install(Kuma(core.Zone,
				WithInstallationMode(HelmInstallationMode),
				WithHelmReleaseName(releaseName),
				WithGlobalAddress(global.GetKDSServerAddress()),
				WithHelmOpt("ingress.enabled", "true"),
			)).
			Install(NamespaceWithSidecarInjection(TestNamespace)).
			Install(Parallel(
				democlient.Install(democlient.WithNamespace(TestNamespace), democlient.WithMesh("default")),
				testserver.Install(),
			)).
			Setup(zoneCluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugUniversal(globalCluster, "default")
		DebugKube(zoneCluster, "default", TestNamespace)
	})

	E2EAfterAll(func() {
		ControlPlaneAssertions(globalCluster)
		ControlPlaneAssertions(zoneCluster)
		Expect(zoneCluster.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(zoneCluster.DeleteKuma()).To(Succeed())
		Expect(globalCluster.DeleteKuma()).To(Succeed())
		Expect(globalCluster.DismissCluster()).To(Succeed())
		Expect(zoneCluster.DismissCluster()).To(Succeed())
	})

	It("should deploy Zone via Helm chart against a Universal Global", func() {
		// mesh is synced to zone
		Eventually(func(g Gomega) {
			output, err := zoneCluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "meshes")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(output).To(ContainSubstring("default"))
		}, "5s", "500ms").Should(Succeed())

		// and dataplanes are synced to global
		Eventually(func(g Gomega) {
			out, _, err := globalCluster.GetKuma().Exec("curl", "--fail", "--show-error", "http://localhost:5681/dataplanes")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(out).Should(ContainSubstring("demo-client"))
		}, "30s", "1s").Should(Succeed())
	})

	It("communication in between apps in zone works", func() {
		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(zoneCluster, "demo-client", "http://test-server.kuma-test.svc.cluster.local",
				client.FromKubernetesPod(TestNamespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
		}, "30s", "1s").Should(Succeed())
	})
}
