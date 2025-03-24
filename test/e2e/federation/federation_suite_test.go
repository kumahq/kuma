package federation_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/federation"
	"github.com/kumahq/kuma/test/framework/report"
)

func TestE2E(t *testing.T) {
	test.RunE2ESpecs(t, "Federation Suite")
}

var (
	_ = ReportAfterSuite("report suite", report.DumpReport)
	_ = Describe("Federation with Kube Global", Label("job-3"), federation.FederateKubeZoneCPToKubeGlobal, Ordered)
	_ = Describe("Federation with Universal Global", Label("job-3"), federation.FederateKubeZoneCPToUniversalGlobal, Ordered)
)
