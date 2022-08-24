package ownership_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/ownership"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Ownership tests")
}

var _ = Describe("Test Multizone Ownership for Universal", Label("job-2"), ownership.MultizoneUniversal)
