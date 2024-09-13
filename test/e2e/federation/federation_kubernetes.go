package federation

import (
	"fmt"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func FederateKubeZoneCPToKubeGlobal() {
	var global, zone Cluster
	var releaseName string

	BeforeAll(func() {
		global = NewK8sCluster(NewTestingT(), Kuma1, Silent).
			WithTimeout(6 * time.Second).
			WithRetries(60)
		zone = NewK8sCluster(NewTestingT(), Kuma2, Silent).
			WithTimeout(6 * time.Second).
			WithRetries(60)

		releaseName = fmt.Sprintf("kuma-%s", strings.ToLower(random.UniqueId()))

		err := NewClusterSetup().
			Install(Kuma(core.Global,
				WithInstallationMode(HelmInstallationMode),
				WithHelmReleaseName(releaseName),
				WithEnv("KUMA_DEFAULTS_SKIP_MESH_CREATION", "true"),
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
		DebugKube(global, "default", Config.KumaNamespace)
		DebugKube(zone, "default", TestNamespace)
	})

	E2EAfterAll(func() {
		Expect(zone.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(global.DeleteKuma()).To(Succeed())
		Expect(zone.DeleteKuma()).To(Succeed())
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

			out, err := zone.GetKumactlOptions().RunKumactlAndGetOutput("export", "--profile", "federation", "--format", "kubernetes")
			Expect(err).ToNot(HaveOccurred())

			err = k8s.KubectlApplyFromStringE(global.GetTesting(), global.GetKubectlOptions(), out)
			Expect(err).ToNot(HaveOccurred())
			err = zone.(*K8sCluster).UpgradeKuma(core.Zone,
				WithHelmReleaseName(releaseName),
				WithHelmWait(),
				WithGlobalAddress(global.GetKuma().GetKDSServerAddress()),
			)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should sync data plane proxies to global cp", func() {
			Eventually(func(g Gomega) {
				out, err := k8s.RunKubectlAndGetOutputE(global.GetTesting(), global.GetKubectlOptions(), "get", "dataplanes", "-A")
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(out).Should(ContainSubstring("demo-client"))
			}, "120s", "1s").Should(Succeed())
		})

		It("should sync data policies to global cp", func() {
			Eventually(func(g Gomega) {
				out, err := k8s.RunKubectlAndGetOutputE(global.GetTesting(), global.GetKubectlOptions(), "get", "meshcircuitbreakers", "-A")
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(out).Should(ContainSubstring("mesh-circuit-breaker-all-default-zw856xvxdb7558d9"))
			}, "120s", "1s").Should(Succeed())
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
