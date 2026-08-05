package gatewayapi

import (
	"testing"

	"github.com/kumahq/kuma/v3/pkg/test"
)

func TestGatewayAPI(t *testing.T) {
	test.RunSpecs(t, "Gateway API support")
}
