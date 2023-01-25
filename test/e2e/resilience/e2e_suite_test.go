package resilience_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/resilience"
)

func TestE2E(t *testing.T) {
	test.RunE2ESpecs(t, "E2E Resilience Suite")
}

var _ = Describe("Test Multizone Resilience for K8s", Label("job-1"), resilience.ResilienceMultizoneK8s, Ordered)
