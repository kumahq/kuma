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

var _ = Describe("Test Kubernetes Multizone Compatibility", compatibility.CpCompatibilityMultizoneKubernetes)
var _ = Describe("Test Universal Compatibility", compatibility.UniversalCompatibility)
