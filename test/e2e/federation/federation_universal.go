package federation

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/random"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func FederateKubeZoneCPToUniversalGlobal() {
	var global, zone Cluster
	var releaseName string

	BeforeAll(func() {
		global = NewUniversalCluster(NewTestingT(), Kuma4, Silent)
		zone = NewK8sCluster(NewTestingT(), Kuma2, Silent).
			WithTimeout(6 * time.Second).
			WithRetries(60)

		releaseName = fmt.Sprintf("kuma-%s", strings.ToLower(random.UniqueId()))

		err := NewClusterSetup().
			Install(Kuma(core.Global,
				WithSkipDefaultMesh(true),
			)).
			Setup(global)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(Kuma(core.Zone,
				WithInstallationMode(HelmInstallationMode),
				WithHelmReleaseName(releaseName),
			)).
			Install(NamespaceWithSidecarInjection(TestNamespace)).
			Install(MTLSMeshKubernetes("default")).
			Install(MeshTrafficPermissionAllowAllKubernetes("default")).
			Install(democlient.Install()).
			Install(testserver.Install()).
			Setup(zone)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugUniversal(global, "default")
		DebugKube(zone, "default", TestNamespace)
		PrintLogs(global, zone)
	})

	E2EAfterAll(func() {
		Expect(zone.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(zone.DeleteKuma()).To(Succeed())
		Expect(global.DeleteKuma()).To(Succeed())
		Expect(global.DismissCluster()).To(Succeed())
		Expect(zone.DismissCluster()).To(Succeed())
	})

	Context("on federation", func() {
		BeforeAll(func() {
			Eventually(func(g Gomega) {
				_, err := client.CollectEchoResponse(zone, "demo-client", "test-server",
					client.FromKubernetesPod(TestNamespace, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
			}, "30s", "1s").Should(Succeed())

			out, err := zone.GetKumactlOptions().RunKumactlAndGetOutput("export", "--profile", "federation", "--format", "universal")
			Expect(err).ToNot(HaveOccurred())

			tmpfile, err := os.CreateTemp("", "export-uni")
			Expect(err).ToNot(HaveOccurred())
			_, err = tmpfile.WriteString(out)
			Expect(err).ToNot(HaveOccurred())

			Expect(global.GetKumactlOptions().RunKumactl("apply", "-f", tmpfile.Name())).To(Succeed())
			err = zone.(*K8sCluster).UpgradeKuma(core.Zone,
				WithHelmReleaseName(releaseName),
				WithHelmWait(),
				WithGlobalAddress(global.GetKuma().GetKDSServerAddress()),
			)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should sync data plane proxies to global cp", func() {
			Eventually(func(g Gomega) {
				// we use API on localhost, because the auth data has changed on the global, so we would need to reconfigure kumactl
				out, _, err := global.GetKuma().Exec("curl", "--fail", "--show-error", "http://localhost:5681/dataplanes")
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(out).Should(ContainSubstring("demo-client"))
			}, "30s", "1s").Should(Succeed())
		})

		It("should sync data policies to global cp", func() {
			Eventually(func(g Gomega) {
				out, _, err := global.GetKuma().Exec("curl", "--fail", "--show-error", "http://localhost:5681/meshcircuitbreakers")
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(out).Should(ContainSubstring("mesh-circuit-breaker-all-default-zw856xvxdb7558d9"))
			}, "30s", "1s").Should(Succeed())
		})

		It("should not break the traffic", func() {
			Consistently(func(g Gomega) {
				_, err := client.CollectEchoResponse(zone, "demo-client", "test-server",
					client.FromKubernetesPod(TestNamespace, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
			}, "3s", "100s").Should(Succeed())
		})
	})
}
