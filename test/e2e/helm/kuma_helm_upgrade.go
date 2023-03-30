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

	E2EAfterEach(func() {
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

			releaseName := fmt.Sprintf(
				"kuma-%s",
				strings.ToLower(random.UniqueId()),
			)

			err := NewClusterSetup().
				Install(Kuma(core.Standalone,
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

			err = k8sCluster.UpgradeKuma(core.Standalone, WithHelmReleaseName(releaseName), WithHelmChartPath(Config.HelmChartPath))
			Expect(err).ToNot(HaveOccurred())

			// when
			out, err := k8s.RunKubectlAndGetOutputE(
				k8sCluster.GetTesting(),
				k8sCluster.GetKubectlOptions(),
				"get", "crd", "meshtrafficpermissions.kuma.io", "-oyaml",
			)

			// then CRD is upgraded
			Expect(out).To(ContainSubstring("AllowWithShadowDeny"))
			// remove this when+then after initialChartVersion is changed to 2.1.x or later
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
