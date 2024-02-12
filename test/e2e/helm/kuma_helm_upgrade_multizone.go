package helm

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/random"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/util/versions"
	. "github.com/kumahq/kuma/test/framework"
)

func UpgradingWithHelmChartMultizone() {
	var global, zone Cluster
	var globalCP ControlPlane

	releaseName := fmt.Sprintf("kuma-%s", strings.ToLower(random.UniqueId()))
	var oldestSupportedVersion string

	BeforeAll(func() {
		vers, err := versions.Supported(Config.VersionsYamlPath)
		Expect(err).ToNot(HaveOccurred())
		oldestSupportedVersion = versions.OldestUpgradableToLatest(vers)
	})

	BeforeAll(func() {
		global = NewK8sCluster(NewTestingT(), Kuma1, Silent).
			WithTimeout(6 * time.Second).
			WithRetries(60)
		zone = NewK8sCluster(NewTestingT(), Kuma2, Silent).
			WithTimeout(6 * time.Second).
			WithRetries(60)
	})

	E2EAfterAll(func() {
		Expect(zone.DeleteKuma()).To(Succeed())
		Expect(zone.DismissCluster()).To(Succeed())
		Expect(global.DeleteKuma()).To(Succeed())
		Expect(global.DismissCluster()).To(Succeed())
	})

	It("should install a Kuma version 2 minor releases behind the current version on Global", func() {
		err := NewClusterSetup().
			Install(Kuma(core.Global,
				WithInstallationMode(HelmInstallationMode),
				WithHelmChartPath(Config.HelmChartName),
				WithHelmReleaseName(releaseName),
				WithHelmChartVersion(oldestSupportedVersion),
				WithoutHelmOpt("global.image.tag"),
			)).
			Setup(global)
		Expect(err).ToNot(HaveOccurred())

		globalCP = global.GetKuma()
		Expect(globalCP).ToNot(BeNil())
	})

	It("should install a Kuma version 2 minor releases behind the current version on Zone", func() {
		err := NewClusterSetup().
			Install(Kuma(core.Zone,
				WithInstallationMode(HelmInstallationMode),
				WithHelmChartPath(Config.HelmChartName),
				WithHelmReleaseName(releaseName),
				WithHelmChartVersion(oldestSupportedVersion),
				WithGlobalAddress(globalCP.GetKDSServerAddress()),
				WithHelmOpt("ingress.enabled", "true"),
				WithoutHelmOpt("global.image.tag"),
			)).
			Setup(zone)
		Expect(err).ToNot(HaveOccurred())
	})

	numberOfPolicies := func(c Cluster) (int, error) {
		output, err := c.GetKumactlOptions().RunKumactlAndGetOutput("get", "meshtimeouts", "-m", "default", "-o", "json")
		if err != nil {
			return 0, err
		}
		t := struct {
			Total int `json:"total"`
		}{}
		if err := json.Unmarshal([]byte(output), &t); err != nil {
			return 0, err
		}
		return t.Total, nil
	}

	It("should sync policies from Global to Zone", func() {
		// when apply policy on Global
		Expect(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: mt1
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Mesh
  from:
    - targetRef:
        kind: Mesh
      default:
        idleTimeout: 20s
        http:
          requestTimeout: 2s
          maxStreamDuration: 20s`, Config.KumaNamespace, "default"))(global)).To(Succeed())

		// then the policy is synced to Zone
		Eventually(func(g Gomega) int {
			n, err := numberOfPolicies(zone)
			g.Expect(err).ToNot(HaveOccurred())
			return n
		}, "30s", "1s").Should(Equal(1))
	})

	It("should upgrade Kuma on Global", func() {
		err := global.(*K8sCluster).UpgradeKuma(core.Global,
			WithHelmReleaseName(releaseName),
			WithHelmChartPath(Config.HelmChartPath),
			ClearNoHelmOpts(),
		)
		Expect(err).ToNot(HaveOccurred())
	})

	// Disabled due to https://github.com/kumahq/kuma/issues/9184
	PIt("should upgrade Kuma on Zone", func() {
		// when
		err := zone.(*K8sCluster).UpgradeKuma(core.Zone,
			WithHelmReleaseName(releaseName),
			WithHelmChartPath(Config.HelmChartPath),
			ClearNoHelmOpts(),
		)
		Expect(err).ToNot(HaveOccurred())

		// then
		Consistently(func(g Gomega) int {
			n, err := numberOfPolicies(global)
			g.Expect(err).ToNot(HaveOccurred())
			return n
		}, "10s", "100ms").Should(Equal(1))
		// and
		Consistently(func(g Gomega) int {
			n, err := numberOfPolicies(zone)
			g.Expect(err).ToNot(HaveOccurred())
			return n
		}, "10s", "100ms").Should(Equal(1))
	})
}
