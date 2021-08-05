package helm

import (
	"fmt"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/random"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

var OldChart = "0.4.5"
var UpstreamImageRegistry = "kumahq"

var InitCluster = func(cluster Cluster) {}

func UpgradingWithHelmChart() {
	var cluster Cluster
	var deployOptsFuncs = KumaK8sDeployOpts

	AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}

		Expect(cluster.DeleteKuma(deployOptsFuncs...)).To(Succeed())
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	type testCase struct {
		initialChartVersion string
	}

	DescribeTable(
		"should successfully upgrade Kuma via Helm",
		func(given testCase) {
			c, err := NewK8sClusterWithTimeout(
				NewTestingT(),
				Kuma1,
				Silent,
				6*time.Second)
			Expect(err).ToNot(HaveOccurred())

			cluster = c.WithRetries(60)
			InitCluster(cluster)

			releaseName := fmt.Sprintf(
				"kuma-%s",
				strings.ToLower(random.UniqueId()),
			)

			deployOptsFuncs = append(deployOptsFuncs,
				WithInstallationMode(HelmInstallationMode),
				WithHelmChartPath(HelmRepo),
				WithHelmReleaseName(releaseName),
				WithHelmChartVersion(given.initialChartVersion),
				WithoutHelmOpt("global.image.tag"),
				WithHelmOpt("global.image.registry", UpstreamImageRegistry))

			err = NewClusterSetup().
				Install(Kuma(core.Standalone, deployOptsFuncs...)).
				Setup(cluster)
			Expect(err).ToNot(HaveOccurred())

			Expect(cluster.VerifyKuma()).To(Succeed())

			k8sCluster := cluster.(*K8sCluster)

			upgradeOptsFuncs := append(KumaK8sDeployOpts,
				WithHelmReleaseName(releaseName))

			err = k8sCluster.UpgradeKuma(core.Standalone, upgradeOptsFuncs...)
			Expect(err).ToNot(HaveOccurred())
		},
		Entry("should successfully upgrade from chart v"+OldChart, testCase{
			initialChartVersion: OldChart,
		}),
	)
}
