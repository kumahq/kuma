package universal_test

import (
	"testing"

	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/trafficpermission/universal"
	"github.com/kumahq/kuma/test/framework"
)

func TestE2ETrafficPermissionUniversal(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		test.RunSpecs(t, "Traffic Permission Universal Suite")
	} else {
		t.SkipNow()
	}
}

var _ = Describe("Traffic Permission on Universal", universal.TrafficPermissionUniversal)
