package universal_test

import (
	"testing"

	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/membership/universal"
	"github.com/kumahq/kuma/test/framework"
)

func TestE2EMembershipUniversal(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		test.RunSpecs(t, "E2E Membership Universal Suite")
	} else {
		t.SkipNow()
	}
}

var _ = Describe("Test Universal", universal.MembershipUniversal)
