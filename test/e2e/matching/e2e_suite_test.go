package matching_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/matching"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Matching Suite")
}

var _ = Describe("Test Matching on Universal", matching.Universal)
