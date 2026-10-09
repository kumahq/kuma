package zoneingresstagless_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/v2/pkg/test"
	"github.com/kumahq/kuma/v2/test/e2e/zoneingresstagless"
	"github.com/kumahq/kuma/v2/test/framework/report"
)

func TestE2E(t *testing.T) {
	test.RunE2ESpecs(t, "E2E ZoneIngress With Inbound Tags Disabled Suite")
}

var (
	_ = ReportAfterSuite("report suite", report.DumpReport)
	_ = Describe("ZoneIngress with inbound tags disabled", Label("job-1"), Ordered, zoneingresstagless.ZoneIngressWithInboundTagsDisabled, Serial)
)
