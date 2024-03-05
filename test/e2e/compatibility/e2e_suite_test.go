package compatibility_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/compatibility"
)

func TestE2E(t *testing.T) {
	test.RunE2ESpecs(t, "E2E Compatibility Suite")
}

var _ = Describe("Test Kubernetes Multizone Compatibility", Label("job-1"), compatibility.CpCompatibilityMultizoneKubernetes)
