package ownership_test

import (
	"testing"

	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/ownership"
	"github.com/kumahq/kuma/test/framework"
)

func TestE2EOwnership(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		test.RunSpecs(t, "Ownership tests")
	} else {
		t.SkipNow()
	}
}

var _ = Describe("Test Multizone Ownership for Universal", ownership.MultizoneUniversal)
