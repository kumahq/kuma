package retry_test

import (
	"testing"

	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/retry"
	"github.com/kumahq/kuma/test/framework"
)

func TestE2ERetry(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		test.RunSpecs(t, "E2E Retry Suite")
	} else {
		t.SkipNow()
	}
}

var _ = Describe("Test Retry on Universal", retry.RetryOnUniversal)
