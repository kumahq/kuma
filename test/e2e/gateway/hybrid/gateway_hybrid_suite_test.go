package hybrid_test

import (
	"testing"

	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/gateway/hybrid"
	"github.com/kumahq/kuma/test/framework"
)

var _ = Describe("Test Gateway on Hybrid", hybrid.GatewayHybrid)

func TestE2EGateway(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		test.RunSpecs(t, "E2E Gateway Suite")
	} else {
		t.SkipNow()
	}
}
