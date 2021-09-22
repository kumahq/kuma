package gateway_test

import (
	"testing"

	"github.com/kumahq/kuma/test/e2e/gateway"

	"github.com/kumahq/kuma/pkg/test"
	. "github.com/onsi/ginkgo"
)

var _ = Describe("Test Gateway on Universal", gateway.GatewayOnUniversal)

func TestE2EGateway(t *testing.T) {
	test.RunSpecs(t, "E2E Gateway Suite")
}
