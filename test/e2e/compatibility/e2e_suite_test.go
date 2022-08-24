package compatibility_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/compatibility"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Compatibility Suite")
}

var _ = Describe("Test Kubernetes Multizone Compatibility", Label("job-1"), Label("arm-not-supported"), compatibility.CpCompatibilityMultizoneKubernetes)

// Set FlakeAttempts because sometimes there is a problem with fetching Kuma binaries from pulp.
var _ = Describe("Test Universal Compatibility", Label("job-1"), Label("arm-not-supported"), FlakeAttempts(3), compatibility.UniversalCompatibility)
