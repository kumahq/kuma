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

			opts := []KumaDeploymentOption{
				WithInstallationMode(HelmInstallationMode),
				WithHelmChartPath(Config.HelmChartName),
				WithHelmReleaseName(releaseName),
				WithHelmChartVersion(version),
				WithoutHelmOpt("global.image.tag"),
			}

			// Bitnami registry changes:
			// We switched the kubectl image source from Bitnami to the official registry.k8s.io
			// on supported release branches. These changes were not backported to 2.8.x, which is
			// out of support and still uses the Bitnami image. After Bitnami's registry changes,
			// installation of 2.8 fails in our upgrade tests because the Bitnami kubectl repository
			// does not provide versioned tags.
			// Our upgrade tests purposely install older versions (like 2.8.4) and then upgrade them,
			// so fixing this properly would require releasing a new, long out-of-support 2.8 just to
			// change the registry. Instead, for 2.8.x only, we set kubectl.image.tag to "latest"
			// (a tag that still exists), which should keep these tests passing until upgrades only
			// involve supported versions that already use the new registry.
			if strings.HasPrefix(version, "2.8.") {
				opts = append(opts, WithHelmOpt("kubectl.image.tag", "latest"))
			}

			// nolint:staticcheck
			err := NewClusterSetup().
				Install(Kuma(core.Standalone, opts...)).
				Setup(cluster)
			Expect(err).ToNot(HaveOccurred())

			k8sCluster := cluster.(*K8sCluster)

			err = k8sCluster.UpgradeKuma(core.Zone,
				WithHelmReleaseName(releaseName),
				WithHelmChartPath(Config.HelmChartPath),
				WithoutHelmOpt("kubectl.image.tag"),
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
		SupportedVersionEntries(),
	)
}
