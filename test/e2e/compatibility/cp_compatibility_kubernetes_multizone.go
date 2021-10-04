package compatibility

import (
	"fmt"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

var OldChart = "0.6.3"
var UpstreamImageRegistry = "kumahq"

func CpCompatibilityMultizoneKubernetes() {
	var globalCluster Cluster
	var globalReleaseName string
	var globalDeployOptsFuncs = KumaK8sDeployOpts

	var zoneCluster Cluster
	var zoneDeployOptsFuncs = KumaZoneK8sDeployOpts
	var zoneReleaseName string

	// Ensure that the upstream Kuma help repository is configured
	// and refreshed. This is needed for helm to be able to pull the
	// OldChart version of the Kuma helm chart.
	BeforeSuite(func() {
		t := NewTestingT()
		opts := helm.Options{}

		// Adding the same repo multiple times is idempotent. The
		// `--force-update` flag prevents heml emitting an error
		// in this case.
		_, err := helm.RunHelmCommandAndGetOutputE(t, &opts,
			"repo", "add", "--force-update", "kuma", "https://kumahq.github.io/charts")
		Expect(err).To(Succeed())

		_, err = helm.RunHelmCommandAndGetOutputE(t, &opts, "repo", "update")
		Expect(err).To(Succeed())
	})

	BeforeEach(func() {
		// Global CP
		c, err := NewK8sClusterWithTimeout(
			NewTestingT(),
			Kuma1,
			Silent,
			6*time.Second)
		Expect(err).ToNot(HaveOccurred())
		globalCluster = c.WithRetries(60)

		globalReleaseName = fmt.Sprintf(
			"kuma-%s",
			strings.ToLower(random.UniqueId()),
		)

		globalDeployOptsFuncs = append(globalDeployOptsFuncs,
			WithInstallationMode(HelmInstallationMode),
			WithHelmChartPath(HelmRepo),
			WithHelmReleaseName(globalReleaseName),
			WithHelmChartVersion(OldChart),
			WithoutHelmOpt("global.image.tag"),
			WithHelmOpt("global.image.registry", UpstreamImageRegistry))

		err = NewClusterSetup().
			Install(Kuma(core.Global, globalDeployOptsFuncs...)).
			Setup(globalCluster)
		Expect(err).ToNot(HaveOccurred())

		Expect(globalCluster.VerifyKuma()).To(Succeed())

		// Zone CP
		c, err = NewK8sClusterWithTimeout(
			NewTestingT(),
			Kuma2,
			Silent,
			6*time.Second)
		Expect(err).ToNot(HaveOccurred())
		zoneCluster = c.WithRetries(60)

		zoneReleaseName = fmt.Sprintf(
			"kuma-%s",
			strings.ToLower(random.UniqueId()),
		)

		zoneDeployOptsFuncs = append(zoneDeployOptsFuncs,
			WithInstallationMode(HelmInstallationMode),
			WithHelmChartPath(HelmRepo),
			WithHelmReleaseName(zoneReleaseName),
			WithHelmChartVersion(OldChart),
			WithoutHelmOpt("global.image.tag"),
			WithHelmOpt("global.image.registry", UpstreamImageRegistry),
			WithGlobalAddress(globalCluster.GetKuma().GetKDSServerAddress()),
			WithHelmOpt("ingress.enabled", "true"),
		)

		err = NewClusterSetup().
			Install(Kuma(core.Zone, zoneDeployOptsFuncs...)).
			Install(NamespaceWithSidecarInjection(TestNamespace)).
			Setup(zoneCluster)
		Expect(err).ToNot(HaveOccurred())

		Expect(zoneCluster.VerifyKuma()).To(Succeed())
	})

	AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}

		Expect(zoneCluster.DeleteKuma(zoneDeployOptsFuncs...)).To(Succeed())
		Expect(zoneCluster.DismissCluster()).To(Succeed())

		Expect(globalCluster.DeleteKuma(globalDeployOptsFuncs...)).To(Succeed())
		Expect(globalCluster.DismissCluster()).To(Succeed())
	})

	It("should sync resources between new global and old zone", func() {
		// when global is upgraded
		upgradeOptsFuncs := append(KumaK8sDeployOpts, WithHelmReleaseName(globalReleaseName))
		err := globalCluster.(*K8sCluster).UpgradeKuma(core.Global, upgradeOptsFuncs...)
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
	})
}
