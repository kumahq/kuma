package dp_test

import (
	"testing"

	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/auth/dp"
	"github.com/kumahq/kuma/test/framework"
)

func TestE2EDpAuth(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		test.RunSpecs(t, "E2E Auth DP Suite")
	} else {
		t.SkipNow()
	}
}

var _ = Describe("Test Universal", dp.DpAuthUniversal)
