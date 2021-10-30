package gateway_test

import (
	"testing"

	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/gateway"
	"github.com/kumahq/kuma/test/framework"
)

var _ = Describe("Test Gateway on Universal", gateway.GatewayOnUniversal)

func TestE2EGateway(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		test.RunSpecs(t, "E2E Gateway Suite")
	} else {
		t.SkipNow()
	}
}
