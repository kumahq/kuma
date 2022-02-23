package helm

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
)

func UpgradingWithHelmChart() {
	var cluster Cluster

	AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}

		Expect(cluster.DeleteKuma()).To(Succeed())
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	type testCase struct {
		initialChartVersion string
	}

	DescribeTable(
		"should successfully upgrade Kuma via Helm",
		func(given testCase) {
			t := NewTestingT()
			cluster = NewK8sCluster(t, Kuma1, Silent).
				WithTimeout(6 * time.Second).
				WithRetries(60)

			// Sometimes it might be necessary to run some stuff on k8s first
			if Config.SuiteConfig.Helm.ExtraYamlPath != "" {
				k8s.KubectlApply(t, cluster.(*K8sCluster).GetKubectlOptions(), Config.SuiteConfig.Helm.ExtraYamlPath)
			}

			releaseName := fmt.Sprintf(
				"kuma-%s",
				strings.ToLower(random.UniqueId()),
			)

			err := NewClusterSetup().
				Install(Kuma(core.Standalone,
					WithEnv("KUMA_API_SERVER_AUTH_ALLOW_FROM_LOCALHOST", "true"),
					WithInstallationMode(HelmInstallationMode),
					WithHelmChartPath(Config.HelmChartName),
					WithHelmReleaseName(releaseName),
					WithHelmChartVersion(given.initialChartVersion),
					WithoutHelmOpt("global.image.tag"),
					WithHelmOpt("global.image.registry", Config.KumaImageRegistry),
				)).
				Setup(cluster)
			Expect(err).ToNot(HaveOccurred())

			k8sCluster := cluster.(*K8sCluster)

			err = k8sCluster.UpgradeKuma(core.Standalone, WithHelmReleaseName(releaseName))
			Expect(err).ToNot(HaveOccurred())
		},
		func() []TableEntry {
			var out []TableEntry
			for _, version := range Config.SuiteConfig.Helm.Versions {
				out = append(out, Entry("should successfully upgrade from chart v"+version, testCase{
					initialChartVersion: version,
				}))
			}
			return out
		}(),
	)
}
