package transparentproxy_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/transparentproxy"
	"github.com/kumahq/kuma/test/framework/report"
)

func TestE2E(t *testing.T) {
	test.RunE2ESpecs(t, "E2E Transparent Proxy Suite")
}

var (
	_ = ReportAfterSuite("report suite", report.DumpReport)
	_ = Describe("Test Transparent Proxy Configuration in ConfigMap", Label("job-0"), transparentproxy.TransparentProxyConfigMap, Ordered)
)
