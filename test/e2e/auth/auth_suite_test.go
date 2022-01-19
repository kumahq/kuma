package auth_test

import (
	"testing"

	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/auth"
	"github.com/kumahq/kuma/test/framework"
)

func TestE2EAuth(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		test.RunSpecs(t, "E2E Auth Suite")
	} else {
		t.SkipNow()
	}
}

var _ = Describe("Test Universal", auth.AuthUniversal)
