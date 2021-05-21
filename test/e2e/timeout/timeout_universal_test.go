package timeout_test

import (
	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/test/e2e/timeout"
)

var _ = Describe("Test Timeout policy on Universal", timeout.TimeoutPolicyOnUniversal)
