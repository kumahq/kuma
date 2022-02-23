package universal_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/trafficpermission/universal"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Traffic Permission Universal Suite")
}

var _ = Describe("Traffic Permission on Universal", universal.TrafficPermissionUniversal)
