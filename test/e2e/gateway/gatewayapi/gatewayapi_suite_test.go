package gatewayapi_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/gateway/gatewayapi"
)

var _ = Describe("Test Gateway API", gatewayapi.GatewayAPI)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Gateway API Suite")
}
