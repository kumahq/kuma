package compatibility_test

import (
	"testing"

	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/compatibility"
	"github.com/kumahq/kuma/test/framework"
)

func TestE2ECompatibility(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		test.RunSpecs(t, "E2E Compatibility Suite")
	} else {
		t.SkipNow()
	}
}

var _ = Describe("Test Kubernetes Multizone Compatibility", compatibility.CpCompatibilityMultizoneKubernetes)
var _ = Describe("Test Universal Compatibility", compatibility.UniversalCompatibility)
