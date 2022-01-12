package attachment_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestAllowedRoutes(t *testing.T) {
	test.RunSpecs(t, "Gateway API AllowedRoutes support")
}
