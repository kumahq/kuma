package hybrid_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/gateway/hybrid"
)

var _ = Describe("Test Gateway on Hybrid", Label("arm-not-supported"), hybrid.GatewayHybrid)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Gateway Hybrid Suite")
}
