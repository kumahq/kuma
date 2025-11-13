package gatewayapi

import (
	"testing"

	"github.com/kumahq/kuma/v2/pkg/test"
)

func TestGatewayAPI(t *testing.T) {
	test.RunSpecs(t, "Gateway API support")
}
