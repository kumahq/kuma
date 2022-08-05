package gateway_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/gateway"
)

var _ = Describe("Test Gateway on Universal", Label("arm-not-supported"), gateway.GatewayOnUniversal)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Gateway Suite")
}
