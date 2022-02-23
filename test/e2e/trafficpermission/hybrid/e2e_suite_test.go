package hybrid_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/trafficpermission/hybrid"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Traffic Permission Hybrid Suite")
}

var _ = Describe("Traffic Permission on Hybrid", hybrid.TrafficPermissionHybrid)
