package webhooks_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/v2/pkg/test"
	"github.com/kumahq/kuma/v2/test/e2e/webhooks"
	"github.com/kumahq/kuma/v2/test/framework/report"
)

func TestE2E(t *testing.T) {
	test.RunE2ESpecs(t, "E2E Webhooks Suite")
}

var (
	_ = ReportAfterSuite("report suite", report.DumpReport)
	_ = Describe("Cert-Manager CA Injection", Label("job-0"), webhooks.CertManagerCAInjection, Ordered)
	_ = Describe("Cert-Manager Helm Validation", Label("job-0"), webhooks.CertManagerHelmValidation)
)
