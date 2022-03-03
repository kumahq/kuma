package gatewayapi

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestGatewayAPI(t *testing.T) {
	test.RunSpecs(t, "Gateway API support")
}
