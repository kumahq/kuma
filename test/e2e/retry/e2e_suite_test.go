package retry_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/retry"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Retry Suite")
}

var _ = Describe("Test Retry on Universal", retry.RetryOnUniversal)
