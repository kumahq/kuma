package compatibility

import (
	"fmt"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

// Ensure that the upstream Kuma help repository is configured
// and refreshed. This is needed for helm to be able to pull the
// OldChart version of the Kuma helm chart.
var _ = E2EBeforeSuite(func() {
	t := NewTestingT()
	opts := helm.Options{}

	// Adding the same repo multiple times is idempotent. The
	// `--force-update` flag prevents helm emitting an error
	// in this case.
	Expect(helm.RunHelmCommandAndGetOutputE(t, &opts,
		"repo", "add", "--force-update", "kuma", Config.HelmRepoUrl)).Error().To(BeNil())

	Expect(helm.RunHelmCommandAndGetOutputE(t, &opts, "repo", "update")).Error().To(BeNil())
})

func CpCompatibilityMultizoneKubernetes() {
	var globalCluster Cluster
	var globalReleaseName string

	var zoneCluster Cluster
	var zoneReleaseName string

	BeforeEach(func() {
		// Global CP
		globalCluster = NewK8sCluster(NewTestingT(), Kuma1, Silent).
			WithTimeout(6 * time.Second).
			WithRetries(60)

		globalReleaseName = fmt.Sprintf(
			"kuma-%s",
			strings.ToLower(random.UniqueId()),
		)

		E2EDeferCleanup(func() {
			Expect(globalCluster.DeleteKuma()).To(Succeed())
			Expect(globalCluster.DismissCluster()).To(Succeed())
		})

		// Zone CP
		zoneCluster = NewK8sCluster(NewTestingT(), Kuma2, Silent).
			WithTimeout(6 * time.Second).
			WithRetries(60)

		zoneReleaseName = fmt.Sprintf(
			"kuma-%s",
			strings.ToLower(random.UniqueId()),
		)

		E2EDeferCleanup(func() {
			Expect(zoneCluster.DeleteKuma()).To(Succeed())
			Expect(zoneCluster.DismissCluster()).To(Succeed())
		})
	})

	DescribeTable("Cross version check", func(globalConf, zoneConf []KumaDeploymentOption) {
		// Start a global
		err := NewClusterSetup().
			Install(Kuma(core.Global,
				append(globalConf, WithEnv("KUMA_API_SERVER_AUTH_ALLOW_FROM_LOCALHOST", "true"),
					WithInstallationMode(HelmInstallationMode),
					WithHelmReleaseName(globalReleaseName))...,
			)).
			Setup(globalCluster)
		Expect(err).ToNot(HaveOccurred())

		// Start a zone
		err = NewClusterSetup().
			Install(Kuma(core.Zone,
				append(zoneConf,
					WithEnv("KUMA_API_SERVER_AUTH_ALLOW_FROM_LOCALHOST", "true"),
					WithInstallationMode(HelmInstallationMode),
					WithHelmReleaseName(zoneReleaseName),
					WithGlobalAddress(globalCluster.GetKuma().GetKDSServerAddress()),
					WithHelmOpt("ingress.enabled", "true"))...,
			)).
			Install(NamespaceWithSidecarInjectionOnAnnotation(TestNamespace)).
			Setup(zoneCluster)
		Expect(err).ToNot(HaveOccurred())

		// and new resource is created on Global
		err = YamlK8s(`
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: demo
`)(globalCluster)

		// then the resource is synchronized when old remote is connected (KDS is backwards compatible)
		Expect(err).ToNot(HaveOccurred())
		Eventually(func() (string, error) {
			return k8s.RunKubectlAndGetOutputE(zoneCluster.GetTesting(), zoneCluster.GetKubectlOptions(), "get", "meshes")
		}, "30s", "1s").Should(ContainSubstring("demo"))

		// when new resources is created on Zone
		err = DemoClientK8s("default")(zoneCluster)

		// then resource is synchronized to Global
		Expect(err).ToNot(HaveOccurred())
		Eventually(func() (string, error) {
			return k8s.RunKubectlAndGetOutputE(globalCluster.GetTesting(), globalCluster.GetKubectlOptions("default"), "get", "dataplanes")
		}, "30s", "1s").Should(ContainSubstring("demo-client"))
	}, Entry(
		"Sync new global and old zone",
		[]KumaDeploymentOption{
			WithInstallationMode(HelmInstallationMode),
			WithHelmChartPath(Config.HelmChartPath),
		},
		[]KumaDeploymentOption{
			WithHelmChartPath(Config.HelmChartName),
			WithoutHelmOpt("global.image.tag"),
			WithHelmChartVersion(Config.SuiteConfig.Compatibility.HelmVersion),
		},
	), Entry(
		"Sync old global and new zone",
		[]KumaDeploymentOption{
			WithHelmChartPath(Config.HelmChartName),
			WithoutHelmOpt("global.image.tag"),
			WithHelmChartVersion(Config.SuiteConfig.Compatibility.HelmVersion),
		},
		[]KumaDeploymentOption{
			WithInstallationMode(HelmInstallationMode),
			WithHelmChartPath(Config.HelmChartPath),
		},
	))
}
