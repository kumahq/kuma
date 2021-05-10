package retry_test

import (
	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/test/e2e/retry"
)

var _ = Describe("Test Retry on Universal", retry.RetryOnUniversal)
