package ownership_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/framework"
)

func TestE2EOwnership(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		test.RunSpecs(t, "Ownership tests")
	} else {
		t.SkipNow()
	}
}
