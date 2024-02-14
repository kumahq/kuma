package helm_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/helm"
)

func TestE2E(t *testing.T) {
	test.RunE2ESpecs(t, "E2E Helm Suite")
}

var (
	_ = Describe("Test Zone and Global with Helm chart", Label("job-2"), helm.ZoneAndGlobalWithHelmChart, Ordered)
	_ = Describe("Test Zone and Global universal mode with Helm chart", Label("job-0"), helm.ZoneAndGlobalInUniversalModeWithHelmChart, Ordered)
	_ = Describe("Test Global and Zone universal mode with Helm chart", Label("job-0"), helm.GlobalAndZoneInUniversalModeWithHelmChart, Ordered)
	_ = Describe("Test App deployment with Helm chart", Label("job-2"), helm.AppDeploymentWithHelmChart)
	_ = Describe("Test upgrading Standalone with Helm chart", Label("job-2"), helm.UpgradingWithHelmChartStandalone, Ordered)
	_ = Describe("Test upgrading Multizone with Helm chart", Label("job-2"), helm.UpgradingWithHelmChartMultizone, Ordered)
)
