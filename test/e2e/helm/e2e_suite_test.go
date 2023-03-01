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
	_ = Describe("Test Zone and Global with Helm chart", Label("job-2"), Label("arm-not-supported"), helm.ZoneAndGlobalWithHelmChart, Ordered)
	_ = Describe("Test Zone and Global universal mode with Helm chart", Label("job-0"), Label("arm-not-supported"), helm.ZoneAndGlobalInUniversalModeWithHelmChart, Ordered)
	_ = Describe("Test App deployment with Helm chart", Label("job-2"), Label("arm-not-supported"), helm.AppDeploymentWithHelmChart)
	_ = Describe("Test upgrading with Helm chart", Label("job-2"), Label("arm-not-supported"), helm.UpgradingWithHelmChart)
)
