package universal_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/membership/universal"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Membership Universal Suite")
}

var _ = Describe("Test Universal", universal.MembershipUniversal)
