package resilience_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/framework"
)

func TestE2EResilience(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		test.RunSpecs(t, "Resilience tests")
	} else {
		t.SkipNow()
	}
}
