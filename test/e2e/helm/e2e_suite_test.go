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
	_ = Describe("Test Zone and Global universal mode with Helm chart", Label("job-2"), helm.ZoneAndGlobalInUniversalModeWithHelmChart, Ordered)
	// Skipped as it fails with: error while running command: exit status 1; Error: INSTALLATION FAILED: rendered manifests contain a resource that already exists. Unable to continue with install: ServiceAccount "kuma-control-plane" in namespace "kuma-system" exists and cannot be imported into the current release: invalid ownership metadata; annotation validation error: key "meta.helm.sh/release-name" must equal "kuma-c0vo8o": current value is "kuma-8yy3uv"
	// Likely something needs to be improved to be able to run this test
	_ = PDescribe("Test Global and Zone universal mode with Helm chart", Label("job-0"), helm.GlobalAndZoneInUniversalModeWithHelmChart, Ordered)
	_ = Describe("Test upgrading Standalone with Helm chart", Label("job-2"), helm.UpgradingWithHelmChartStandalone, Ordered)
	_ = Describe("Test upgrading Multizone with Helm chart", Label("job-2"), helm.UpgradingWithHelmChartMultizone, Ordered)
)
