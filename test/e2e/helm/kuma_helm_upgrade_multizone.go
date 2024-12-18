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
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/api"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/versions"
)

func UpgradingWithHelmChartMultizone() {
	namespace := "helm-upgrade-ns"
	var global, zoneK8s, zoneUniversal Cluster
	var globalCP ControlPlane

	releaseName := fmt.Sprintf("kuma-%s", strings.ToLower(random.UniqueId()))
	var oldestSupportedVersion string

	BeforeAll(func() {
		oldestSupportedVersion = versions.OldestUpgradableToBuildVersion(Config.SupportedVersions())
	})

	BeforeAll(func() {
		global = NewK8sCluster(NewTestingT(), Kuma1, Silent).
			WithTimeout(6 * time.Second).
			WithRetries(60)
		zoneK8s = NewK8sCluster(NewTestingT(), Kuma2, Silent).
			WithTimeout(6 * time.Second).
			WithRetries(60)
		zoneUniversal = NewUniversalCluster(NewTestingT(), Kuma3, Silent)
	})

	E2EAfterAll(func() {
		Expect(zoneUniversal.DismissCluster()).To(Succeed())
		Expect(zoneK8s.DeleteNamespace(namespace)).To(Succeed())
		Expect(zoneK8s.DeleteKuma()).To(Succeed())
		Expect(zoneK8s.DismissCluster()).To(Succeed())
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
			Setup(zoneK8s)
		Expect(err).ToNot(HaveOccurred())
	})

	numberOfResources := func(c Cluster, resource string) (int, error) {
		output, err := c.GetKumactlOptions().RunKumactlAndGetOutput("get", resource, "-o", "json")
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

	numberOfPolicies := func(c Cluster) (int, error) {
		return numberOfResources(c, "meshtimeouts")
	}

	numberOfDPPs := func(c Cluster) (int, error) {
		return numberOfResources(c, "dataplanes")
	}

	numberOfZoneIngresses := func(c Cluster) (int, error) {
		return numberOfResources(c, "zone-ingresses")
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
			n, err := numberOfPolicies(zoneK8s)
			g.Expect(err).ToNot(HaveOccurred())
			return n
		}, "30s", "1s").Should(Equal(1))
	})

	It("should sync DPPs from Zone to Global", func() {
		// when start test server on Zone
		err := NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(testserver.Install(testserver.WithNamespace(namespace))).Setup(zoneK8s)
		Expect(err).ToNot(HaveOccurred())

		// then the DPP is synced to Global
		Eventually(func(g Gomega) int {
			n, err := numberOfDPPs(global)
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

	It("should deploy a universal zone with the latest Kuma", func() {
		err := NewClusterSetup().
			Install(Kuma(core.Zone, WithGlobalAddress(global.GetKuma().GetKDSServerAddress()))).
			Install(IngressUniversal(global.GetKuma().GenerateZoneIngressToken)).
			Setup(zoneUniversal)
		Expect(err).ToNot(HaveOccurred())

		Eventually(func(g Gomega) int {
			n, err := numberOfZoneIngresses(zoneK8s)
			g.Expect(err).ToNot(HaveOccurred())
			return n
		}, "30s", "1s").Should(Equal(2))
	})

	It("should upgrade Kuma on Zone", func() {
		// when
		err := zoneK8s.(*K8sCluster).UpgradeKuma(core.Zone,
			WithHelmReleaseName(releaseName),
			WithHelmChartPath(Config.HelmChartPath),
			ClearNoHelmOpts(),
		)
		Expect(err).ToNot(HaveOccurred())

		Eventually(func(g Gomega) {
			result := &system.ZoneInsightResource{}
			api.FetchResource(g, global, result, "", "kuma-2")
			g.Expect(len(result.Spec.Subscriptions)).To(BeNumerically(">", 1))
			newZoneConnected := false
			for _, sub := range result.Spec.Subscriptions {
				if sub.Version.KumaCp.Version != oldestSupportedVersion {
					newZoneConnected = true
					break
				}
			}
			g.Expect(newZoneConnected).To(BeTrue())
		}, "30s", "100ms").Should(Succeed())

		// then
		Consistently(func(g Gomega) {
			policiesGlobal, err := numberOfPolicies(global)
			g.Expect(err).ToNot(HaveOccurred())

			policiesZone, err := numberOfPolicies(zoneK8s)
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(policiesGlobal).To(And(Equal(policiesZone), Equal(1)))

			dppsGlobal, err := numberOfDPPs(global)
			g.Expect(err).ToNot(HaveOccurred())

			dppsZone, err := numberOfDPPs(zoneK8s)
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(dppsGlobal).To(And(Equal(dppsZone), Equal(1)))

			zoneIngressesGlobal, err := numberOfZoneIngresses(global)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(zoneIngressesGlobal).To(Equal(2))
		}, "5s", "100ms").Should(Succeed())
	})
}
