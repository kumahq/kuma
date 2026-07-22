package helm_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/v3/pkg/test"
	"github.com/kumahq/kuma/v3/test/e2e/helm"
	"github.com/kumahq/kuma/v3/test/framework/report"
)

func TestE2E(t *testing.T) {
	test.RunE2ESpecs(t, "E2E Helm Suite")
}

var (
	_ = ReportAfterSuite("report suite", report.DumpReport)
	_ = Describe("Zone with Helm chart and Universal Global", Label("job-0"), Ordered, helm.ZoneWithHelmChartAndUniversalGlobal, Serial)
	_ = Describe("Upgrade Zone with Helm chart", Label("job-2"), helm.UpgradingZoneWithHelmChart, Serial)
)
