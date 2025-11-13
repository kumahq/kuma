package cni_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/v2/pkg/test"
	"github.com/kumahq/kuma/v2/test/e2e/cni"
	"github.com/kumahq/kuma/v2/test/framework/report"
)

func TestE2E(t *testing.T) {
	test.RunE2ESpecs(t, "E2E CNI Suite")
}

var (
	_ = ReportAfterSuite("report suite", report.DumpReport)
	_ = Describe("Taint controller", Label("job-1"), Label("kind-not-supported"), Label("legacy-k3s-not-supported"), cni.AppDeploymentWithCniAndTaintController)
	_ = Describe("Connectivity - Exclude Outbound Port", Label("job-0"), Label("kind-not-supported"), Label("legacy-k3s-not-supported"), cni.ExcludeOutboundPort, Ordered)
)
