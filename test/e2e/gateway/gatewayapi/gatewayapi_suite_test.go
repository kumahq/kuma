package gatewayapi_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestE2E(t *testing.T) {
	test.RunE2ESpecs(t, "E2E Gateway API Suite")
}
