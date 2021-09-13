package matching_test

import (
	"testing"

	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/matching"
	"github.com/kumahq/kuma/test/framework"
)

var _ = Describe("Test Matching on Universal", matching.Universal)

func TestE2EMatching(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		test.RunSpecs(t, "E2E Retry Suite")
	} else {
		t.SkipNow()
	}
}
