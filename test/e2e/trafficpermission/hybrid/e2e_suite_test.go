package hybrid_test

import (
	"testing"

	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/trafficpermission/hybrid"
	"github.com/kumahq/kuma/test/framework"
)

func TestE2ETrafficPermissionHybrid(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		test.RunSpecs(t, "Traffic Permission Hybrid Suite")
	} else {
		t.SkipNow()
	}
}

var _ = Describe("Traffic Permission on Hybrid", hybrid.TrafficPermissionHybrid)
