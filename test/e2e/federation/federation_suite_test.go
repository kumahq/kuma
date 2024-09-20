package federation_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/federation"
)

func TestE2E(t *testing.T) {
	test.RunE2ESpecs(t, "Federation Suite")
}

var (
	_ = Describe("Federation with Kube Global", Label("job-3"), federation.FederateKubeZoneCPToKubeGlobal, Ordered)
	// TODO: fix the flaky test in the future https://github.com/kumahq/kuma/issues/11420
	_ = Describe("Federation with Universal Global", Label("job-3"), federation.FederateKubeZoneCPToUniversalGlobal, Ordered, FlakeAttempts(3))
)
