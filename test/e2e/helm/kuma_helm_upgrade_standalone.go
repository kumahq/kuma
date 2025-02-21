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

func UpgradingWithHelmChartStandalone() {
	var cluster Cluster

	E2EAfterEach(func() {
		Expect(cluster.DeleteKuma()).To(Succeed())
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	DescribeTable("upgrade Kuma via Helm",
		func(version string) {
			cluster = NewK8sCluster(NewTestingT(), Kuma1, Silent).
				WithTimeout(6 * time.Second).
				WithRetries(60)

			releaseName := fmt.Sprintf(
				"kuma-%s",
				strings.ToLower(random.UniqueId()),
			)

			// nolint:staticcheck
			err := NewClusterSetup().
				Install(Kuma(core.Standalone,
					WithInstallationMode(HelmInstallationMode),
					WithHelmChartPath(Config.HelmChartName),
					WithHelmReleaseName(releaseName),
					WithHelmChartVersion(version),
					WithoutHelmOpt("global.image.tag"),
				)).
				Setup(cluster)
			Expect(err).ToNot(HaveOccurred())

			k8sCluster := cluster.(*K8sCluster)

			err = k8sCluster.UpgradeKuma(core.Zone,
				WithHelmReleaseName(releaseName),
				WithHelmChartPath(Config.HelmChartPath),
				ClearNoHelmOpts(),
			)
			Expect(err).ToNot(HaveOccurred())

			// when
			out, err := k8s.RunKubectlAndGetOutputE(
				k8sCluster.GetTesting(),
				k8sCluster.GetKubectlOptions(),
				"get", "crd", "meshtrafficpermissions.kuma.io", "-oyaml",
			)

			// then CRD is upgraded
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(ContainSubstring("AllowWithShadowDeny"))
		},
		EntryDescription("from version: %s"),
		SupportedVersionEntries(NewTestingT()),
	)
}
