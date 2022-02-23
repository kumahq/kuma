package timeout_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/timeout"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Timeout Suite")
}

var _ = Describe("Test Timeout policy on Universal", timeout.TimeoutPolicyOnUniversal)
