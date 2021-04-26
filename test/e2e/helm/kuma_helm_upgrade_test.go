package helm_test

import (
	"fmt"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/random"

	"github.com/kumahq/kuma/pkg/config/core"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
)

var _ = PDescribe("Test upgrading with Helm chart", func() {
	var cluster Cluster
	var deployOptsFuncs []DeployOptionsFunc

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

			releaseName := fmt.Sprintf(
				"kuma-%s",
				strings.ToLower(random.UniqueId()),
			)

			deployOptsFuncs = []DeployOptionsFunc{
				WithInstallationMode(HelmInstallationMode),
				WithHelmChartPath("kuma/kuma"),
				WithHelmReleaseName(releaseName),
				WithHelmChartVersion(given.initialChartVersion),
				WithoutHelmOpt("global.image.tag"),
				WithoutHelmOpt("global.image.registry"),
			}

			err = NewClusterSetup().
				Install(Kuma(core.Standalone, deployOptsFuncs...)).
				Install(KumaDNS()).
				Setup(cluster)
			Expect(err).ToNot(HaveOccurred())

			Expect(cluster.VerifyKuma()).To(Succeed())

			k8sCluster := cluster.(*K8sCluster)

			upgradeOptsFuncs := []DeployOptionsFunc{
				WithHelmReleaseName(releaseName),
			}

			err = k8sCluster.UpgradeKuma(core.Standalone, upgradeOptsFuncs...)
			Expect(err).ToNot(HaveOccurred())
		},
		Entry("should successfully upgrade from chart v0.4.4", testCase{
			initialChartVersion: "0.4.4",
		}),
		Entry("should successfully upgrade from chart v0.4.5", testCase{
			initialChartVersion: "0.4.5",
		}),
	)
})
