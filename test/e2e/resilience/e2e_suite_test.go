package resilience_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/v2/pkg/test"
	"github.com/kumahq/kuma/v2/test/e2e/resilience"
	"github.com/kumahq/kuma/v2/test/framework/report"
)

func TestE2E(t *testing.T) {
	test.RunE2ESpecs(t, "E2E Resilience Suite")
}

var (
	_ = ReportAfterSuite("report suite", report.DumpReport)
	_ = Describe("Test Multizone Resilience for K8s", Label("job-1"), resilience.ResilienceMultizoneK8s, Ordered)
)
