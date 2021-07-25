package universal_multizone_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/framework"
)

func TestE2ETrafficRouteUniversalMultizone(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		test.RunSpecs(t, "Traffic Route Universal Multizone Suite")
	} else {
		t.SkipNow()
	}
}
