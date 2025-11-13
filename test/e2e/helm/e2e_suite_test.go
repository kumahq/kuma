package helm_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/v2/pkg/test"
	"github.com/kumahq/kuma/v2/test/e2e/helm"
	"github.com/kumahq/kuma/v2/test/framework/report"
)

func TestE2E(t *testing.T) {
	test.RunE2ESpecs(t, "E2E Helm Suite")
}

var (
	_ = ReportAfterSuite("report suite", report.DumpReport)
	_ = Describe("Zone and Global with Helm chart", Label("job-2"), helm.ZoneAndGlobalWithHelmChart, Ordered)
	_ = Describe("Zone and Global universal mode with Helm chart", Label("job-2"), helm.ZoneAndGlobalInUniversalModeWithHelmChart, Ordered)
	_ = Describe("Global and Zone universal mode with Helm chart", Label("job-0"), helm.GlobalAndZoneInUniversalModeWithHelmChart, Ordered)
	_ = Describe("Upgrade Standalone with Helm", Label("job-0"), helm.UpgradingWithHelmChartStandalone, Ordered)
	_ = Describe("Upgrade Multizone with Helm", Label("job-2"), helm.UpgradingWithHelmChartMultizone, Ordered)
)
