package federation_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/v2/pkg/test"
	"github.com/kumahq/kuma/v2/test/e2e/federation"
	"github.com/kumahq/kuma/v2/test/framework/report"
)

func TestE2E(t *testing.T) {
	test.RunE2ESpecs(t, "Federation Suite")
}

var (
	_ = ReportAfterSuite("report suite", report.DumpReport)
	_ = Describe("Federation with Kube Global", Label("job-3"), Ordered, federation.FederateKubeZoneCPToKubeGlobal, Serial)
	_ = Describe("Federation with Universal Global", Label("job-3"), Ordered, federation.FederateKubeZoneCPToUniversalGlobal, Serial)
)
