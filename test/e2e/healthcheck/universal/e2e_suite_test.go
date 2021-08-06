package universal_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/framework"
)

func TestE2EHealthCheckUniversal(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		test.RunSpecs(t, "Health Check Universal Suite")
	} else {
		t.SkipNow()
	}
}
